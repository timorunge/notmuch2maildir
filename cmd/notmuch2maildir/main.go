// Package main provides the notmuch2maildir CLI entry point.
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/mail"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/pflag"

	"github.com/timorunge/notmuch2maildir/internal/nm2md"
)

type exitCode int

const (
	exitSuccess exitCode = iota
	exitError
	exitUsageError
)

const maxStdinRead = 10 << 20 // 10 MB

var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Stdin, os.Stdout, os.Stderr, os.Args[1:], version)
	stop()
	os.Exit(int(code))
}

func run(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer, args []string, version string) exitCode {
	fs := pflag.NewFlagSet("notmuch2maildir", pflag.ContinueOnError)
	fs.SortFlags = false
	fs.SetOutput(stderr)
	fs.Usage = func() { printUsage(stdout, stderr, fs) }

	// Mode flags.
	thread := fs.BoolP("thread", "t", false, "Thread mode: reconstruct a full mail thread")
	prompt := fs.BoolP("prompt", "p", false, "Open a prompt to enter the search query")

	// Thread options.
	msgID := fs.StringP("message-id", "m", "", "Message-ID for thread reconstruction")

	// Configuration flags.
	configFile := fs.String("notmuch-config", "", "notmuch configuration file (default: notmuch's own resolution)")
	outputDir := fs.String("output-dir", defaultCacheDir(), "Output directory for search results")

	// Advanced flags.
	executable := fs.String("notmuch-executable", "notmuch", "Path to notmuch binary")

	// Hidden flags.
	debug := fs.Bool("debug", false, "Print debugging output")
	_ = fs.MarkHidden("debug")

	// Informational flags.
	help := fs.BoolP("help", "h", false, "Show this help message")
	showVersion := fs.Bool("version", false, "Show the version of notmuch2maildir")

	if err := fs.Parse(args); err != nil {
		return exitUsageError
	}

	if *help {
		printUsage(stdout, stderr, fs)
		return exitSuccess
	}
	if *showVersion {
		_, _ = fmt.Fprintf(stdout, "notmuch2maildir %s\n", version)
		return exitSuccess
	}

	if err := validateFlags(*thread, *prompt, *msgID, fs.Args()); err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return exitUsageError
	}

	nmOpts, err := buildOptions(*configFile, *outputDir, *executable, *debug, stderr)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return exitUsageError
	}

	if *thread {
		err = runThread(ctx, stdin, nmOpts, *msgID)
	} else {
		err = runSearch(ctx, stdin, stdout, nmOpts, *prompt, fs.Args())
	}
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return exitError
	}

	return exitSuccess
}

func runSearch(ctx context.Context, stdin io.Reader, stdout io.Writer, opts nm2md.Options, prompt bool, args []string) error {
	var query string
	if prompt {
		_, _ = fmt.Fprintln(stdout, "Search query:")
		input, err := bufio.NewReader(stdin).ReadString('\n')
		if err != nil {
			return errors.New("cannot read search query from stdin")
		}
		query = strings.TrimSpace(input)
		if query == "" {
			return errors.New("empty search query")
		}
	} else {
		query = strings.Join(args, " ")
	}
	return nm2md.Search(ctx, opts, query)
}

func runThread(ctx context.Context, stdin io.Reader, opts nm2md.Options, msgID string) error {
	if msgID == "" {
		if f, ok := stdin.(*os.File); ok && isTerminal(f) {
			return errors.New("no message-id provided (use -m or pipe a mail to stdin)")
		}
		id, err := readMessageIDFromStdin(stdin)
		if err != nil {
			return err
		}
		msgID = id
	}
	return nm2md.Thread(ctx, opts, msgID)
}

func readMessageIDFromStdin(stdin io.Reader) (string, error) {
	msg, err := mail.ReadMessage(io.LimitReader(stdin, maxStdinRead))
	if err != nil {
		return "", errors.New("cannot parse mail from stdin")
	}
	id := msg.Header.Get("Message-Id")
	if id == "" {
		return "", errors.New("no Message-Id header found in piped mail")
	}
	return id, nil
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func validateFlags(thread, prompt bool, msgID string, args []string) error {
	if thread && prompt {
		return errors.New("--thread and --prompt cannot be combined")
	}
	if thread && len(args) > 0 {
		return errors.New("--thread does not accept positional arguments")
	}
	if msgID != "" && !thread {
		return errors.New("--message-id requires --thread")
	}
	if prompt && len(args) > 0 {
		return errors.New("--prompt does not accept positional arguments")
	}
	if !thread && !prompt && len(args) == 0 {
		return errors.New("no search query provided")
	}
	return nil
}

func buildOptions(configFile, outputDir, executable string, debug bool, stderr io.Writer) (nm2md.Options, error) {
	cfgPath, err := expandHome(configFile)
	if err != nil {
		return nm2md.Options{}, fmt.Errorf("cannot resolve notmuch config path: %w", err)
	}
	if cfgPath != "" {
		if _, statErr := os.Stat(cfgPath); statErr != nil {
			return nm2md.Options{}, fmt.Errorf("notmuch configuration file not accessible: %w", statErr)
		}
	}

	resolved, err := exec.LookPath(executable)
	if err != nil {
		return nm2md.Options{}, fmt.Errorf("notmuch executable not found: %w", err)
	}

	outPath, err := expandHome(outputDir)
	if err != nil {
		return nm2md.Options{}, fmt.Errorf("cannot resolve output directory: %w", err)
	}

	var logger *slog.Logger
	if debug {
		logger = slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else {
		logger = slog.New(slog.DiscardHandler)
	}

	return nm2md.Options{
		Config: nm2md.Config{
			NotmuchExecutable: resolved,
			NotmuchConfigFile: cfgPath,
			OutputDir:         outPath,
		},
		Output: stderr,
		Logger: logger,
	}, nil
}

func defaultCacheDir() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "notmuch", "search_results")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		// Tilde fallback; buildOptions passes this through expandHome.
		return "~/.cache/notmuch/search_results"
	}
	return filepath.Join(home, ".cache", "notmuch", "search_results")
}

func expandHome(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	rest, ok := strings.CutPrefix(path, "~")
	if !ok {
		return filepath.Clean(path), nil
	}
	if rest != "" && rest[0] != '/' {
		return "", fmt.Errorf("~user syntax is not supported: %s", path)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Clean(filepath.Join(home, rest)), nil
}

func printUsage(w, orig io.Writer, fs *pflag.FlagSet) {
	fs.SetOutput(w)
	_, _ = fmt.Fprintf(w, "notmuch2maildir - Search mail and reconstruct threads using notmuch\n\nUsage:\n  notmuch2maildir [OPTIONS] QUERY\n  notmuch2maildir -p\n  notmuch2maildir -t -m <message-id>\n  notmuch2maildir -t < email.eml\n\nOptions:\n")
	fs.PrintDefaults()
	fs.SetOutput(orig)
}
