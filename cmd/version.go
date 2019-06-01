package cmd

import (
	"fmt"

	"github.com/timorunge/notmuch2maildir/internal/nm2md"
)

// VersionOptions is an empty struct.
type VersionOptions struct{}

var versionOptions VersionOptions

// Execute the version command.
func (cmd *VersionOptions) Execute(args []string) error {
	fmt.Printf("notmuch2maildir v%s", nm2md.Version)
	return nil
}

func init() {
	parser.AddCommand("version", "Show the version of notmuch2maildir", "", &versionOptions)
}
