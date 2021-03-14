package termcorder_test

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"termcord/pkg/termcorder"

	"github.com/stretchr/testify/assert"
)

func TestNewTermcording(t *testing.T) {
	t.Parallel()
	cfg := &termcorder.Config{}
	tc, err := termcorder.NewTermcording(cfg)
	assert.IsType(t, &termcorder.Termcording{}, tc)
	assert.NoError(t, err)
}

func TestStart(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	cmd := exec.Command("echo", "success!")
	cfg := &termcorder.Config{Filename: "termcording", Interactive: false}
	tc, err := termcorder.NewTermcording(cfg)
	assert.NoError(t, err)
	tc.Start(cmd, buf)
	want := "success!"
	got := buf.String()
	assert.Contains(t, got, want)
}

func TestTermcordingFromFlags(t *testing.T) {
	t.Parallel()
	t.Run("Test setting flags", func(t *testing.T) {
		os.Args = []string{"./termcord", "-h", "-q", "foo", "bar", "baz"}
		want := &termcorder.Config{
			Filename:    "foo",
			CmdName:     "bar",
			CmdArgs:     []string{"baz"},
			Interactive: false,
			PrintHelp:   true,
			QuietMode:   true,
		}

		got, err := termcorder.TermcordingFromFlags()

		assert.NoError(t, err)
		assert.Equal(t, want, got.Config)
	})

	t.Run("Test setting filename", func(t *testing.T) {
		os.Args = []string{"./termcord", "foo.txt", "bar"}
		want := "foo.txt"
		got, err := termcorder.TermcordingFromFlags()

		assert.NoError(t, err)
		assert.Equal(t, want, got.Config.Filename)
	})

	t.Run("Default to current shell if no command is provided", func(t *testing.T) {
		os.Args = []string{"./termcord"}
		shell, _ := os.LookupEnv("SHELL")
		defer os.Setenv("SHELL", shell)
		shell = "/foo/bar"
		os.Setenv("SHELL", shell)
		got, err := termcorder.TermcordingFromFlags()

		assert.NoError(t, err)
		assert.Equal(t, shell, got.Config.CmdName)
	})

	t.Run("Return an error if shell is not set and no command is provided", func(t *testing.T) {
		os.Args = []string{"./termcord"}
		shell, _ := os.LookupEnv("SHELL")
		defer os.Setenv("SHELL", shell)
		os.Unsetenv("SHELL")
		_, err := termcorder.TermcordingFromFlags()

		assert.Error(t, err)
	})

}
