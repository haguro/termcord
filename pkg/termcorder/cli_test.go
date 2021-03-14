package termcorder_test

import (
	"bytes"
	"os"
	"testing"

	"termcord/pkg/termcorder"

	"github.com/stretchr/testify/assert"
)

func TestNewTermcording(t *testing.T) {
	t.Parallel()
	cfg := termcorder.Config{}
	buf, _ := os.Open(os.DevNull)
	tc, closer, err := termcorder.NewTermcording(&cfg, buf)
	defer closer()
	assert.IsType(t, &termcorder.Termcording{}, tc)
	assert.NoError(t, err)
	//TODO more tests
}

func TestStart(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	cfg := &termcorder.Config{Filename: "termcording", CmdName: "echo", CmdArgs: []string{"success!"}, Interactive: false}
	tc, closer, err := termcorder.NewTermcording(cfg, buf)
	assert.NoError(t, err)
	assert.IsType(t, func() error { return nil }, closer)
	tc.Start()
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

		got, closer, err := termcorder.TermcordingFromFlags()

		assert.IsType(t, func() error { return nil }, closer)
		assert.NoError(t, err)
		assert.Equal(t, want, got.Config)
	})

	t.Run("Test setting filename", func(t *testing.T) {
		os.Args = []string{"./termcord", "foo.txt", "bar"}
		want := "foo.txt"
		got, _, err := termcorder.TermcordingFromFlags()

		assert.NoError(t, err)
		assert.Equal(t, want, got.Config.Filename)
	})

	t.Run("Default to current shell if no command is provided", func(t *testing.T) {
		os.Args = []string{"./termcord"}
		shell, _ := os.LookupEnv("SHELL")
		defer os.Setenv("SHELL", shell)
		shell = "/foo/bar"
		os.Setenv("SHELL", shell)
		got, _, err := termcorder.TermcordingFromFlags()

		assert.NoError(t, err)
		assert.Equal(t, shell, got.Config.CmdName)
	})

	t.Run("Return an error if shell is not set and no command is provided", func(t *testing.T) {
		os.Args = []string{"./termcord"}
		shell, _ := os.LookupEnv("SHELL")
		defer os.Setenv("SHELL", shell)
		os.Unsetenv("SHELL")
		_, _, err := termcorder.TermcordingFromFlags()

		assert.Error(t, err)
	})

}
