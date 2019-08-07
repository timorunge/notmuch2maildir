// +build mage

package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/magefile/mage/sh"
)

var (
	buildDate = time.Now().UTC().Format(time.RFC3339)
	buildFile = "./cmd/notmuch2maildir/notmuch2maildir.go"
	ldflags   = "-w -s -X main.buildDate=%s -X main.gitCommit=%s -X main.version=%s"
)

func init() {
	os.Setenv("GO111MODULE", "on")
}

// Build notmuch2maildir binary
func Build() error {
	return sh.Run("go", "build", "-ldflags", fmt.Sprintf(ldflags, buildDate, gitCommit(), gitBranch()), buildFile)
}

// Ci runs several mage tasks at once
func Ci() error {
	if err := Fmt(); err != nil {
		return err
	}
	if err := Test(); err != nil {
		return err
	}
	if err := Build(); err != nil {
		return err
	}
	if err := Clean(); err != nil {
		return err
	}
	return nil
}

// Clean removes object files
func Clean() error {
	sh.Rm("./notmuch2maildir")
	return sh.Run("go", "clean")
}

// Fmt runs gofmt over all go files
func Fmt() error {
	files, err := goFiles()
	if err != nil {
		return err
	}
	failed := false
	for _, file := range files {
		_, err := sh.Output("gofmt", "-l", file)
		if err != nil {
			fmt.Printf("gofmt error on %s: %v", file, err)
			failed = true
		}
	}
	if failed {
		return errors.New("gofmt failed")
	}
	return nil
}

// Install notmuch2maildir binary
func Install() error {
	return sh.Run("go", "install", "-ldflags", fmt.Sprintf(ldflags, buildDate, gitCommit(), gitBranch()), buildFile)
}

// Test notmuch2maildir
func Test() error {
	return sh.Run("go", "test", "./...")
}

// gitBranch is getting the used git branch.
func gitBranch() string {
	buf := &bytes.Buffer{}
	_, err := sh.Exec(nil, buf, nil, "git", "rev-parse", "--abbrev-ref", "HEAD", "-q")
	if err != nil {
		return "Unknown"
	}
	return strings.TrimSuffix(buf.String(), "\n")
}

// gitCommit is getting the used git commit.
func gitCommit() string {
	buf := &bytes.Buffer{}
	_, err := sh.Exec(nil, buf, nil, "git", "show", "--format=%H", "HEAD", "-q")
	if err != nil {
		return "Unknown"
	}
	return strings.TrimSuffix(buf.String(), "\n")
}

// goFiles finds all go files.
func goFiles() ([]string, error) {
	goFiles := []string{}
	err := filepath.Walk(".", func(path string, file os.FileInfo, err error) error {
		if ".go" == filepath.Ext(path) {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	return goFiles, err
}
