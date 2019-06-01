package maildir

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultPerm describes the default permissions which should be used when
// creating a Maildir.
const DefaultPerm os.FileMode = 0700

// maildirDirectories describes the directories inside a Maildir.
var maildirDirectories = [3]string{"cur", "new", "tmp"}

// A Maildir represents the directory structure for the Maildir format.
type Maildir struct {
	Path string
	Perm os.FileMode
}

// NewMaildir is returning the struct for a new mail directory.
func NewMaildir(path string, perm os.FileMode) *Maildir {
	return &Maildir{
		Path: path,
		Perm: perm,
	}
}

// Clear removes all maildir (sub)directories from a Maildir structure.
func (d Maildir) Clear() error {
	for _, dir := range maildirDirectories {
		err := os.RemoveAll(filepath.Join(string(d.Path), dir))
		if err != nil {
			return err
		}
	}
	err := d.Create()
	if err != nil {
		return err
	}
	return nil
}

// Create creates the directory structure for a Maildir.
func (d Maildir) Create() error {
	err := os.Mkdir(string(d.Path), d.Perm)
	if err != nil && os.IsNotExist(err) {
		return err
	}
	for _, dir := range maildirDirectories {
		err = os.Mkdir(filepath.Join(string(d.Path), dir), d.Perm)
		if err != nil && os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// Delete deletes the entire Maildir.
func (d Maildir) Delete() error {
	return os.RemoveAll(d.Path)
}

// IsExist checks if the directory is existing.
func (d Maildir) IsExist() bool {
	_, err := os.Stat(d.Path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// IsNotExist checks if the directory is not existing.
func (d Maildir) IsNotExist() bool {
	if d.IsExist() {
		return false
	}
	return true
}

// SymlinkFile is creating a symlink from a single message to the Maildir.
func (d Maildir) SymlinkFile(srcFile string) error {
	srcFileInfo, err := os.Stat(srcFile)
	if err != nil {
		return err
	}
	srcFileDir := filepath.Dir(srcFile)
	destFile := fmt.Sprintf("%s/%s/%s", d.Path, filepath.Base(srcFileDir), srcFileInfo.Name())
	err = os.Symlink(srcFile, destFile)
	if err != nil {
		return err
	}
	return nil
}
