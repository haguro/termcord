package tc

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestNewTermcordingNoOptionsNoError(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("echo", "hello")
	_, err := NewTermcorder(cmd)
	if err != nil {
		t.Fatalf("unexpected error status: %v", err)
	}
}

func TestRawModeReturnsErrorOnSingleCommand(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("echo", "hello")
	tt, _ := NewTermcorder(cmd)
	err := RawMode(tt)
	if err == nil {
		t.Fatalf("unexpected error status: %v", err)
	}
}

func TestWithOutputWriters(t *testing.T) {
	t.Parallel()
}

func TestWithInputWriters(t *testing.T) {
	t.Parallel()
}

func TestRecordSingleCommand(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	cmd := exec.Command("echo", "hello")
	err := Record(cmd, WithOutputWriters(out))
	got := out.String()
	if err != nil {
		t.Fatalf("unexpected error status: %v", err)
	}
	if !strings.HasPrefix(got, "hello") {
		t.Errorf("want %q, got %q", "hello", got)
	}
}
