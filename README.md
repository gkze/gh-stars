# stars

A command-line interface to your Github Stars. Some useful features:

* Downloads metadata about all of your starred projects and saves it to disk
* Unstars projects older than `n` months (by default, 2)
  * Also unstars projects that have been archived (by default - you can opt out).
* Can let you display starred projects by criteria:
  * Language
  * Topics (labels)
  * Randomly
* Can limit displayed results as specified
* Can open queried starred projects in your browser for viewing

**_NOTE:_** Currently only macOS is supported. Support for other platforms will
be considered if there is demand.

My personal workflow is to save all of my stars, prune old and archived ones,
and display several random stars in my browser for me to view / explore. This
is a type of [Spaced Repetation Learning](https://en.wikipedia.org/wiki/Spaced_repetition)
(think flash cards), that way I can stay relatively up-to-date on what my starred
projects are. This is useful to me when I build software and need to know if
there is a project already out there that solves my problems / fits my needs.

## Development

To get started, you will need [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
and [Go](https://golang.org/doc/install) on your system. Then, you can run the
following commands to build the binary:

```bash
git clone git@github.com:gkze/stars.git
cd stars
go build # need Golang 1.11+
```

**_NOTE:_** As mentioned in the comment above, you will need to have Go 1.11
installed at minimum. This project utilizes [Go modules](https://github.com/golang/go/wiki/Modules),
which are only supported in Go 1.11 and above.

## Installation

There are various methods availabel to install `stars` on your system:

### Homebrew

```bash
brew install gkze/gkze/stars
```

### Go

```bash
go get -u github.com/gkze/stars/cmd/stars
```

Binaries are also available on the releases page.

## Configuration

You will need a `~/.netrc` with a [personal access token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/) configured:

```bash
$ cat ~/.netrc
machine api.github.com
    login [your github username here]
    password [your github token here]
```

## Usage

```bash
NAME:
   stars - Command-line interface to your GitHub stars

USAGE:
   stars [global options] command [command options] [arguments...]

VERSION:
   0.4.12

COMMANDS:
     save     Save all stars
     topics   list all topics of starred projects
     show     Show popular stars given filters
     clear    Clear local stars cache
     cleanup  Clean up old stars
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```
