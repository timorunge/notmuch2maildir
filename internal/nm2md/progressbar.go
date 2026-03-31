// Terminal progress indicator.

package nm2md

import (
	"fmt"
	"io"
	"math/rand/v2"
	"sync"
	"time"
)

var progressMessages = [...]string{
	"@todo Insert witty loading message",
	"Adjusting flux compensator",
	"Computing the secret to life, the universe, and everything",
	"Counting backwards from infinity",
	"Dividing by zero",
	"Following the white rabbit",
	"Laughing at your mails - I mean, loading",
	"Mining some bitcoins",
	"Oh shit, you were waiting for me to do something? Oh okay, well then",
	"Testing your patience",
	"Work, work",
}

type progressBar struct {
	w       io.Writer
	once    sync.Once
	done    chan struct{}
	stopped chan struct{}
}

func startProgress(w io.Writer) *progressBar {
	pb := &progressBar{w: w, done: make(chan struct{}), stopped: make(chan struct{})}
	_, _ = fmt.Fprint(w, progressMessages[rand.IntN(len(progressMessages))]) //nolint:gosec // cosmetic message selection, not security-sensitive
	go func() {
		defer close(pb.stopped)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-pb.done:
				return
			case <-ticker.C:
				_, _ = fmt.Fprint(w, ".")
			}
		}
	}()
	return pb
}

func (pb *progressBar) stop() {
	pb.once.Do(func() {
		close(pb.done)
		<-pb.stopped
		_, _ = fmt.Fprintln(pb.w)
	})
}
