package main

import (
	"os"

	"github.com/haguro/termcord/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args, os.Stdin, os.Stdout, os.Stderr, cli.FileRecorderSetup))
}
