package main

import (
	"log"
	"os"
	"os/exec"
	"termcord/cmd"
)

func main() {
	config := cmd.ParseArgs()

	f, err := os.OpenFile(config.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	c := exec.Command(config.CmdName, config.CmdArgs...)

	if err := cmd.Run(c, f, config); err != nil {
		log.Fatal(err)
	}
}
