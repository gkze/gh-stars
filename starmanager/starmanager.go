package starmanager

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/gkze/stars/auth"
	"github.com/gkze/stars/utils"
	"github.com/google/go-github/v25/github"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"go.uber.org/multierr"
	"golang.org/x/oauth2"
)

const (
	// GitHubHost - the main public GitHub
	GitHubHost string = "github.com"

	// GitHubAPIHost - the GitHub API host
	GitHubAPIHost string = "api." + GitHubHost

	// CachePath - the path to the cache db file
	CachePath string = ".cache"

	// CacheFile - the filename of the db cache
	CacheFile string = "stars.db"

	// PageSize - the default response page size (GitHub maximum is 100 so we
	// use that)
	PageSize int = 100

	// DefaultConcurrency limits how many goroutines to run during network I/O
	// operations.
	DefaultConcurrency int = 10
)

// Star represents the starred repository that is saved locally
type Star struct {
	// Archives indicates whether the GitHub Repository has been archived
	Archived bool `storm:"index"`

	// Description holds the repository description (can be emppty)
	Description string `storm:"index"`

	// Language is the dominant programming language in the repository.
	// The official list of all recognized programming languages can be found
	// here:
	// https://github.com/github/linguist/blob/master/lib/linguist/languages.yml
	Language string `storm:"index"`

	// PushedAt represents the timestamp of the last (git) push to this
	// repository
	PushedAt time.Time `storm:"index"`

	// Stargazers is the number of users who have starred this repository
	Stargazers int

	// StarredAt represents the timestamp of when the current user starred this
	// repository.
	StarredAt time.Time `storm:"index"`

	// Topics are the tags/labels for the repository.
	Topics []string `storm:"index"`

	// URL is the full web URL of the repository (html_url field in the GitHub
	// API)
	URL string `storm:"id,index,unique"`
}

// StarManager is the central object used to manage stars for a GitHub account
type StarManager struct {
	username string
	password string
	context  context.Context
	client   *github.Client
	db       *storm.DB
}

// New constructs a new StarManager object. This method reads authentication
// configuration from the user's ~/.netrc file for the GitHub API host.
func New(logLevel log.Level) (*StarManager, error) {
	log.Tracef("Setting log level to %+v\n", logLevel)
	log.SetLevel(logLevel)

	log.Debug("Initializing auth")
	cfg, err := auth.NewConfig()
	if err != nil {
		log.Errorf("Could not parse netrc config: %v", err)

		return nil, err
	}

	log.Debug("Parsing auth credentials from .netrc")
	netrcAuth, err := auth.NewNetrc(cfg)
	username, password, err := netrcAuth.GetAuth(GitHubAPIHost)
	if err != nil {
		log.Errorf(
			"Could not find authentication credentials for %s in netrc configuration",
			GitHubAPIHost,
		)

		return nil, err
	}

	log.Trace("Initializing context")
	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(
		ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: password}),
	))

	log.Debug("Determining current user")
	currentUser, err := user.Current()
	if err != nil {
		log.Errorf("Could not determine the current user! %v\n", err)

		return nil, err
	}
	log.Debugf("Current user: %s\n", currentUser.Username)

	log.Debug("Ensuring local cache")
	cacheDir := filepath.Join(currentUser.HomeDir, CachePath)
	cacheFullPath := filepath.Join(cacheDir, CacheFile)

	for _, p := range []struct {
		path string
		mode os.FileMode
	}{
		{cacheDir, os.ModeDir},
		{cacheFullPath, 0},
	} {
		err := utils.CreateIfNotExists(p.path, p.mode, afero.NewOsFs())
		if err != nil {
			log.Infof("An error occurred while attempting to create %s: %v\n",
				p.path, err,
			)
		}
	}

	log.Debug("Initializing Storm/Bolt")
	db, err := storm.Open(cacheFullPath, storm.Batch())
	if err != nil {
		log.Errorf("An error occurred opening the db! %v", err)

		return nil, err
	}

	return &StarManager{
		username: username,
		password: password,
		context:  ctx,
		client:   client,
		db:       db,
	}, nil
}

