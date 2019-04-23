package auth

import (
	"io/ioutil"
	"os"
	"os/user"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Netrc is a valid netrc configuration file
var Netrc = `
machine host.domain.tld
	login user
	password secret
`

// ValidUser is a valid root user
var ValidUser = &user.User{
	Uid:      "1",
	Gid:      "1",
	Username: "root",
	HomeDir:  "/",
}

// InvalidUser is an invalid user
var InvalidUser = &user.User{
	Uid:      "99999",
	Gid:      "99999",
	Username: "invaliduser",
	HomeDir:  "/invaliduser",
}

// Tests whether
func TestAuthParseNetrcInvalidUser(t *testing.T) {
	MockAuth := &NetrcAuth{
		Config: &Config{
			User:     InvalidUser,
			Filename: ".netrc",
		},
	}

	n, err := MockAuth.ParseNetrc()
	assert.Error(t, err)
	assert.Nil(t, n)
}

func TestAuthParseNetrcInvalidNetrcFilename(t *testing.T) {
	MockAuth := &NetrcAuth{
		Config: &Config{
			User:     ValidUser,
			Filename: ".invalidnetrcfilename",
		},
	}

	n, err := MockAuth.ParseNetrc()
	assert.Error(t, err)
	assert.Nil(t, n)
}

func TestAuthParseNetrcValidUser(t *testing.T) {
	tfd, err := ioutil.TempFile("/tmp", ".netrc")
	assert.NoError(t, err)
	assert.NotNil(t, tfd)

	tempName := tfd.Name()

	bw, err := tfd.WriteString(Netrc)
	assert.NoError(t, err)
	assert.NotEqual(t, bw, 0)

	assert.NoError(t, tfd.Close())

	MockAuth := &NetrcAuth{
		Config: &Config{
			User:     ValidUser,
			Filename: tempName,
		},
	}

	n, err := MockAuth.ParseNetrc()
	assert.NoError(t, err)
	assert.NotNil(t, n)

	// This is normally done inside New() but we set it manually here
	MockAuth.Netrc = n

	user, pass, err := MockAuth.GetAuth("host.domain.tld")
	assert.NoError(t, err)
	assert.Equal(t, user, "user")
	assert.Equal(t, pass, "secret")

	os.Remove(tempName)
}
