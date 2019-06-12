package progressbar

import (
	"fmt"
	"time"
)

// Char is the character that will be shown as the progress bar.
// RefreshInterval is the time in ms to add a new char to the progress bar.
const (
	Char            string        = "."
	RefreshInterval time.Duration = 100
)

// ProgressBar is defining the struct for itself.
type ProgressBar struct {
	Char            string
	RefreshInterval time.Duration

	finishedChan chan struct{}
}

// NewProgressBar returing a struct for a ProgressBar.
func NewProgressBar(refreshInterval time.Duration, char string) *ProgressBar {
	return &ProgressBar{
		Char:            char,
		RefreshInterval: refreshInterval,

		finishedChan: make(chan struct{}),
	}
}

// Start is starting the execution of a ProgressBar.
func (pb *ProgressBar) Start() {
	go pb.refresh()
}

// StartWithMsg is starting the execution of a ProgressBar with one of the random messages.
func (pb *ProgressBar) StartWithMsg(msg string) {
	fmt.Print(msg)
	pb.Start()
}

// Finish is finishing the execution of the ProgressBar.
func (pb *ProgressBar) Finish() {
	close(pb.finishedChan)
}

// FinishWithMsg is finishing the execution of the ProgressBar with a final message.
func (pb *ProgressBar) FinishWithMsg(msg string) {
	pb.Finish()
	fmt.Print(msg)
}

// printUpdate is printing the defined character to the ProgressBar.
func (pb *ProgressBar) printUpdate() {
	fmt.Print(pb.Char)
}

// refresh is checking for the status of the finished channel in a infinite loop.
func (pb *ProgressBar) refresh() {
	for {
		select {
		case <-pb.finishedChan:
			return
		case <-time.After(time.Millisecond * pb.RefreshInterval):
			pb.printUpdate()
		}
	}
}
