// Package nm2md implements notmuch search-to-maildir conversion.
package nm2md

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

const symlinkWorkers = 15

// Config holds notmuch and output directory settings.
type Config struct {
	NotmuchExecutable string
	NotmuchConfigFile string
	OutputDir         string
}

// Options configures a search or thread operation.
type Options struct {
	Config
	Output io.Writer
	Logger *slog.Logger
}

func (o Options) logger() *slog.Logger {
	if o.Logger != nil {
		return o.Logger
	}
	return slog.New(slog.DiscardHandler)
}

func (o Options) output() io.Writer {
	if o.Output != nil {
		return o.Output
	}
	return io.Discard
}

// Search runs a notmuch search and symlinks matching messages into the output maildir.
func Search(ctx context.Context, opts Options, query string) error {
	if query == "" {
		return errors.New("no search query defined")
	}
	logger := opts.logger()

	pb := startProgress(opts.output())
	defer pb.stop()

	files, err := runNotmuch(ctx, opts.NotmuchExecutable, opts.NotmuchConfigFile,
		"search", []string{"--output=files", query})
	if err != nil {
		logger.Debug("notmuch search failed", "error", err)
		return fmt.Errorf("notmuch returned a non-zero exit status: %w", err)
	}
	if len(files) == 0 {
		return errors.New("notmuch was not able to find any mails matching your search query")
	}

	dir, err := prepareMaildir(opts.OutputDir)
	if err != nil {
		logger.Debug("maildir preparation failed", "error", err)
		return err
	}

	return symlinkAll(logger, dir, files)
}

// Thread resolves a message-id to its thread and symlinks all thread messages.
func Thread(ctx context.Context, opts Options, msgID string) error {
	if msgID == "" {
		return errors.New("no message-id defined")
	}
	logger := opts.logger()

	plainID := strings.Trim(msgID, "<>")
	threadIDs, err := runNotmuch(ctx, opts.NotmuchExecutable, opts.NotmuchConfigFile,
		"search", []string{"--output=threads", "id:" + plainID})
	if err != nil {
		logger.Debug("notmuch thread lookup failed", "error", err)
		return fmt.Errorf("notmuch returned a non-zero exit status: %w", err)
	}
	if len(threadIDs) == 0 {
		return errors.New("notmuch was not able to find the mail thread")
	}
	if len(threadIDs) > 1 {
		logger.Debug("multiple threads found for message-id, using first", "count", len(threadIDs), "msgID", msgID)
	}

	return Search(ctx, opts, threadIDs[0])
}

func prepareMaildir(outputDir string) (*maildir, error) {
	dir := &maildir{path: outputDir}
	_, statErr := os.Stat(outputDir)
	switch {
	case errors.Is(statErr, fs.ErrNotExist):
		if err := dir.create(); err != nil {
			return nil, fmt.Errorf("cannot create mailbox: %w", err)
		}
	case statErr != nil:
		return nil, fmt.Errorf("cannot access output directory: %w", statErr)
	default:
		if err := dir.clear(); err != nil {
			return nil, fmt.Errorf("cannot clear mailbox: %w", err)
		}
	}
	return dir, nil
}

func symlinkAll(logger *slog.Logger, dir *maildir, files []string) error {
	sem := make(chan struct{}, symlinkWorkers)
	var wg sync.WaitGroup
	var failures atomic.Int64
	for _, file := range files {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := dir.symlinkFile(file); err != nil {
				failures.Add(1)
				logger.Debug("symlink failed", "file", file, "error", err)
			}
		})
	}
	wg.Wait()
	if f := failures.Load(); f == int64(len(files)) {
		return fmt.Errorf("all %d symlink operations failed", len(files))
	}
	return nil
}
