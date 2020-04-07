package utils

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/scylladb/go-set"
	"github.com/spf13/afero"
	"go.uber.org/multierr"
	xurls "mvdan.cc/xurls/v2"
)

// StringInSlice checks whether a given string is in a slice
func StringInSlice(s string, sl []string) bool {
	for _, c := range sl {
		if c == s {
			return true
		}
	}

	return false
}

// CreateIfNotExists examines a path and if it is not present, creates the
// passed file type for the given path
func CreateIfNotExists(path string, mode os.FileMode, fs afero.Fs) error {
	_, err := fs.Stat(path)
	if err != nil && os.IsNotExist(err) {
		switch mode {
		case os.ModeDir:
			err := fs.MkdirAll(path, 0755)
			if err != nil {
				return err
			}
		case 0:
			_, err := fs.Create(path)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ExtractURLs extracts a list of valid URLs from a given URL
func ExtractURLs(urlStr string) ([]*url.URL, error) {
	urls := []*url.URL{}
	errs := []error{}
	body := []byte{}

	resp, err := http.Get(urlStr)
	if err != nil {
		return urls, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return urls, err
		}
	} else {
		return urls, fmt.Errorf("received unsuccessful response: %+v", resp)
	}

	results := xurls.Strict().FindAll(body, -1)
	if results == nil {
		return urls, errors.New("no URLs found")
	}

	for _, res := range results {
		u, err := url.Parse(string(res))
		if err != nil {
			errs = append(errs, err)
		}

		urls = append(urls, u)
	}

	if len(errs) > 0 {
		return urls, multierr.Combine(errs...)
	}

	return urls, nil
}

// FilterGitHubURLs returns a list of valid GitHub URLs given a list of URLs
func FilterGitHubURLs(urls []*url.URL, host string) []*url.URL {
	extracted := []*url.URL{}
	reserved := set.NewStringSet("trending", "site", "privacy", "terms")

	isValidPath := func(u *url.URL) bool {
		pathParts := strings.Split(strings.Trim(u.EscapedPath(), "/"), "/")

		// hostname has to match target github host
		return u.Hostname() == host &&
			// path does not contain reserved keywords
			!set.NewStringSet(pathParts...).HasAny(reserved.List()...) &&
			// path only has two parts (e.g. owner/repo)
			len(pathParts) == 2
	}

	for _, u := range urls {
		if isValidPath(u) {
			extracted = append(extracted, u)
		}
	}

	return extracted
}

// BoundedLineBuf is a io.Writer-compatible object that absorbs written bytes
// and truncates each line in its buffer to the specified max line length
// when FlushTo is called
type BoundedLineBuf struct {
	lineLen int
	buf     []byte
}

// NewBoundedLineBuf constructs a new BoundedLineBuf
func NewBoundedLineBuf(buf []byte, lineLen int) *BoundedLineBuf {
	return &BoundedLineBuf{buf: buf, lineLen: lineLen}
}

// Write satisfies io.Writer for BoundedLineBuf
func (blb *BoundedLineBuf) Write(p []byte) (int, error) {
	blb.buf = append(blb.buf, p...)
	return len(p), nil
}

// FlushTo flushes the buffer to a target io.Writer
func (blb *BoundedLineBuf) FlushTo(w io.Writer) (int, error) {
	totalWritten := 0

	line := []byte{}
	for _, chr := range blb.buf {
		line = append(line, chr)

		if chr == '\n' {
			if blb.lineLen > 0 && len(line) > blb.lineLen {
				line = append(line[0:blb.lineLen-3], []byte("...\n")...)
			}

			written, err := w.Write(line)
			totalWritten += written

			if err != nil {
				return totalWritten, err
			}

			line = nil
			continue
		}
	}

	return totalWritten, nil
}

// Compile-time interface satisfaction check
var _ io.Writer = (*BoundedLineBuf)(nil)
