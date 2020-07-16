package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/gkze/stars/starmanager"
	"github.com/gkze/stars/utils"
	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	// Version is version information dynamically injected at build time
	Version string

	// concurrency specifies the default global concurrency level (goroutine count)
	// that will be applied to all network I/O operations
	concurrency int

	// global log level
	logLevel string

	// StarManager object
	sm *starmanager.StarManager

	// root command
	starsCmd *cobra.Command
)

func mkVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version of stars",
		Long:  "Displays the version of the currently running stars CLI binary",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("stars version %s\n", Version)
			return nil
		},
	}
}

func mkSaveAllStarsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "save",
		Short: "Save starred repositories",
		Long:  "Fetches all of the current user's starred projects to the local filesystem",
		RunE: func(cmd *cobra.Command, args []string) error {
			return sm.SaveAllStars(concurrency)
		},
	}
}

func mkAddStarsCmd() *cobra.Command {
	var (
		addMonths int
		fromURL   string
		fromUser  string
		fromOrg   string
	)

	addStarsCmd := &cobra.Command{
		Use:   "add",
		Short: "Add (star) repositories",
		Long:  "Star repositories, specified in various ways",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Exactly one of the options must be passed
			if (fromURL == "" && fromUser == "" && fromOrg == "" ||
				fromURL != "" && fromUser != "" && fromOrg != "") ||
				(fromURL != "" && fromUser != "") ||
				(fromUser != "" && fromOrg != "") ||
				(fromURL != "" && fromOrg != "") {

				return errors.New(
					"Can only pass one of: -u/--from-urls, -o/--from-org, -g/--from-user",
				)
			}

			if fromURL != "" {
				if _, err := url.Parse(fromURL); err != nil {
					return fmt.Errorf("invalid URL %s: %w", fromURL, err)
				}

				urls, err := utils.ExtractURLs(fromURL)
				if err != nil {
					return err
				}
				log.Infof("Discovered %d URLs at %s\n", len(urls), fromURL)

				ghUrls := utils.FilterGitHubURLs(urls, starmanager.GitHubHost)
				log.Printf("Found %d GitHub URLs", len(ghUrls))

				if len(ghUrls) == 0 {
					log.Printf("No GitHub URLs found - exiting")
					return nil
				}

				log.Infof("Starring %d GitHub repositories\n", len(ghUrls))
				sm.StarRepositoriesFromURLs(ghUrls, addMonths, concurrency)
			}

			if fromOrg != "" {
				log.Infof("Attempting to repositories from %s\n", fromOrg)
				return sm.StarRepositoriesFromOrg(fromOrg, addMonths, concurrency)
			}

			if fromUser != "" {
				log.Infof("Attempting to star repositories from %s\n", fromUser)
				return sm.StarRepositoriesFromUser(fromUser, addMonths, concurrency)
			}

			return nil
		},
	}

	addStarsCmd.PersistentFlags().IntVarP(&addMonths, "months", "m", 2, "How many")
	addStarsCmd.PersistentFlags().StringVarP(
		&fromURL, "from-url", "u", "", "URL to crawl to add new stars from",
	)
	addStarsCmd.PersistentFlags().StringVarP(
		&fromOrg, "from-org", "r", "", "Organization to add new stars from",
	)
	addStarsCmd.PersistentFlags().StringVarP(
		&fromUser, "from-user", "s", "", "User to add new stars from",
	)

	return addStarsCmd
}

func mkTopicsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "topics",
		Short: "List all topics of all stars",
		Long:  "Displays a list of topics, sorted by occurrece count, for all of a user's starred projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := sm.SaveIfEmpty(concurrency); err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
			for i, pair := range sm.GetTopics() {
				if i == 0 {
					fmt.Fprintf(w, "TOPIC\tOCCURRENCES\n")
				}

				fmt.Fprintf(w, "%s\t%d\n", pair.Key, pair.Value)
			}

			return w.Flush()
		},
	}
}

