package nm2md

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaildirCreate(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "testmaildir")
	d := &maildir{path: dir}

	if err := d.create(); err != nil {
		t.Fatalf("create() error: %v", err)
	}

	for _, sub := range maildirSubDirs {
		info, err := os.Stat(filepath.Join(dir, sub))
		if err != nil {
			t.Errorf("subdirectory %q not created: %v", sub, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%q is not a directory", sub)
		}
	}
}

func TestMaildirCreateIdempotent(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "testmaildir")
	d := &maildir{path: dir}

	if err := d.create(); err != nil {
		t.Fatalf("first create() error: %v", err)
	}
	if err := d.create(); err != nil {
		t.Fatalf("second create() error: %v", err)
	}
}

func TestMaildirClear(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "testmaildir")
	d := &maildir{path: dir}

	if err := d.create(); err != nil {
		t.Fatalf("create() error: %v", err)
	}

	sentinel := filepath.Join(dir, "cur", "testfile")
	if err := os.WriteFile(sentinel, []byte("test"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := d.clear(); err != nil {
		t.Fatalf("clear() error: %v", err)
	}

	if _, err := os.Stat(sentinel); !errors.Is(err, fs.ErrNotExist) {
		t.Error("sentinel file should have been removed by clear()")
	}

	for _, sub := range maildirSubDirs {
		if _, err := os.Stat(filepath.Join(dir, sub)); err != nil {
			t.Errorf("subdirectory %q missing after clear: %v", sub, err)
		}
	}
}

func TestMaildirSymlinkFile(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	srcCur := filepath.Join(srcDir, "cur")
	if err := os.Mkdir(srcCur, 0o700); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	srcFile := filepath.Join(srcCur, "testmsg")
	if err := os.WriteFile(srcFile, []byte("From: test"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	dstDir := filepath.Join(t.TempDir(), "dst")
	dst := &maildir{path: dstDir}
	if err := dst.create(); err != nil {
		t.Fatalf("create() error: %v", err)
	}

	if err := dst.symlinkFile(srcFile); err != nil {
		t.Fatalf("symlinkFile() error: %v", err)
	}

	link := filepath.Join(dstDir, "cur", "testmsg")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("Readlink: %v", err)
	}
	if target != srcFile {
		t.Errorf("symlink target = %q, want %q", target, srcFile)
	}
}

func TestMaildirSymlinkFileRejectsDirectory(t *testing.T) {
	t.Parallel()

	srcDir := filepath.Join(t.TempDir(), "cur", "subdir")
	if err := os.MkdirAll(srcDir, 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	dstDir := filepath.Join(t.TempDir(), "dst")
	dst := &maildir{path: dstDir}
	if err := dst.create(); err != nil {
		t.Fatalf("create() error: %v", err)
	}

	err := dst.symlinkFile(srcDir)
	if err == nil {
		t.Fatal("expected error for directory source, got nil")
	}
	if !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "not a regular file")
	}
}

func TestMaildirSymlinkFileInvalidSubDir(t *testing.T) {
	t.Parallel()

	srcFile := filepath.Join(t.TempDir(), "baddir", "testmsg")
	if err := os.MkdirAll(filepath.Dir(srcFile), 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(srcFile, []byte("test"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	dstDir := filepath.Join(t.TempDir(), "dst")
	dst := &maildir{path: dstDir}
	if err := dst.create(); err != nil {
		t.Fatalf("create() error: %v", err)
	}

	err := dst.symlinkFile(srcFile)
	if err == nil {
		t.Fatal("expected error for invalid subdirectory, got nil")
	}
}
