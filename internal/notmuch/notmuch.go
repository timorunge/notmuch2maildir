package notmuch

import (
	"fmt"
	"os/exec"
	"strings"
)

// Notmuch represents the command to execute.
type Notmuch struct {
	Args       []string
	Command    string
	ConfigFile string
	Executable string
}

// NewNotmuch is returning the struct for a new Notmuch command.
func NewNotmuch(executeable string, configFile string, command string, args []string) *Notmuch {
	return &Notmuch{
		Args:       args,
		Command:    command,
		ConfigFile: configFile,
		Executable: executeable,
	}
}

// Run runs a Notmuch command and returns a list of Stdout output.
func (n Notmuch) Run() ([]string, error) {
	cmd, args := n.command()
	stdOut, err := exec.Command(cmd, args...).Output()
	res := strings.Split(string(stdOut), "\n")
	res = res[:len(res)-1]
	return res, err
}

// command is returning the executeable, the command and all arguments (in a
// slice).
func (n Notmuch) command() (string, []string) {
	var args []string
	if n.ConfigFile != "" {
		args = append(args, fmt.Sprintf("%s=%s", "--config", n.ConfigFile))
	}
	args = append(args, n.Command)
	if len(n.Args) > 0 {
		for _, arg := range n.Args {
			args = append(args, arg)
		}
	}
	return n.Executable, args
}
