# GitHub Star Management Guide

This guide explains how to effectively manage your GitHub stars using gh-stars, a command-line tool for saving, showing, and cleaning up your starred repositories based on various criteria.

## Table of Contents

1. [Installation](#installation)
2. [Saving Stars](#saving-stars)
3. [Showing Stars](#showing-stars)
4. [Adding Stars](#adding-stars)
5. [Cleaning Up Stars](#cleaning-up-stars)
6. [Managing Topics](#managing-topics)

## Installation

To install gh-stars, follow these steps:

1. [Installation instructions to be added]

## Saving Stars

To save all of your starred repositories to the local cache, use the `save` command:

```bash
gh-stars save
```

This command will fetch all of your starred projects and save them to the local filesystem. By default, it uses a concurrency level of 10, but you can adjust this using the `--concurrency` flag:

```bash
gh-stars save --concurrency 5
```

## Showing Stars

To display a list of your starred repositories, use the `show` command:

```bash
gh-stars show
```

You can customize the output using various flags:

- `--count` or `-c`: Specify the number of stars to show (default: 6)
- `--language` or `-l`: Filter by programming language
- `--topic` or `-t`: Filter by topic
- `--random` or `-r`: Randomize the results
- `--browse` or `-b`: Open the starred repositories in your browser
- `--width` or `-d`: Set the maximum width for the output

Examples:

```bash
# Show 10 random stars
gh-stars show --count 10 --random

# Show Go projects with the "cli" topic
gh-stars show --language go --topic cli

# Open the top 5 starred repositories in your browser
gh-stars show --count 5 --browse
```

## Adding Stars

You can add new stars using the `add` command with various options:

- `--from-url` or `-u`: Star repositories from a given URL
- `--from-org` or `-r`: Star repositories from a specific organization
- `--from-user` or `-s`: Star repositories from a specific user
- `--from-list` or `-l`: Star repositories from a list of GitHub URLs

Examples:

```bash
# Star repositories from a URL
gh-stars add --from-url https://example.com/awesome-list

# Star repositories from an organization
gh-stars add --from-org microsoft

# Star repositories from a user
gh-stars add --from-user octocat

# Star repositories from a list file
gh-stars add --from-list my-stars-list.txt
```

## Cleaning Up Stars

To remove old or unwanted stars, use the `cleanup` command:

```bash
gh-stars cleanup
```

By default, this command will remove stars older than 2 months. You can customize the cleanup process using these flags:

- `--months` or `-m`: Specify the number of months to keep (default: 2)
- `--include-archived` or `-a`: Include archived repositories in the cleanup process

Example:

```bash
# Remove stars older than 6 months, including archived repositories
gh-stars cleanup --months 6 --include-archived
```

## Managing Topics

To view a list of all topics associated with your starred repositories, use the `topics` command:

```bash
gh-stars topics
```

This command will display a list of topics sorted by occurrence count, giving you an overview of the most common topics among your starred projects.

## Additional Commands

- `version`: Show the version of gh-stars
- `clear`: Clear the local stars cache
- `completion`: Generate shell completion scripts for bash or zsh

For more information on any command, use the `--help` flag:

```bash
gh-stars [command] --help
```

By using these commands and options, you can effectively manage your GitHub stars, keeping your list organized and relevant to your interests.