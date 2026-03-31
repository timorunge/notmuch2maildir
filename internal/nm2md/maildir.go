// Maildir directory structure management.

package nm2md

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const defaultMaildirPerm os.FileMode = 0o700

var maildirSubDirs = [...]string{"cur", "new", "tmp"}

type maildir struct {
	path string
}

func (d *maildir) create() error {
	if err := os.MkdirAll(d.path, defaultMaildirPerm); err != nil {
		return fmt.Errorf("create maildir %s: %w", d.path, err)
	}
	for _, sub := range maildirSubDirs {
		if err := os.Mkdir(filepath.Join(d.path, sub), defaultMaildirPerm); err != nil && !errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("create subdirectory %s: %w", sub, err)
		}
	}
	return nil
}

func (d *maildir) clear() error {
	for _, sub := range maildirSubDirs {
		if err := os.RemoveAll(filepath.Join(d.path, sub)); err != nil {
			return fmt.Errorf("remove subdirectory %s: %w", sub, err)
		}
	}
	return d.create()
}

func (d *maildir) symlinkFile(srcFile string) error {
	info, err := os.Stat(srcFile)
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("source is not a regular file: %s", srcFile)
	}
	subDir := filepath.Base(filepath.Dir(srcFile))
	if subDir != "cur" && subDir != "new" && subDir != "tmp" {
		return fmt.Errorf("unexpected maildir subdirectory %q in %q", subDir, srcFile)
	}
	return os.Symlink(srcFile, filepath.Join(d.path, subDir, info.Name()))
}