// ClearCache resets the filesystem-local cache database file.
func (s *StarManager) ClearCache() error {
	log.Debug("Clearing out cache")
	return os.Remove(s.db.Bolt.Path())
}

// StarRepository stars a given repository by owner and repository name
func (s *StarManager) StarRepository(owner, repo string) error {
	log.Debugf("Starring %s/%s\n", owner, repo)

	resp, err := s.client.Activity.Star(s.context, owner, repo)
	if err != nil {
		log.Errorf("An error occurred starring a repository! %+v", resp)
		return err
	}

	if resp.StatusCode == (http.StatusOK | http.StatusNoContent) {
		log.Infof("Successfully starred %s/%s", owner, repo)
	} else {
		log.Warningf("Got non-200 response for %s/%s: %d",
			owner, repo, resp.StatusCode,
		)
	}

	return nil
}

func (s *StarManager) starReposFromURLsChan(
	g *sync.WaitGroup,
	urlsChan chan *url.URL,
	successfulStars chan int,
	notOlderThanMonths int,
) {
	g.Add(1)
	defer g.Done()

	wg := sync.WaitGroup{}
	defer wg.Wait()

	then := time.Now().AddDate(0, -notOlderThanMonths, 0)

	for u := range urlsChan {
		log.Debugf("Got URL: %+v\n", u)

		parts := strings.Split(strings.Trim((*u).EscapedPath(), "/"), "/")
		if len(parts) != 2 {
			log.Errorf("%+v invalid", u)
			continue
		}

		log.Infof("Evaluating %s/%s\n", parts[0], parts[1])
		repo, _, err := s.client.Repositories.Get(s.context, parts[0], parts[1])
		if err != nil {
			log.Errorf("encountered error: %+v", err)
			continue
		}

		owner := repo.GetOwner().GetLogin()
		name := repo.GetName()

		log.Debugf("Checking whether %s/%s is starred\n", owner, name)
		starred, _, err := s.client.Activity.IsStarred(s.context, owner, name)
		if err != nil {
			log.Errorf("Encountered error: %+v", err)
			continue
		}
		if starred {
			log.Infof("%s/%s already starred - skipping\n", owner, name)
			continue
		}

		if !(repo.GetPushedAt().Before(then) || repo.GetArchived()) {
			if err := s.StarRepository(owner, name); err != nil {
				log.Errorf(
					"failed to star %s/%s\n",
					repo.GetOwner().GetLogin(),
					repo.GetName(),
				)
				continue
			}

			wg.Add(1)
			go func() { defer wg.Done(); successfulStars <- 1 }()
			wg.Wait()
		} else {
			log.Infof(
				"%s/%s does not qualify - archived: %t, pushed: %s\n",
				repo.GetOwner().GetLogin(),
				repo.GetName(),
				repo.GetArchived(),
				repo.GetPushedAt(),
			)
		}
	}
}

// StarRepositoriesFromURLs stars each repository in the given slice of
// repository URLs
func (s *StarManager) StarRepositoriesFromURLs(
	urls []*url.URL, notOlderThanMonths, maxConcurrency int,
) (int, error) {
	log.Debugf("Preparing to star %d repositories\n", len(urls))

	urlsCh := make(chan *url.URL)
	starCountCh := make(chan int)
	wg := sync.WaitGroup{}

	log.Debugf("Spawning %d goroutines\n", maxConcurrency)
	for i := 0; i < maxConcurrency; i++ {
		go s.starReposFromURLsChan(&wg, urlsCh, starCountCh, notOlderThanMonths)
	}

	for _, u := range urls {
		urlsCh <- u
	}
	close(urlsCh)

	go func() {
		wg.Wait()
		close(starCountCh)
	}()

	total := 0
	for count := range starCountCh {
		total += count
	}

	if total == 0 {
		log.Warn("Added 0 repos")
		return 0, nil
	}

	log.Infof("Successfully starred %d repos\n", total)
	return total, nil
}

