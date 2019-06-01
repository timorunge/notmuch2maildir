package cmd

import (
	"os"

	"github.com/jessevdk/go-flags"
)

// ApplicationOptions describes all available options.
type ApplicationOptions struct {
	Debug             bool   `long:"debug" description:"Print debugging output" hidden:"true"`
	NotmuchConfigFile string `short:"c" long:"notmuch-config" description:"Notmuch configuration file which should be used" default:"~/.notmuch-config"`
	NotmuchExecutable string `short:"n" long:"notmuch-executable" description:"Notmuch executeable which should be used" default:"notmuch" hidden:"true"`
	OutputDir         string `short:"o" long:"output-dir" description:"Output directory for storing the Notmuch search results" default:"~/.cache/notmuch/search_results" required:"true"`
}

var (
	parser             = flags.NewParser(&applicationOptions, flags.Default)
	applicationOptions ApplicationOptions
)

// Execute is starting the execution of the notmuch2maildir.
func Execute() {
	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
}
