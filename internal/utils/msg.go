package utils

import (
	"fmt"
	"os"
)

// PrintDebugErrorMsg prints a debug error message to Stderr.
func PrintDebugErrorMsg(err error, debug bool) {
	if debug {
		fmt.Fprintf(os.Stderr, "DEBUG: %s\n", err.Error())
	}
}

// PrintDebugMsg prints a debug message to Stdout.
func PrintDebugMsg(msg string, debug bool) {
	if debug {
		fmt.Fprintf(os.Stdout, "DEBUG: %s\n", msg)
	}
}
