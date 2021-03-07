package cmd_test

import (
	"bytes"
	"os/exec"
	"termcord/cmd"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := cmd.Config{
		File: buf,
		Cmd:  exec.Command("echo", "success"),
	}

	cmd.Run(cfg)
	got := buf.String()
	assert.Contains(t, got, "success")
}
