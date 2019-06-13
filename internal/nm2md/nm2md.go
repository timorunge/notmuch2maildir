package nm2md

import (
	"errors"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/remeh/sizedwaitgroup"
	"github.com/timorunge/notmuch2maildir/internal/maildir"
	"github.com/timorunge/notmuch2maildir/internal/notmuch"
	"github.com/timorunge/notmuch2maildir/internal/progressbar"
	"github.com/timorunge/notmuch2maildir/internal/utils"
)

// swgLimit is the maximum amount of go routines that should be created for the
// creation of symlinks.
const (
	swgLimit int = 15
)

// NM2MD is the base struct for the Notmuch2Maildir
type NM2MD struct {
	ApplicationOptions ApplicationOptions
	SearchQuery        string
}

// ApplicationOptions describes alll available options for the application.
type ApplicationOptions struct {
	Debug             bool
	NotmuchConfigFile string
	NotmuchExecutable string
	OutputDir         string
}

// NewNM2MD is creating a new NM2MD operation.
func NewNM2MD(applicationOptions ApplicationOptions, searchQuery string) *NM2MD {
	return &NM2MD{
		ApplicationOptions: applicationOptions,
		SearchQuery:        searchQuery,
	}
}

// Execute is doing all the magic.
func (nm2md NM2MD) Execute() error {
	if nm2md.SearchQuery == "" {
		return errors.New("No search query defined")
	}

	_, err := os.Stat(nm2md.ApplicationOptions.NotmuchConfigFile)
	if os.IsNotExist(err) {
		utils.PrintDebugErrorMsg(err, nm2md.ApplicationOptions.Debug)
		return errors.New("Notmuch configuration file not found")
	}

	errorChan := make(chan error)
	finishedChan := make(chan bool)
	var wg sync.WaitGroup

	wg.Add(1)
	outputMaildirChan := make(chan *maildir.Maildir, 1)
	go func() {
		defer wg.Done()
		outputMaildir, err := nm2md.OutputMaildir()
		if err != nil {
			errorChan <- err
		} else {
			outputMaildirChan <- outputMaildir
		}
	}()

	wg.Add(1)
	fileListChan := make(chan []string, 1)
	go func() {
		defer wg.Done()
		nm := notmuch.NewNotmuch(
			nm2md.ApplicationOptions.NotmuchExecutable,
			nm2md.ApplicationOptions.NotmuchConfigFile,
			"search", []string{"--output=files", nm2md.SearchQuery})
		fileList, err := nm.Run()
		if err != nil {
			utils.PrintDebugErrorMsg(err, nm2md.ApplicationOptions.Debug)
			errorChan <- errors.New("Notmuch returned a non zero exit status")
		} else if len(fileList) == 0 {
			errorChan <- errors.New("Notmuch was not able to find any mails matching your search query")
		} else {
			fileListChan <- fileList
		}
	}()

	go func() {
		wg.Wait()
		close(finishedChan)
	}()

	r := rand.New(rand.NewSource(time.Now().Unix()))
	pb := progressbar.NewProgressBar(progressbar.RefreshInterval, progressbar.Char)
	go pb.StartWithMsg(ProgressBarMsgs[r.Intn(len(ProgressBarMsgs))])

	select {
	case err := <-errorChan:
		pb.FinishWithMsg("\n")
		return err
	case <-finishedChan:
		fileList, outputMaildir := <-fileListChan, <-outputMaildirChan
		nm2md.Symlink(outputMaildir, fileList)
		pb.FinishWithMsg("\n")
	}
	return nil
}

// OutputMaildir is creating / cleaning the output maildir.
func (nm2md NM2MD) OutputMaildir() (*maildir.Maildir, error) {
	outputDir, err := utils.AbsDir(nm2md.ApplicationOptions.OutputDir)
	if err != nil {
		utils.PrintDebugErrorMsg(err, nm2md.ApplicationOptions.Debug)
		return nil, errors.New("Can not get output directory")
	}
	outputMaildir := maildir.NewMaildir(outputDir, maildir.DefaultPerm)
	if outputMaildir.IsNotExist() {
		err = outputMaildir.Create()
		if err != nil {
			utils.PrintDebugErrorMsg(err, nm2md.ApplicationOptions.Debug)
			return nil, errors.New("Can not create mailbox for the Notmuch search results")
		}
	} else {
		err = outputMaildir.Clear()
		if err != nil {
			utils.PrintDebugErrorMsg(err, nm2md.ApplicationOptions.Debug)
			return nil, errors.New("Can not clear mailbox for the Notmuch search results")
		}
	}
	return outputMaildir, nil
}

// Symlink is creating all symbolic links from a Notmuch search result to a maildir.
func (nm2md NM2MD) Symlink(destMaildir *maildir.Maildir, fileList []string) []error {
	var errs []error
	var swg sizedwaitgroup.SizedWaitGroup = sizedwaitgroup.New(swgLimit)
	for _, file := range fileList {
		swg.Add()
		go func(file string) {
			defer swg.Done()
			err := destMaildir.SymlinkFile(file)
			if err != nil {
				utils.PrintDebugErrorMsg(err, nm2md.ApplicationOptions.Debug)
				errs = append(errs, err)
			}
		}(file)
	}
	swg.Wait()
	if len(errs) > 0 {
		utils.PrintDebugMsg("Some files can not be linked - this usually happens if the maildir and the notmuch index are not synced. Try to run `notmuch new'", nm2md.ApplicationOptions.Debug)
	}
	return errs
}