func (s *StarManager) getReposFromOrgPageChan(
	org string, orgReposPageCh chan int, reposCh chan *github.Repository,
) {
	for pageNo := range orgReposPageCh {
		log.Printf("Fetching repos from page %d of %s org\n", pageNo, org)

		repos, resp, err := s.client.Repositories.ListByOrg(
			s.context,
			org,
			&github.RepositoryListByOrgOptions{
				Type:        "sources",
				ListOptions: github.ListOptions{PerPage: PageSize, Page: pageNo},
			},
		)
		if err != nil {
			log.Errorf("encountered error fetching page %d for org %s\n", pageNo, org)
			log.Errorf("response: %+v\n", resp)
			log.Errorf("error: %+v\n", err)

			continue
		}

		log.Infof("Successfully fetched %d repos from page %d of %s org",
			len(repos), pageNo, org,
		)

		for _, repo := range repos {
			reposCh <- repo
		}
	}
}

// StarRepositoriesFromOrg stars a given org's repositories, given that they
// are not archived and are recently pushed to.
func (s *StarManager) StarRepositoriesFromOrg(
	org string, notOlderThanMonths, maxConcurrency int,
) error {
	repoURLs := []*url.URL{}

	log.Infof("Listing first %d repos for %s\n", PageSize, org)
	orgReposPage1, resp, err := s.client.Repositories.ListByOrg(
		s.context,
		org,
		&github.RepositoryListByOrgOptions{
			Type:        "sources",
			ListOptions: github.ListOptions{PerPage: PageSize, Page: 1},
		},
	)
	if err != nil {
		return fmt.Errorf("Got error: %+v (%w)", resp, err)
	}

	log.Infof("Parsing %d repos into URLs\n", len(orgReposPage1))
	for _, p1Repo := range orgReposPage1 {
		repoURL, err := url.Parse(p1Repo.GetHTMLURL())
		if err != nil {
			log.Errorf(
				"encountered error parsing repo url for %s/%s: %+v",
				org, p1Repo.GetName(), err,
			)
			continue
		}

		log.Debugf("Parsed %+v\n", repoURL)

		repoURLs = append(repoURLs, repoURL)
	}

	// Gotta fetch more if that's the case
	if resp.LastPage > 1 {
		log.Infof("Last page is %d. Fetching the rest of the pages\n", resp.LastPage)

		orgPagesCh := make(chan int)
		reposCh := make(chan *github.Repository)
		urlsCh := make(chan *url.URL)
		defer func() { close(orgPagesCh); close(reposCh); close(urlsCh) }()

		startIdx := 2
		numGoroutines := resp.LastPage

		if maxConcurrency > 0 {
			startIdx = 0
			numGoroutines = maxConcurrency
		}

		for i := startIdx; i < numGoroutines; i++ {
			go s.getReposFromOrgPageChan(org, orgPagesCh, reposCh)
		}

		go func(opc chan int, lastPage int) {
			// We start at 2 here since we've already fetched the first page at the
			// beginning of this function. We also end at last page + 1 since we
			// want to include the last page to be fetched.
			for i := 2; i < lastPage+1; i++ {
				opc <- i
			}
		}(orgPagesCh, resp.LastPage)

		go func(rCh chan *github.Repository, uCh chan *url.URL) {
			for repo := range reposCh {
				repoURL, err := url.Parse(repo.GetHTMLURL())
				if err != nil {
					log.Errorf(
						"encountered error parsing repo url for %s/%s: %+v",
						org, repo.GetName(), err,
					)
					continue
				}

				urlsCh <- repoURL
			}
		}(reposCh, urlsCh)

		go func(uCh chan *url.URL, uSl []*url.URL) {
			for u := range uCh {
				repoURLs = append(uSl, u)
			}
		}(urlsCh, repoURLs)
	}

	_, starErr := s.StarRepositoriesFromURLs(
		repoURLs, notOlderThanMonths, maxConcurrency,
	)
	return starErr
}

