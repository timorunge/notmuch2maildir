# notmuch2maildir

[![Go Report](https://goreportcard.com/badge/github.com/timorunge/notmuch2maildir)](https://goreportcard.com/report/github.com/timorunge/notmuch2maildir)
[![Build Status](https://travis-ci.org/timorunge/notmuch2maildir.svg?branch=master)](https://travis-ci.org/timorunge/notmuch2maildir)

`notmuch2maildir` is a simple CLI tool written in Go for searching your mails in
a MUA like e.g. ([neo](https://neomutt.org/))[mutt](http://mutt.org/) using the
[Notmuch](https://notmuchmail.org/) mail indexer.

The idea is based of the
[original version](https://notmuchmail.org/notmuch-mutt/) of the `mutt-notmuch`
script and [mutt-notmuch-py](https://github.com/honza/mutt-notmuch-py). The
implementation in Go is - for large search results - significantly faster.

## Install

You can use an [official
release](https://github.com/timorunge/notmuch2maildir/releases) of `notmuch2maildir`.
The tarballs for each release contain the `notmuch2maildir` CLI applicaton.

Copy the binary in your `$PATH` or call it directly via
`$YOURDIR/notmuch2maildir`.

To get the latest version of `notmuch2maildir` just run `go get`.

```sh
go get github.com/timorunge/notmuch2maildir
```

If `$GOPATH/bin` is not in your `$PATH` call `notmuch2maildir` directly via
`$GOPATH/bin/notmuch2maildir`.

## Usage

For the usage it's required that `notmuch` itself is in your `$PATH` (or use
the hidden command line flag `-n` / `--notmuch-executable`).

```sh
Usage:
  notmuch2maildir [OPTIONS] <search | thread | version>

Application Options:
  -c, --notmuch-config=     Notmuch configuration file which should be used (default: ~/.notmuch-config)
  -o, --output-dir=         Output directory for storing the Notmuch search results (default: ~/.cache/notmuch/mutt_results)

Help Options:
  -h, --help                Show this help message

Available commands:
  search   Just search Notmuch
  thread   Display a entire mail thread using Notmuch
  version  Show the version of notmuch2maildir
```

`notmuch2maildir` is not creating the parent directory for the search results.

### Search

The search command is creating a maildir based on the search query.

```
Usage:
  notmuch2maildir [OPTIONS] search QUERY

Just search Notmuch

Application Options:
  -c, --notmuch-config=     Notmuch configuration file which should be used (default: ~/.notmuch-config)
  -o, --output-dir=         Output directory for storing the Notmuch search results (default: ~/.cache/notmuch/mutt_results)

Help Options:
  -h, --help                Show this help message

[search command options]
      -p, --promt           Opens a promt to enter the search query
```

### Thread

The thread command is creating a maildir based on the `message-id` of a
source mail.

```
Usage:
  notmuch2maildir [OPTIONS] thread STDIN

Display a entire mail thread using Notmuch

Application Options:
  -c, --notmuch-config=     Notmuch configuration file which should be used (default: ~/.notmuch-config)
  -o, --output-dir=         Output directory for storing the Notmuch search results (default: ~/.cache/notmuch/mutt_results)

Help Options:
  -h, --help                Show this help message

[thread command options]
      -m, --message-id=     The message-id of the source mail
```

## (neo)mutt configuration

Chose the interactive or the query mode and add the following snippets to your
`muttrc`.

### Search

#### Promt mode

```
macro index / "<enter-command>unset wait_key<enter><shell-escape>notmuch2maildir search -p<enter><change-folder-readonly>~/.cache/notmuch/search_results<enter>" \
            "Search mail (using Notmuch)"
```

#### Query mode
```
macro index / "<enter-command>unset wait_key<enter><shell-escape>read -p 'Search query: ' query; notmuch2maildir search -q \$query<enter><change-folder-readonly>~/.cache/notmuch/search_results<enter>" \
            "Search mail (using Notmuch)"
```

### Reconstruct thread

```
macro index T "<enter-command>unset wait_key<enter><pipe-message>notmuch2maildir thread<enter><change-folder-readonly>~/.cache/notmuch/search_results<enter>" \
            "Search and reconstruct thread (Using notmuch)"
```

## License

[BSD 3-Clause "New" or "Revised" License](LICENSE)

## Author Information

- Timo Runge
