# notmuch2maildir

[![CI](https://github.com/timorunge/notmuch2maildir/actions/workflows/ci.yml/badge.svg)](https://github.com/timorunge/notmuch2maildir/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/timorunge/notmuch2maildir)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/timorunge/notmuch2maildir)](https://goreportcard.com/report/github.com/timorunge/notmuch2maildir)
[![License](https://img.shields.io/github/license/timorunge/notmuch2maildir)](LICENSE)
[![Release](https://img.shields.io/github/v/release/timorunge/notmuch2maildir)](https://github.com/timorunge/notmuch2maildir/releases)

A simple CLI tool written in Go for searching your mails in
a MUA like [neomutt](https://neomutt.org/) or [mutt](http://mutt.org/) using the
[notmuch](https://notmuchmail.org/) mail indexer.

The idea is based on the
[original version](https://notmuchmail.org/notmuch-mutt/) of the `mutt-notmuch`
script and [mutt-notmuch-py](https://github.com/honza/mutt-notmuch-py). The
implementation in Go is - for large search results - significantly faster.

## Installation

### Go install

```bash
go install github.com/timorunge/notmuch2maildir/cmd/notmuch2maildir@latest
```

### Binary download

Download pre-built binaries for your platform from the
[releases page](https://github.com/timorunge/notmuch2maildir/releases).

### Build from source

```bash
git clone https://github.com/timorunge/notmuch2maildir.git
cd notmuch2maildir
make build
```

## Quick Start

For the usage it's required that `notmuch` itself is in your `$PATH` (or use
`--notmuch-executable` to specify the path).

Search your mail:

```bash
notmuch2maildir "from:alice subject:meeting"
```

Reconstruct a full thread from a message:

```bash
notmuch2maildir -t -m "<message-id>"
```

## Usage

```
notmuch2maildir - Search mail and reconstruct threads using notmuch

Usage:
  notmuch2maildir [OPTIONS] QUERY
  notmuch2maildir -p
  notmuch2maildir -t -m <message-id>
  notmuch2maildir -t < email.eml

Options:
  -t, --thread                      Thread mode: reconstruct a full mail thread
  -p, --prompt                      Open a prompt to enter the search query
  -m, --message-id string           Message-ID for thread reconstruction
      --notmuch-config string       notmuch configuration file (default: notmuch's own resolution)
      --output-dir string           Output directory for search results (default "$XDG_CACHE_HOME/notmuch/search_results")
      --notmuch-executable string   Path to notmuch binary (default "notmuch")
  -h, --help                        Show this help message
      --version                     Show the version of notmuch2maildir
```

### Configuration

By default, notmuch2maildir lets notmuch resolve its own configuration using
its built-in search order:

1. `--notmuch-config` flag (passed through as `notmuch --config`)
2. `NOTMUCH_CONFIG` environment variable
3. `$XDG_CONFIG_HOME/notmuch/<profile>/config` (or `default` if no profile)
4. `$HOME/.notmuch-config`

Most users don't need `--notmuch-config` at all.

### Search

Search mode is the default. Pass your query as positional arguments:

```bash
notmuch2maildir "from:alice subject:meeting"
notmuch2maildir from:bob date:2024..
```

Use `--` to separate queries that start with a dash:

```bash
notmuch2maildir -- -tag:spam from:alice
```

Use `-p` / `--prompt` for interactive input:

```bash
notmuch2maildir -p
```

### Thread

Thread mode (`-t` / `--thread`) reconstructs a full mail thread. Provide the
message-id via `-m` or pipe a mail to stdin:

```bash
notmuch2maildir -t -m "<message-id>"
notmuch2maildir -t < email.eml
```

## (neo)mutt configuration

Choose the interactive or the query mode and add the following snippets to your
`muttrc`.

### Search

#### Prompt mode

```
macro index / "<enter-command>unset wait_key<enter><shell-escape>notmuch2maildir -p<enter><change-folder-readonly>~/.cache/notmuch/search_results<enter>" \
            "Search mail (using notmuch)"
```

#### Query mode
```
macro index / "<enter-command>unset wait_key<enter><shell-escape>read -p 'Search query: ' query; notmuch2maildir \$query<enter><change-folder-readonly>~/.cache/notmuch/search_results<enter>" \
            "Search mail (using notmuch)"
```

### Reconstruct thread

```
macro index T "<enter-command>unset wait_key<enter><pipe-message>notmuch2maildir -t<enter><change-folder-readonly>~/.cache/notmuch/search_results<enter>" \
            "Search and reconstruct thread (Using notmuch)"
```

## Development

```bash
make help     # Show all available targets
make check    # Run all quality gates (fmt, tidy, vet, lint, test)
make lint     # Run golangci-lint
make test     # Run tests with race detector
make build    # Build static binary
```

## License

[BSD 3-Clause "New" or "Revised" License](LICENSE)
