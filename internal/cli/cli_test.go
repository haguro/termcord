package cli_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/haguro/termcord/internal/cli"
	"github.com/stretchr/testify/assert"
)

func TestRunPrintHelp(t *testing.T) {
	t.Parallel()
	args := []string{"termcord", "-h"}
	in, out, errOut := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	r := cli.Run(args, in, out, errOut)
	if r != 0 {
		t.Errorf("expected to return 0, returned %d", r)
	}
	got := errOut.String()
	assert.Contains(t, got, "termcord is", "help is expected to contain an intro line")
	assert.Contains(t, got, "Usage:", "help is expected to contain a usage statement")
	assert.Contains(t, got, "Options:", "help is expected to contain options descriptions")
}

func TestRunRecordCommandOutput(t *testing.T) {
	t.Parallel()
	args := []string{"termcord", "-i", "echo", "hello"}
	in, out, errOut := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	r := cli.Run(args, in, out, errOut)
	if r != 0 {
		t.Errorf("expected to return 0, returned %d", r)
	}
	got := out.String()
	assert.Contains(t, got, "hello", `expected output to contain "hello"`, got)
	assert.Contains(t, got, "Starting recording session", "expected output to contain a recording start prompt")
	assert.Contains(t, got, "Recording session ended", "expected output to contain a recording end prompt")
}

func debugOutBuffers(out, errOut *bytes.Buffer) {
	fmt.Printf("out: %q\n", out.String())
	fmt.Printf("err: %q\n", errOut.String())
}
