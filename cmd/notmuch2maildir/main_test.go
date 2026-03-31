package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		wantCode exitCode
		wantOut  string
		wantErr  string
	}{
		{"Help", []string{"--help"}, exitSuccess, "--thread", ""},
		{"Version", []string{"--version"}, exitSuccess, "v1.2.3", ""},
		{"UnknownFlag", []string{"--nonexistent"}, exitUsageError, "", ""},
		{"NoArgs", []string{}, exitUsageError, "", "no search query provided"},
		{"ThreadAndPrompt", []string{"-t", "-p"}, exitUsageError, "", "--thread and --prompt cannot be combined"},
		{"MessageIDWithoutThread", []string{"-m", "abc@example.com", "tag:inbox"}, exitUsageError, "", "--message-id requires --thread"},
		{"ThreadWithPositionalArgs", []string{"-t", "tag:inbox"}, exitUsageError, "", "--thread does not accept positional arguments"},
		{"PromptWithPositionalArgs", []string{"-p", "tag:inbox"}, exitUsageError, "", "--prompt does not accept positional arguments"},
		{"InvalidExecutable", []string{"--notmuch-executable", "/nonexistent/binary/xyz", "tag:inbox"}, exitUsageError, "", "notmuch executable not found"},
		{"ConfigFileNotFound", []string{"--notmuch-config", "/nonexistent/config", "tag:inbox"}, exitUsageError, "", "not accessible"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), &bytes.Buffer{}, &stdout, &stderr, tt.args, "v1.2.3")
			if code != tt.wantCode {
				t.Fatalf("exit code = %d, want %d; stderr: %s", code, tt.wantCode, stderr.String())
			}
			if tt.wantOut != "" && !strings.Contains(stdout.String(), tt.wantOut) {
				t.Fatalf("stdout missing %q; got: %s", tt.wantOut, stdout.String())
			}
			if tt.wantErr != "" && !strings.Contains(stderr.String(), tt.wantErr) {
				t.Fatalf("stderr missing %q; got: %s", tt.wantErr, stderr.String())
			}
		})
	}
}

func TestValidateFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		thread  bool
		prompt  bool
		msgID   string
		args    []string
		wantErr string
	}{
		{
			name:   "ValidSearchArgs",
			thread: false, prompt: false, msgID: "", args: []string{"tag:inbox"},
		},
		{
			name:   "ValidThreadWithMsgID",
			thread: true, prompt: false, msgID: "<abc@host>", args: []string{},
		},
		{
			name:   "ValidThreadNoMsgID",
			thread: true, prompt: false, msgID: "", args: []string{},
		},
		{
			name:   "ValidPromptMode",
			thread: false, prompt: true, msgID: "", args: []string{},
		},
		{
			name:   "ThreadAndPrompt",
			thread: true, prompt: true, msgID: "", args: []string{},
			wantErr: "--thread and --prompt cannot be combined",
		},
		{
			name:   "ThreadWithArgs",
			thread: true, prompt: false, msgID: "", args: []string{"tag:inbox"},
			wantErr: "--thread does not accept positional arguments",
		},
		{
			name:   "MessageIDWithoutThread",
			thread: false, prompt: false, msgID: "<abc@host>", args: []string{"tag:inbox"},
			wantErr: "--message-id requires --thread",
		},
		{
			name:   "PromptWithArgs",
			thread: false, prompt: true, msgID: "", args: []string{"tag:inbox"},
			wantErr: "--prompt does not accept positional arguments",
		},
		{
			name:   "NoQueryNoPromptNoThread",
			thread: false, prompt: false, msgID: "", args: []string{},
			wantErr: "no search query provided",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateFlags(tt.thread, tt.prompt, tt.msgID, tt.args)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestReadMessageIDFromStdin(t *testing.T) {
	t.Parallel()

	validMail := "From: sender@example.com\r\nTo: recipient@example.com\r\nMessage-Id: <abc123@mail.example.com>\r\nSubject: Test\r\n\r\nBody text.\r\n"

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{
			name:  "ValidMailWithMessageID",
			input: validMail,
			want:  "<abc123@mail.example.com>",
		},
		{
			name:    "EmptyInput",
			input:   "",
			wantErr: "cannot parse mail from stdin",
		},
		{
			name:    "InvalidMail",
			input:   "this is not a valid RFC 2822 message\x00",
			wantErr: "cannot parse mail from stdin",
		},
		{
			name:    "MailWithoutMessageIDHeader",
			input:   "From: sender@example.com\r\nSubject: No ID\r\n\r\nBody.\r\n",
			wantErr: "no Message-Id header found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := readMessageIDFromStdin(strings.NewReader(tt.input))
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want to contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("message-id = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExpandHome(t *testing.T) {
	t.Parallel()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "EmptyPath",
			input: "",
			want:  "",
		},
		{
			name:  "TildeOnly",
			input: "~",
			want:  home,
		},
		{
			name:  "TildeWithSubdir",
			input: "~/.config/notmuch",
			want:  filepath.Join(home, ".config", "notmuch"),
		},
		{
			name:  "AbsolutePath",
			input: "/etc/notmuch.conf",
			want:  "/etc/notmuch.conf",
		},
		{
			name:  "RelativePath",
			input: "some/relative/path",
			want:  "some/relative/path",
		},
		{
			name:    "TildeUserRejected",
			input:   "~bob/.config",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := expandHome(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expandHome(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDefaultCacheDir(t *testing.T) {
	dir := defaultCacheDir()
	if !strings.Contains(dir, "notmuch") {
		t.Errorf("defaultCacheDir() = %q, expected to contain %q", dir, "notmuch")
	}
	if !strings.Contains(dir, "search_results") {
		t.Errorf("defaultCacheDir() = %q, expected to contain %q", dir, "search_results")
	}

	t.Run("RespectsXDGCacheHome", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", "/tmp/xdg-test")
		got := defaultCacheDir()
		want := "/tmp/xdg-test/notmuch/search_results"
		if got != want {
			t.Fatalf("defaultCacheDir() = %q, want %q", got, want)
		}
	})
}
