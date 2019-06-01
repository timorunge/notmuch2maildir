package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// shortUserHomeDirectory describes the shorthand variable which refers to the
// users home directory.
const shortUserHomeDirectory string = "~"

// AbsDir is returning the absolute dir of a path.
func AbsDir(path string) (string, error) {
	if strings.HasPrefix(path, shortUserHomeDirectory) {
		home, err := os.UserHomeDir()
		dir := fmt.Sprintf("%s/%s", home, strings.TrimPrefix(path, shortUserHomeDirectory))
		return filepath.Clean(dir), err
	}
	return filepath.Clean(path), nil
}
