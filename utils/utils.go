package utils

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/jdxcode/netrc"
)

// GetNetrcAuth - Returns the username and password (in this case the API token) for a given host
// configured in .netrc in the user's home directory.
func GetNetrcAuth(hostname string) (string, string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", "", err
	}

	netrcPath := filepath.Join(usr.HomeDir, ".netrc")

	if _, err := os.Stat(netrcPath); os.IsNotExist(err) {
		return "", "", err
	}

	n, err := netrc.Parse(filepath.Join(usr.HomeDir, ".netrc"))
	if err != nil {
		return "", "", err
	}

	auth := n.Machine(hostname)

	return auth.Get("login"), auth.Get("password"), nil
}

// StringInSlice checks whether a given string is in a slice
func StringInSlice(s string, sl []string) bool {
	for _, c := range sl {
		if c == s {
			return true
		}
	}

	return false
}

// CreateIfNotExists examines a path and if it is not present, creates the passed file type for the
// given path.
func CreateIfNotExists(path string, mode os.FileMode) error {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		switch mode {
		case os.ModeDir:
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return err
			}
		case 0:
			_, err := os.Create(path)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
