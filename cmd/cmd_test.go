package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	t.Parallel()
	c := exec.Command("echo", "success!")
	buf := &bytes.Buffer{}
	cfg := &Config{Filename: "termcording", Interactive: false}
	Run(c, buf, cfg)
	want := "success!"
	got := buf.String()
	assert.Contains(t, got, want)
}

func TestParseArgs(t *testing.T) {
	t.Parallel()
	t.Run("Test setting flags", func(t *testing.T) {
		os.Args = []string{"./termcord", "-h", "-q", "foo", "bar", "baz"}
		want := &Config{
			Filename:    "foo",
			CmdName:     "bar",
			CmdArgs:     []string{"baz"},
			Interactive: false,
			PrintHelp:   true,
			QuietMode:   true,
		}
		got, err := ParseArgs()

		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("Test setting filename", func(t *testing.T) {
		os.Args = []string{"./termcord", "foo.txt", "bar"}
		want := "foo.txt"
		got, err := ParseArgs()

		assert.NoError(t, err)
		assert.Equal(t, want, got.Filename)
	})

	t.Run("Default to current shell if no command is provided", func(t *testing.T) {
		os.Args = []string{"./termcord"}
		shell, _ := os.LookupEnv("SHELL")
		defer os.Setenv("SHELL", shell)
		shell = "/foo/bar"
		os.Setenv("SHELL", shell)
		got, err := ParseArgs()

		assert.NoError(t, err)
		assert.Equal(t, shell, got.CmdName)
	})

	t.Run("Return an error if shell is not set and no command is provided", func(t *testing.T) {
		os.Args = []string{"./termcord"}
		shell, _ := os.LookupEnv("SHELL")
		defer os.Setenv("SHELL", shell)
		os.Unsetenv("SHELL")
		_, err := ParseArgs()

		assert.Error(t, err)
	})

}

func TestPtmxFromCmd(t *testing.T) {
	t.Parallel()
	//TODO: We're essentially testing the pty package here. Do we need this?
	c := exec.Command("echo", "success!")
	ptmx, _, err := ptmxFromCmd(c, false)
	assert.IsType(t, new(os.File), ptmx)
	assert.NoError(t, err)
}
