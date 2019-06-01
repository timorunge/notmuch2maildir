// +build mage

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/sh"
)

func init() {
	os.Setenv("GO111MODULE", "on")
}

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

// Build notmuch2maildir binary
func Build() error {
	return sh.Run("go", "build")
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
	return sh.Run("go", "install")
}

// Test notmuch2maildir
func Test() error {
	return sh.Run("go", "test", "./...")
}
