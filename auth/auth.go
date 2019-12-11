package auth

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/jdxcode/netrc"
)

const (
	// NetrcUsernameField is the value that holds the username in the netrc format
	NetrcUsernameField string = "login"

	// NetrcPasswordField is the value that holds the passwod in the netrc format
	NetrcPasswordField string = "password"

	// NetrcDefaultFilename is the default name of the netrc configuration file.
	NetrcDefaultFilename string = ".netrc"
)

// Interface is a generic authentication interface
type Interface interface {
	// GetAuth retrieves the authentication credentials for a given host, or
	// throws an error
	GetUsernamePassword(host string) (string, string, error)
}

// Config represents a configuration structure passed to the Netrc object
// in order to initialize it
type Config struct {
	User     *user.User
	Filename string
}

// NewConfig returns a new configuration structure for the netrc authenticator.
func NewConfig() (*Config, error) {
	curUser, err := user.Current()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(filepath.Join(curUser.HomeDir, NetrcDefaultFilename)); os.IsNotExist(err) {
		return nil, fmt.Errorf("$HOME/.netrc does not exist")
	}

	return &Config{
		User:     curUser,
		Filename: NetrcDefaultFilename,
	}, nil
}

// NetrcAuth reresents the implementation of the netrc authentication manager
// interface
type NetrcAuth struct {
	// Config is the reference to the config struct
	Config *Config

	// Netrc is the parsed netrc file
	Netrc *netrc.Netrc
}

// NewNetrc creates a new netrc auth manager, or returns an error
func NewNetrc(cfg *Config) (*NetrcAuth, error) {
	auth := &NetrcAuth{Config: cfg}

	n, err := auth.ParseNetrc()
	if err != nil {
		return nil, err
	}

	auth.Netrc = n

	return auth, nil
}

// ParseNetrc parses the netrc that is given to the auth manager
func (a *NetrcAuth) ParseNetrc() (*netrc.Netrc, error) {
	netrc, err := netrc.Parse(filepath.Join(
		a.Config.User.HomeDir,
		a.Config.Filename,
	))
	if err != nil {
		return nil, err
	}

	return netrc, nil
}

// GetAuth retrieves authentication credentials from the parsed netrc file,
// given that the host exists
func (a *NetrcAuth) GetAuth(host string) (string, string, error) {
	netrcHost := a.Netrc.Machine(host)
	if netrcHost == nil {
		return "", "", fmt.Errorf("no auth for %s configured", host)
	}

	return netrcHost.Get(NetrcUsernameField), netrcHost.Get(NetrcPasswordField), nil
}
