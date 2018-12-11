# stars

A command-line interface to your GitHub stars

## Development

```bash
git clone git@github.com:gkze/stars-go.git
cd stars-go
go build # need Golang 1.11+
```

## Installation

Binaries are available on the releases page. You can also install it easily with:

```bash
go get -u github.com/gkze/stars
```

You also need a `~/.netrc` with a personal access token configured:

```bash
$ cat ~/.netrc
machine api.github.com
    login gkze
    password [your github token here]
```

Usage:

```bash
NAME:
   stars - Command-line interface to YOUR GitHub stars

USAGE:
   stars [global options] command [command options] [arguments...]

VERSION:
   0.3.9

COMMANDS:
     save         Save all stars
     list-topics  list all topics of starred projects
     show         Show popular stars given filters
     clear        Clear local stars cache
     cleanup      Clean up old stars
     help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```
