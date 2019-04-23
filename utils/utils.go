package utils

import (
	"os"
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
