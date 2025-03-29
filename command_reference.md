# Command Reference

This page provides a comprehensive reference guide for all available gh-stars commands, their options, and example usage scenarios.

## Table of Contents

1. [Global Flags](#global-flags)
2. [Commands](#commands)
   - [add](#add)
   - [cleanup](#cleanup)
   - [clear](#clear)
   - [completion](#completion)
   - [save](#save)
   - [show](#show)
   - [topics](#topics)
   - [version](#version)

## Global Flags

The following flags can be used with any command:

- `-w, --concurrency int`: Limit goroutines for network I/O operations (default 10)
- `-h, --help`: Display help information for the command
- `-o, --log-level string`: Set the log level (default "info")

## Commands

### add

Add (star) repositories

Usage: `stars add [flags]`

Flags:
- `-m, --months int`: How many months to add stars for (default 2)
- `-u, --from-url string`: URL to crawl to add new stars from
- `-r, --from-org string`: Organization to add new stars from
- `-s, --from-user string`: User to add new stars from
- `-l, --from-list string`: List of GitHub URLs to add new stars from

Examples:
```bash
# Star repositories from a specific URL
stars add -u https://example.com/awesome-repos

# Star repositories from a GitHub organization
stars add -r awesome-org

# Star repositories from a specific user
stars add -s awesome-user

# Star repositories from a list file
stars add -l repo-list.txt
```

### cleanup

Clean up old stars

Usage: `stars cleanup [flags]`

Flags:
- `-m, --months int`: Number of months to delete projects older than (default 2)
- `-w, --concurrency int`: Number of concurrent cleanup operations
- `-a, --include-archived`: Include archived stars in the cleanup process

Example:
```bash
# Clean up stars older than 3 months, including archived projects
stars cleanup -m 3 -a
```

### clear

Clear local stars cache

Usage: `stars clear`

Example:
```bash
stars clear
```

### completion

Generate shell completion script

Usage: `stars completion [bash|zsh]`

Examples:
```bash
# Generate Bash completion script
stars completion bash > ~/.stars-completion.bash
echo 'source ~/.stars-completion.bash' >> ~/.bashrc

# Generate Zsh completion script
stars completion zsh > ~/.stars-completion.zsh
echo 'source ~/.stars-completion.zsh' >> ~/.zshrc
```

### save

Save starred repositories

Usage: `stars save`

Example:
```bash
stars save
```

### show

Show stars

Usage: `stars show [flags]`

Flags:
- `-c, --count int`: Number of stars to show (default 6)
- `-l, --language string`: Limit to projects written only in this language
- `-t, --topic string`: Limit to projects with this topic
- `-r, --random`: Randomize results
- `-b, --browse`: Open stars in browser instead of writing them to stdout
- `-d, --width int`: Maximum width (as descriptions can sometimes get lengthy)

Examples:
```bash
# Show 10 random stars
stars show -c 10 -r

# Show 5 Go projects
stars show -c 5 -l Go

# Show 3 projects with the "machine-learning" topic and open them in the browser
stars show -c 3 -t machine-learning -b
```

### topics

List all topics of all stars

Usage: `stars topics`

Example:
```bash
stars topics
```

### version

Show version of stars

Usage: `stars version`

Example:
```bash
stars version
```

## Additional Information

For more detailed information on using the gh-stars CLI, please refer to the [README.md](https://github.com/gkze/gh-stars/blob/master/README.md) file in the project repository.