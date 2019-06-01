package utils

import (
	"fmt"
	"os"
	"testing"
)

type testPairAbsDir struct {
	path        string
	expectedRes string
}

var (
	testPairs = []testPairAbsDir{
		{
			path:        "~/.cache/mutt_results",
			expectedRes: fmt.Sprintf("%s/%s", userHomeDir, ".cache/mutt_results"),
		},
		{
			path:        "/home/foo/.cache/mutt_results/",
			expectedRes: "/home/foo/.cache/mutt_results",
		},
		{
			path:        "/home/bar/tmp/../.cache/mutt_results",
			expectedRes: "/home/bar/.cache/mutt_results",
		},
	}
	userHomeDir, _ = os.UserHomeDir()
)

func TestAbsDir(t *testing.T) {
	for _, pair := range testPairs {
		res, _ := AbsDir(pair.path)
		if res != pair.expectedRes {
			t.Error(
				"For", pair.path,
				"the expected result is", pair.expectedRes,
				"but got", res,
			)
		}
	}
}