// SaveStarredRepository saves a single starred repository to the local cache.
func (s *StarManager) SaveStarredRepository(
	star *github.StarredRepository, wg *sync.WaitGroup,
) error {
	defer wg.Done()

	err := s.db.Save(&Star{
		PushedAt:    star.GetRepository().GetPushedAt().Time,
		StarredAt:   star.StarredAt.Time,
		URL:         star.GetRepository().GetHTMLURL(),
		Language:    strings.ToLower(star.GetRepository().GetLanguage()),
		Stargazers:  star.GetRepository().GetStargazersCount(),
		Description: star.GetRepository().GetDescription(),
		Topics:      star.GetRepository().Topics,
		Archived:    star.GetRepository().GetArchived(),
	})
	if err != nil {
		return err
	}

	log.Infof(
		"Saved %s (with topics %s)\n",
		star.GetRepository().GetHTMLURL(),
		star.GetRepository().Topics,
	)
	return nil
}

// SaveStarredPage saves an entire page of starred repositories concurrently,
// optionally sending server responses to a channel if it is provided.
func (s *StarManager) SaveStarredPage(
	pageno int, responses chan *github.Response,
) chan error {
	wg := sync.WaitGroup{}
	errors := make(chan error)

	page, response, err := s.client.Activity.ListStarred(
		s.context,
		s.username,
		&github.ActivityListStarredOptions{
			ListOptions: github.ListOptions{
				PerPage: PageSize,
				Page:    pageno,
			},
		},
	)
	if err != nil {
		log.Infof(
			"An error occurred while fetching page %d of %s's GitHub stars!\n",
			pageno,
			s.username,
		)

		errors <- err
	}

	if responses != nil {
		responses <- response
	}

	log.Infof("Attempting to save starred projects on page %d...\n", pageno)
	for _, r := range page {
		wg.Add(1)
		go s.SaveStarredRepository(r, &wg)
	}

	wg.Wait()
	close(errors)

	return errors
}

// SaveAllStars saves all of the user's starred repositories
func (s *StarManager) SaveAllStars(maxConcurrency int) error {
	wg := sync.WaitGroup{}
	responses := make(chan *github.Response, 1)

	// Fetch the first page to determine the last page number from the response
	// "Link" header
	log.Info("Attempting to save first page...")
	go s.SaveStarredPage(1, responses)
	firstPageResponse := <-responses

	// We start from 2 by default because we already fetched the first page
	// above, we want to avoid repeating that operation
	startPage := 2
	startIdx := startPage
	numGoroutines := firstPageResponse.LastPage

	// If concurrency is explictly bounded, we alter the parameters to
	// accomodate the requirements. By default we start at startPage, otherwise
	// we start at 0 since we would like to spin up exactly numGoroutunes
	// goroutines.
	if maxConcurrency > 0 {
		numGoroutines = maxConcurrency
		startIdx = 0
	}

	log.Info("Attempting to save the rest of the pages...")
	pagesToSave := make(chan int)
	errsCh := make(chan error)

	for i := startIdx; i <= numGoroutines; i++ {
		wg.Add(1)
		go func(pagesCh chan int, errs chan error, g *sync.WaitGroup) {
			defer g.Done()
			for pageNo := range pagesCh {
				for e := range s.SaveStarredPage(pageNo, nil) {
					errs <- e
				}
			}
		}(pagesToSave, errsCh, &wg)
	}

	for i := startPage; i < firstPageResponse.LastPage+1; i++ {
		pagesToSave <- i
	}
	close(pagesToSave)

	go func() {
		wg.Wait()
		close(errsCh)
	}()

	var finalErrs error
	for err := range errsCh {
		finalErrs = multierr.Append(finalErrs, err)
	}

	return finalErrs
}

// SaveIfEmpty saves all stars if the local cache is empty
func (s *StarManager) SaveIfEmpty(concurrency int) error {
	if count, _ := s.db.Count(&Star{}); count == 0 {
		return s.SaveAllStars(concurrency)
	}

	return nil
}

