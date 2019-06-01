package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/mail"
	"os"
	"regexp"

	"github.com/timorunge/notmuch2maildir/internal/nm2md"
	"github.com/timorunge/notmuch2maildir/internal/notmuch"
	"github.com/timorunge/notmuch2maildir/internal/utils"
)

// ThreadOptions describes options for the thread command.
type ThreadOptions struct {
	MsgID string `short:"m" long:"message-id" description:"The message-id of the source mail"`
}

var (
	msgID         string
	threadOptions ThreadOptions
)

// Execute the thread command.
func (cmd *ThreadOptions) Execute(args []string) error {
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
func (cmd *ThreadOptions) Usage() string {
	return "STDIN"
}

func init() {
	parser.AddCommand("thread", "Display a entire mail thread using Notmuch", "Display a entire mail thread using Notmuch", &threadOptions)
}
