package termcorder_test

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/haguro/termcord/pkg/termcorder"

	"github.com/stretchr/testify/assert"
)

func TestNewTermcording(t *testing.T) {
	t.Parallel()
	cfg := &termcorder.Config{}
	var tc *termcorder.Termcording
	tc, err := termcorder.NewTermcording(cfg)
	assert.NoError(t, err)
	assert.Equal(t, tc.Config, cfg)
}

func TestFromFlagsSettingFlags(t *testing.T) {
	args := []string{"./termcord", "-h", "-q", "-k", "foo.txt", "bar", "baz"}
	want := &termcorder.Config{
		Filename:      "foo.txt",
		CmdName:       "bar",
		CmdArgs:       []string{"baz"},
		Interactive:   false,
		PrintHelp:     true,
		QuietMode:     true,
		LogKeystrokes: true,
	}

	got, err := termcorder.FromFlags(args)

	assert.NoError(t, err)
	assert.Equal(t, want, got.Config)
}
func TestFromFlagsSettingFilename(t *testing.T) {
	args := []string{"./termcord", "foo.txt"}
	want := "foo.txt"
	shell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", shell)
	shell = "/foo/bar"
	os.Setenv("SHELL", shell)
	got, err := termcorder.FromFlags(args)

	assert.NoError(t, err)
	assert.Equal(t, want, got.Config.Filename)
}

func TestFromFlagsDafaultToShell(t *testing.T) {
	args := []string{"./termcord"}
	shell, _ := os.LookupEnv("SHELL")
	defer os.Setenv("SHELL", shell)
	shell = "/foo/bar"
	os.Setenv("SHELL", shell)
	got, err := termcorder.FromFlags(args)

	assert.NoError(t, err)
	assert.Equal(t, shell, got.Config.CmdName)
}

func TestFromFlagsDefaultToShellWithFilename(t *testing.T) {
	args := []string{"./termcord", "foo.txt"}
	shell, _ := os.LookupEnv("SHELL")
	defer os.Setenv("SHELL", shell)
	shell = "/foo/bar"
	os.Setenv("SHELL", shell)
	got, err := termcorder.FromFlags(args)

	assert.NoError(t, err)
	assert.Equal(t, shell, got.Config.CmdName)
}

func TestFromFlagsErrorIfNoShellAndNoCommand(t *testing.T) {
	args := []string{"./termcord"}
	shell, _ := os.LookupEnv("SHELL")
	defer os.Setenv("SHELL", shell)
	os.Unsetenv("SHELL")
	_, err := termcorder.FromFlags(args)

	assert.Error(t, err)
}

func TestStartRecordingCommand(t *testing.T) {
	buf := &bytes.Buffer{}
	cmd := exec.Command("echo", "success!")
	cfg := &termcorder.Config{Filename: "termcording", Interactive: false}
	tc, err := termcorder.NewTermcording(cfg, termcorder.Cmd(cmd), termcorder.Output(buf))
	assert.NoError(t, err)
	tc.Start()

	want := "success!"
	got := buf.String()
	assert.Contains(t, got, want)
}

func TestStartPrintingHelp(t *testing.T) {
	t.Parallel()
	args := []string{"./termcord", "-h"}
	buf := &bytes.Buffer{}
	tc, err := termcorder.FromFlags(args, termcorder.Output(buf))
	assert.NoError(t, err)
	tc.Start()

	want := "Usage:"
	got := buf.String()
	assert.Contains(t, got, want)
}
