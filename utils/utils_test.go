package utils

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestStringInSlice(t *testing.T) {
	testCases := []struct {
		s    string
		sl   []string
		sIns bool
	}{
		{
			s:    "sub",
			sl:   []string{"sub", "string"},
			sIns: true,
		},
		{
			s:    "notsub",
			sl:   []string{"sub", "string"},
			sIns: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, StringInSlice(tc.s, tc.sl), tc.sIns)
	}
}

func doTestCreateIfNotExists(
	t *testing.T,
	path string,
	mode os.FileMode,
	exists bool,
) {
	fs := afero.NewMemMapFs()

	err := CreateIfNotExists(path, mode, fs)
	assert.NoError(t, err)

	fi, err := fs.Stat(path)
	assert.NoError(t, err)

	switch mode {
	case os.ModeDir:
		assert.True(t, fi.Mode().IsDir())
	case 0:
		assert.True(t, fi.Mode().IsRegular())
	}
}

func TestCreateIfNotExists(t *testing.T) {
	testCases := []struct {
		path   string
		mode   os.FileMode
		exists bool
	}{
		{
			path:   "/tmp/nonexistenttestfile",
			mode:   0,
			exists: false,
		},
		{
			path:   "/tmp/nonexistenttestdir",
			mode:   os.ModeDir,
			exists: false,
		},
	}

	for _, tc := range testCases {
		doTestCreateIfNotExists(t, tc.path, tc.mode, tc.exists)
	}
}
