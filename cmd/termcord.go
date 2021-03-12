package main

import (
	"log"
	"os"
	"os/exec"
	"termcord/pkg/termcorder"
)

func main() {
	config, err := termcorder.ParseArgs()
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile(config.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	c := exec.Command(config.CmdName, config.CmdArgs...)

	if err := termcorder.Run(c, f, config); err != nil {
		log.Fatal(err)
	}
}
