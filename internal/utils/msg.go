package utils

import (
	"fmt"
	"os"
)

// PrintDebugErrorMsg prints a debug error message to Stderr.
func PrintDebugErrorMsg(err error, debug bool) {
	if debug {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("DEBUG: %s", err.Error()))
	}
}

// PrintDebugMsg prints a debug message to Stdout.
func PrintDebugMsg(msg string, debug bool) {
	if debug {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("DEBUG: %s", msg))
	}
}
