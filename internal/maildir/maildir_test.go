package maildir

import (
	"os"
	"path/filepath"
	"testing"
)

var (
	userCacheDir, _ = os.UserCacheDir()
	testMaildir     = NewMaildir(filepath.Join(userCacheDir, "notmuch2maildir-test"), DefaultPerm)
)

func TestMaildirCreate(t *testing.T) {
	err := testMaildir.Create()
	if err != nil {
		t.Fatal(
			"Tried to create", testMaildir.Path,
			"but got", err,
		)
	}
}

func TestMaildirIsExist(t *testing.T) {
	if !testMaildir.IsExist() {
		t.Error(
			"Maildir should exist but it is not",
		)
	}
}

func TestMaildirClear(t *testing.T) {
	err := testMaildir.Clear()
	if err != nil {
		t.Error(
			"Tried to clear", testMaildir.Path,
			"but got", err,
		)
	}
}

func TestMaildirDelete(t *testing.T) {
	err := testMaildir.Delete()
	if err != nil {
		t.Error(
			"Tried to delete", testMaildir.Path,
			"but got", err,
		)
	}
}

func TestMaildirIsNotExist(t *testing.T) {
	if !testMaildir.IsNotExist() {
		t.Error(
			"Maildir should not exist but it is",
		)
	}
}
