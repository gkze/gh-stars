package utils

import (
	"testing"

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

func TestCreateIfNotExists(t *testing.T) {
	t.SkipNow()
}
