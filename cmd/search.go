package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/timorunge/notmuch2maildir/internal/nm2md"
	"github.com/timorunge/notmuch2maildir/internal/utils"
)

// SearchOptions describes options for the search command.
type SearchOptions struct {
	Promt bool `short:"p" long:"promt" description:"Opens a promt to enter the search query"`
}

var (
	searchOptions SearchOptions
	searchQuery   string
)

// Execute the search command.
func (cmd *SearchOptions) Execute(args []string) error {
	if cmd.Promt {
		fmt.Println("Search query:")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		searchQuery = input
		if err != nil {
			utils.PrintDebugErrorMsg(err, applicationOptions.Debug)
			return errors.New("Can not get the search query from stdin")
		}
	} else {
		searchQuery = strings.Join(args, " ")
	}
	if searchQuery == "" {
		return errors.New("No search query defined. Use `-p, --promt' for a promt")
	}

	return nm2md.NewNM2MD(
		nm2md.ApplicationOptions{
			Debug:             applicationOptions.Debug,
			NotmuchConfigFile: applicationOptions.NotmuchConfigFile,
			NotmuchExecutable: applicationOptions.NotmuchExecutable,
			OutputDir:         applicationOptions.OutputDir},
		searchQuery).Execute()
}

// Usage is improving the usage section for the search command.
func (cmd *SearchOptions) Usage() string {
	return "QUERY"
}

func init() {
	parser.AddCommand("search", "Just search Notmuch", "Just search Notmuch", &searchOptions)
}
