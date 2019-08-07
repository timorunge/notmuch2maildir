package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/mail"
	"os"
	"regexp"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/timorunge/notmuch2maildir/internal/nm2md"
	"github.com/timorunge/notmuch2maildir/internal/notmuch"
	"github.com/timorunge/notmuch2maildir/internal/utils"
)

var (
	buildDate = "Unknown"
	gitCommit = "Unknown"
	parser    = flags.NewParser(&applicationOptions, flags.Default)
	version   = "Unknown"

	applicationOptions struct {
		Debug             bool   `long:"debug" description:"Print debugging output" hidden:"true"`
		NotmuchConfigFile string `short:"c" long:"notmuch-config" description:"Notmuch configuration file which should be used" default:"~/.notmuch-config"`
		NotmuchExecutable string `short:"n" long:"notmuch-executable" description:"Notmuch executeable which should be used" default:"notmuch" hidden:"true"`
		OutputDir         string `short:"o" long:"output-dir" description:"Output directory for storing the Notmuch search results" default:"~/.cache/notmuch/search_results" required:"true"`
	}
)

func init() {
	parser.AddCommand("search", "Just search Notmuch", "Just search Notmuch", &searchCommand)
	parser.AddCommand("thread", "Display a entire mail thread using Notmuch", "Display a entire mail thread using Notmuch", &threadCommand)
	parser.AddCommand("version", "Show the version of notmuch2maildir", "", &versionCommand)
}

func main() {
	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
}

// SearchCommand describes options for the search command.
type SearchCommand struct {
	Promt bool `short:"p" long:"promt" description:"Opens a promt to enter the search query"`
}

var (
	searchCommand SearchCommand
	searchQuery   string
)

// Execute the search command.
func (cmd *SearchCommand) Execute(args []string) error {
	if err := validateOptions(); err != nil {
		return err
	}

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
func (cmd *SearchCommand) Usage() string {
	return "QUERY"
}

// ThreadCommand describes options for the thread command.
type ThreadCommand struct {
	MsgID string `short:"m" long:"message-id" description:"The message-id of the source mail"`
}

var (
	msgID         string
	threadCommand ThreadCommand
)

// Execute the thread command.
func (cmd *ThreadCommand) Execute(args []string) error {
	if err := validateOptions(); err != nil {
		return err
	}

	if cmd.MsgID != "" {
		msgID = cmd.MsgID
	} else {
		stdinFile, err := os.Stdin.Stat()
		if err != nil {
			utils.PrintDebugErrorMsg(err, applicationOptions.Debug)
			return errors.New("Can not stat stdin")
		}
		if stdinFile.Size() > 0 {
			stdin, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				utils.PrintDebugErrorMsg(err, applicationOptions.Debug)
				return errors.New("Can not read data from stdin")
			}
			mail, err := mail.ReadMessage(bytes.NewReader(stdin))
			if err != nil {
				utils.PrintDebugErrorMsg(err, applicationOptions.Debug)
				return errors.New("Can not read mail from stdin")
			}
			mailHeader := mail.Header
			msgID = mailHeader.Get("message-id")
		} else {
			return errors.New("No data in stdin")
		}
	}
	if msgID == "" {
		return errors.New("No message-id defined")
	}
	plainMsgID := regexp.MustCompile(`<|>`).ReplaceAllString(msgID, "")

	nm := notmuch.NewNotmuch(
		applicationOptions.NotmuchExecutable,
		applicationOptions.NotmuchConfigFile,
		"search",
		[]string{"--output=threads", fmt.Sprintf("id:%s", plainMsgID)})
	threadIDs, err := nm.Run()
	if err != nil {
		utils.PrintDebugErrorMsg(err, applicationOptions.Debug)
		return errors.New("Notmuch returned a non zero exit status")
	}
	if len(threadIDs) == 0 {
		return errors.New("Notmuch was not able to find the mail thread")
	}

	return nm2md.NewNM2MD(
		nm2md.ApplicationOptions{
			Debug:             applicationOptions.Debug,
			NotmuchConfigFile: applicationOptions.NotmuchConfigFile,
			NotmuchExecutable: applicationOptions.NotmuchExecutable,
			OutputDir:         applicationOptions.OutputDir},
		threadIDs[0]).Execute()
}

// Usage is improving the usage section for the thread command.
func (cmd *ThreadCommand) Usage() string {
	return "STDIN"
}

// VersionCommand is an empty struct.
type VersionCommand struct{}

var versionCommand VersionCommand

// Execute the version command.
func (cmd *VersionCommand) Execute(args []string) error {
	fmt.Printf("Version:    %s\nGit commit: %s\nBuild at:   %s",
		version,
		gitCommit,
		buildDate)
	return nil
}

// validateOptions is validating the CLI option values.
func validateOptions() error {
	_, err := os.Stat(applicationOptions.NotmuchConfigFile)
	if os.IsNotExist(err) {
		utils.PrintDebugErrorMsg(err, applicationOptions.Debug)
		return errors.New("Notmuch configuration file not found")
	}
	absDir, err := utils.AbsDir(applicationOptions.OutputDir)
	if err != nil {
		utils.PrintDebugErrorMsg(err, applicationOptions.Debug)
		return errors.New("Can not get output directory")
	}
	dir, err := os.Stat(absDir)
	if err != nil {
		utils.PrintDebugErrorMsg(err, applicationOptions.Debug)
		return fmt.Errorf("Can not find directory \"%s\"", applicationOptions.OutputDir)
	}
	if !dir.IsDir() {
		utils.PrintDebugErrorMsg(err, applicationOptions.Debug)
		return fmt.Errorf("\"%s\" is an existing file, not a directory", dir)
	}
	return nil
}