func mkShowStarsCmd() *cobra.Command {
	var (
		count    int
		language string
		topic    string
		random   bool
		browse   bool
		width    int
	)

	showStarsCmd := &cobra.Command{
		Use:   "show",
		Short: "Show stars",
		Long:  "Displays a tabulated list of stars given project filters",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := sm.SaveIfEmpty(concurrency); err != nil {
				return err
			}

			maxWidth, _, err := terminal.GetSize(0)
			if err != nil {
				return err
			}

			if width > 0 {
				maxWidth = width
			}

			stars, err := sm.GetStars(count, language, topic, random)
			if err != nil {
				return err
			}

			if browse {
				wg := sync.WaitGroup{}
				for _, star := range stars {
					wg.Add(1)
					go func(s *starmanager.Star) {
						defer wg.Done()
						if err := browser.OpenURL(s.URL); err != nil {
							panic(err)
						}
					}(star)
				}
				wg.Wait()
				return nil
			}

			linebuf := utils.NewBoundedLineBuf([]byte{}, maxWidth-1)
			tw := tabwriter.NewWriter(linebuf, 0, 2, 2, ' ', 0)

			if language != "" {
				if _, err := tw.Write([]byte(strings.Join([]string{
					"PUSHED", "STARRED", "STARS", "URL", "DESCRIPTION",
				}, "\t") + "\n")); err != nil {
					return err
				}

				for _, star := range stars {
					pushed := star.PushedAt.Format(time.RFC3339)
					starred := star.StarredAt.Format(time.RFC3339)
					stargazers := strconv.Itoa(star.Stargazers)
					starURL := star.URL
					description := star.Description

					if _, err := tw.Write([]byte(strings.Join([]string{
						pushed, starred, stargazers, starURL, description,
					}, "\t") + "\n")); err != nil {
						return err
					}
				}

				if err := tw.Flush(); err != nil {
					return err
				}

				_, flushErr := linebuf.FlushTo(os.Stdout)
				return flushErr
			}

			if _, err := tw.Write([]byte(strings.Join([]string{
				"PUSHED", "STARRED", "STARS", "LANGUAGE", "URL", "DESCRIPTION",
			}, "\t") + "\n")); err != nil {
				return err
			}

			for _, star := range stars {
				pushed := star.PushedAt.Format(time.RFC3339)
				starred := star.StarredAt.Format(time.RFC3339)
				stargazers := strconv.Itoa(star.Stargazers)
				language = star.Language
				starURL := star.URL
				description := star.Description

				if _, err := tw.Write([]byte(strings.Join([]string{
					pushed, starred, stargazers, language, starURL, description,
				}, "\t") + "\n")); err != nil {
					return err
				}
			}

			if err := tw.Flush(); err != nil {
				return err
			}

			_, flushErr := linebuf.FlushTo(os.Stdout)
			return flushErr
		},
	}

	showStarsCmd.PersistentFlags().IntVarP(
		&count, "count", "c", 6, "Number of stars to show",
	)
	showStarsCmd.PersistentFlags().StringVarP(
		&language, "language", "l", "", "Limit to projects written only in this language",
	)
	showStarsCmd.PersistentFlags().StringVarP(
		&topic, "topic", "t", "", "Limit to projects with this topic",
	)
	showStarsCmd.PersistentFlags().BoolVarP(
		&random, "random", "r", false, "Randomize results",
	)
	showStarsCmd.PersistentFlags().BoolVarP(
		&browse, "browse", "b", false, "Open stars in browser instead of writing them to stdout",
	)
	showStarsCmd.PersistentFlags().IntVarP(
		&width, "width", "d", 0, "Maximum width (as descriptions can sometimes get lengthy)",
	)

	return showStarsCmd
}

func mkClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear local stars cache",
		Long:  "Wipe the file on the local filesystem containing the fetched results of all stars",
		RunE:  func(cmd *cobra.Command, args []string) error { return sm.ClearCache() },
	}
}

func mkCleanupCmd() *cobra.Command {
	var (
		cleanupMonths      int
		includeArchived    bool
		cleanupConcurrency int
	)

	cleanupCmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up old stars",
		Long:  "Un-stars projects older than n months, optionally also unstarring archived projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := sm.SaveIfEmpty(cleanupConcurrency); err != nil {
				return err
			}

			return sm.Cleanup(cleanupMonths, includeArchived)
		},
	}

	cleanupCmd.PersistentFlags().IntVarP(
		&cleanupMonths, "months", "m", 2, "Number of months to delete projects older than",
	)
	cleanupCmd.PersistentFlags().IntVarP(
		&cleanupConcurrency, "concurrency", "w", 2, "Number of months to ",
	)
	cleanupCmd.PersistentFlags().BoolVarP(
		&includeArchived, "include-archived", "a", false, "Include archived stars",
	)

	return cleanupCmd
}

func mkCompletionCmd() *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion script",
		Long:  "Outputs an autocompletion script to be sourced by a target shell",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(`Outputs autocompletion scripts for the CLI. Please refer
to your shell's documentation on how to configure autocompletion.
			`)
			return cmd.Help()
		},
	}

	bashCompletionCmd := &cobra.Command{
		Use:   "bash",
		Short: "Generate bash completion",
		Long:  "Outputs Bash autocompletion script",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.GenBashCompletion(os.Stdout)
		},
	}

	zshCompletionCmd := &cobra.Command{
		Use:   "zsh",
		Short: "Generate Zsh completion",
		Long:  "Outputs Zsh autocompletion script",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.GenZshCompletion(os.Stdout)
		},
	}

	completionCmd.AddCommand(bashCompletionCmd, zshCompletionCmd)

	return completionCmd
}

func main() {
	starsCmd := &cobra.Command{
		Use:   "stars",
		Short: "Stars is a command-line GitHub Stars manager",
		Long: `A CLI written in Golang to facilitate efficient management of a user's
GitHub starred projects / repositories, a.k.a. "Stars"`,
		RunE: func(cmd *cobra.Command, args []string) error { return cmd.Help() },
	}

	starsCmd.PersistentFlags().StringVarP(
		&logLevel, "log-level", "o", "info", "Log level",
	)
	starsCmd.PersistentFlags().IntVarP(
		&concurrency,
		"concurrency",
		"w",
		starmanager.DefaultConcurrency,
		"Limit goroutines for network I/O operations",
	)

	lvl, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Errorf("Error parsing log level: %+v", err)
		os.Exit(1)
	}

	log.Tracef("Setting log level to %+v\n", lvl)
	log.SetLevel(lvl)

	sm, err = starmanager.New(lvl)
	if err != nil {
		log.Printf("Error creating StarManager! %v", err.Error())
	}

	starsCmd.AddCommand(
		mkVersionCmd(),
		mkSaveAllStarsCmd(),
		mkAddStarsCmd(),
		mkTopicsCmd(),
		mkShowStarsCmd(),
		mkClearCmd(),
		mkCleanupCmd(),
		mkCompletionCmd(),
	)

	if err := starsCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
