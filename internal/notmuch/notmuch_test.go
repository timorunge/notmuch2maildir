package notmuch

import (
	"fmt"
	"strings"
	"testing"
)

type testPairNotmuch struct {
	notmuch     Notmuch
	expectedRes string
}

var (
	testPairs = []testPairNotmuch{
		{
			notmuch: Notmuch{
				Args: []string{
					"--output=files",
					"Foo",
				},
				Command:    "search",
				ConfigFile: "~/.notmuch-config",
				Executable: "notmuch",
			},
			expectedRes: "notmuch --config=~/.notmuch-config search --output=files Foo",
		},
		{
			notmuch: Notmuch{
				Args: []string{
					"--output=files",
					"Bar",
				},
				Command:    "search",
				ConfigFile: "~/.config/notmuch/bar-account",
				Executable: "/usr/local/bin/notmuch",
			},
			expectedRes: "/usr/local/bin/notmuch --config=~/.config/notmuch/bar-account search --output=files Bar",
		},
		{
			notmuch: Notmuch{
				Args: []string{
					"--output=files",
					"FooBar",
				},
				Command:    "search",
				ConfigFile: "/home/foobar/.config/notmuch/foobar-account",
				Executable: "/usr/local/bin/notmuch",
			},
			expectedRes: "/usr/local/bin/notmuch --config=/home/foobar/.config/notmuch/foobar-account search --output=files FooBar",
		},
		{
			notmuch: Notmuch{
				Args: []string{
					"--output=threads",
					"id:1234",
				},
				Command:    "search",
				ConfigFile: "~/.notmuch-config",
				Executable: "notmuch",
			},
			expectedRes: "notmuch --config=~/.notmuch-config search --output=threads id:1234",
		},
	}
)

func TestNotmuchCmd(t *testing.T) {
	for _, pair := range testPairs {
		nm := NewNotmuch(pair.notmuch.Executable, pair.notmuch.ConfigFile, pair.notmuch.Command, pair.notmuch.Args)
		cmd, args := nm.command()
		res := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
		if res != pair.expectedRes {
			t.Error(
				"Expected result is", pair.expectedRes,
				"but got", res,
			)
		}
	}
}
