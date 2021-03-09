package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	c := exec.Command("echo", "success!")
	buf := &bytes.Buffer{}
	cfg := Config{}
	Run(c, buf, cfg)
	want := "success!"
	got := buf.String()
	assert.Contains(t, got, want)
}

func TestParseArgs(t *testing.T) {
	//TODO
}

func TestPtmxFromCmd(t *testing.T) {
	//TODO: We're essentially testing the pty package here. Do we need this?
	c := exec.Command("echo", "success!")
	ptmx, fn, err := ptmxFromCmd(c, false)
	assert.Empty(t, fn)
	assert.IsType(t, new(os.File), ptmx)
	assert.NoError(t, err)
}
