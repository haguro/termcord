package tc_test

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/haguro/termcord/pkg/tc"
)

func TestNewTermcordingNoOptionsNoError(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("echo", "hello")
	_, err := tc.NewTermcorder(cmd)
	if err != nil {
		t.Fatalf("unexpected error status: %v", err)
	}
}

func TestRecordSingleCommand(t *testing.T) {
	t.Parallel()
	want := "hello"
	out := &bytes.Buffer{}
	cmd := exec.Command("echo", want)
	err := tc.Record(cmd, tc.WithOutputWriters(out))
	if err != nil {
		t.Fatalf("unexpected error status: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte(want)) {
		t.Errorf("output %q does not contain %q", out.String(), want)
	}
}

func TestNewWithMultipleReadersAndWriters(t *testing.T) {
	t.Parallel()
	want1, want2, want3 := "aaa", "bbb", "ccc"
	inRdr1 := bytes.NewBuffer([]byte(want1))
	inRdr2 := bytes.NewBuffer([]byte(want2))
	inWrtr := &bytes.Buffer{}
	out := &bytes.Buffer{}
	cmd := exec.Command("echo", want3)
	tt, err := tc.NewTermcorder(
		cmd,
		tc.WithInputReaders(inRdr1, inRdr2),
		tc.WithInputWriters(inWrtr),
		tc.WithOutputWriters(out),
	)
	if err != nil {
		t.Fatalf("unexpected error status: %v", err)
	}
	tt.Start()
	if !bytes.Contains(inWrtr.Bytes(), []byte(want1)) ||
		!bytes.Contains(inWrtr.Bytes(), []byte(want2)) {
		t.Errorf("Input writer %q does not contain %q or %q", inWrtr.String(), want1, want2)
	}
	if !bytes.Contains(out.Bytes(), []byte(want1)) ||
		!bytes.Contains(out.Bytes(), []byte(want2)) ||
		!bytes.Contains(out.Bytes(), []byte(want3)) {
		t.Errorf("Output writer %q does not contain %q, %q or %q", out.String(), want1, want2, want3)
	}
}
