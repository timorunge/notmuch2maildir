// notmuch command execution.

package nm2md

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func runNotmuch(ctx context.Context, executable, configFile, command string, args []string) ([]string, error) {
	out, err := exec.CommandContext(ctx, executable, notmuchArgs(configFile, command, args)...).Output() //nolint:gosec // executable is user-controlled CLI input; no shell invocation, so not an injection vector
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return nil, fmt.Errorf("notmuch %s: %s", command, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("notmuch %s: %w", command, err)
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines, nil
}

func notmuchArgs(configFile, command string, args []string) []string {
	out := make([]string, 0, 3+len(args))
	if configFile != "" {
		out = append(out, "--config="+configFile)
	}
	out = append(out, command, "--")
	out = append(out, args...)
	return out
}
