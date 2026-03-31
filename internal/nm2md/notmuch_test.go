package nm2md

import (
	"strings"
	"testing"
)

func TestNotmuchArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configFile string
		command    string
		args       []string
		want       string
	}{
		{
			name:       "with config and args",
			configFile: "~/.notmuch-config",
			command:    "search",
			args:       []string{"--output=files", "Foo"},
			want:       "--config=~/.notmuch-config search -- --output=files Foo",
		},
		{
			name:       "custom config path",
			configFile: "/home/foobar/.config/notmuch/foobar-account",
			command:    "search",
			args:       []string{"--output=files", "FooBar"},
			want:       "--config=/home/foobar/.config/notmuch/foobar-account search -- --output=files FooBar",
		},
		{
			name:       "thread search",
			configFile: "~/.notmuch-config",
			command:    "search",
			args:       []string{"--output=threads", "id:1234"},
			want:       "--config=~/.notmuch-config search -- --output=threads id:1234",
		},
		{
			name:       "empty config",
			configFile: "",
			command:    "search",
			args:       []string{"--output=files", "test"},
			want:       "search -- --output=files test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := strings.Join(notmuchArgs(tt.configFile, tt.command, tt.args), " ")
			if got != tt.want {
				t.Errorf("notmuchArgs() = %q, want %q", got, tt.want)
			}
		})
	}
}