// KV is a generic struct that maintains a string key - int value pair ( :( )
type KV struct {
	Key   string
	Value int
}

// GetTopics returns topics for a repository, otherwise if no repository is
// passed, returns a list of all topics
func (s *StarManager) GetTopics() []KV {
	stars := []Star{}
	topicCounts := map[string]int{}

	s.db.All(&stars)

	for _, star := range stars {
		for _, topic := range star.Topics {
			topicCounts[topic]++
		}
	}

	results := []KV{}

	for topic, count := range topicCounts {
		results = append(results, KV{topic, count})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Value > results[j].Value
	})

	return results
}

// GetStars returns repositories given a project count to return, and an
// optional language and topic to filter by. It can also randomize the results.
func (s *StarManager) GetStars(
	count int, language, topic string, random bool,
) ([]*Star, error) {
	stars := []*Star{}

	if language != "" {
		if err := s.db.Select(q.Eq("Language", language)).Find(&stars); err != nil {
			return nil, err
		}
	} else {
		if err := s.db.All(&stars); err != nil {
			return nil, err
		}
	}

	if topic != "" {
		topicStars := []*Star{}

		for _, star := range stars {
			if utils.StringInSlice(topic, star.Topics) {
				topicStars = append(topicStars, star)
			}
		}

		stars = topicStars
	}

	if random {
		rand.Seed(time.Now().UTC().UnixNano())
		rand.Shuffle(len(stars), func(i, j int) {
			stars[i], stars[j] = stars[j], stars[i]
		})
	} else {
		sort.Slice(stars, func(i, j int) bool {
			return stars[i].Stargazers > stars[j].Stargazers
		})
	}

	if len(stars) > 0 {
		if len(stars) > count {
			return stars[0:count], nil
		}

		return stars, nil
	}

	return nil, errors.New("No stars matching criteria found")
}

// RemoveStar unstars the repository on Github and removes the star from the
// local cache.
func (s *StarManager) RemoveStar(star *Star, wg *sync.WaitGroup) (bool, error) {
	wg.Add(1)
	defer wg.Done()

	starURL, parseErr := url.Parse(star.URL)
	if parseErr != nil {
		return false, parseErr
	}

	splitPath := strings.Split(starURL.Path, "/")

	_, unstarErr := s.client.Activity.Unstar(s.context, splitPath[1], splitPath[2])
	if unstarErr != nil {
		log.Infof("An error occurred while attempting to unstar %s: %s\n",
			star.URL, unstarErr.Error(),
		)
		return false, unstarErr
	}

	deleteErr := s.db.DeleteStruct(star)
	if deleteErr != nil {
		return false, deleteErr
	}

	log.Infof("Removed %s\n", star.URL)

	return true, nil
}

// Cleanup removes stars older than a specified time in months, optionally
// unstarring archived repositories as well
func (s *StarManager) Cleanup(age int, archived bool) error {
	allStars := []*Star{}
	toDelete := make(chan *Star)
	wg := sync.WaitGroup{}
	then := time.Now().AddDate(0, -age, 0)

	if err := s.db.All(&allStars); err != nil {
		return err
	}

	log.Infof("Filtering stars to delete (from %d)...\n", len(allStars))
	for _, star := range allStars {
		if star.PushedAt.Before(then) || star.Archived == archived {
			log.Infof(
				"Queueing %s for deletion (last pushed at %+v, archive status: %t)\n",
				star.URL,
				star.PushedAt,
				star.Archived,
			)

			wg.Add(1)
			go func(ch chan *Star, s *Star, wg *sync.WaitGroup) {
				defer wg.Done()
				ch <- s
			}(toDelete, star, &wg)
		}
	}

	// Cannot close channel in main goroutine as it will block
	go func() {
		wg.Wait()
		close(toDelete)
	}()

	for star := range toDelete {
		go s.RemoveStar(star, &wg)
	}
	wg.Wait()

	return nil
}
