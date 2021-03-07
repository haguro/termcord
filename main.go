package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"termcord/cmd"
)

func main() {
	//TODO: Move the flag/args parsing part to seperate function.
	var fName, cmdName string
	var cmdArgs []string
	var interactive bool

	//TODO flags
	flag.Parse()

	switch flag.NArg() {
	case 0:
		fName = "termcording"
		cmdName = getShell()
	case 1:
		fName = flag.Arg(0)
		cmdName = getShell()
	default:
		fName = flag.Arg(0)
		cmdName = flag.Arg(1)
		cmdArgs = flag.Args()[2:]
	}

	f, err := os.OpenFile(fName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	c := exec.Command(cmdName, cmdArgs...)

	fmt.Println("Starting recording session. Use CTRL-D to end.")
	defer fmt.Printf("\nRecording session ended. Session saved to %s\n", fName)

	config := cmd.Config{File: f, Cmd: c, Iactive: interactive}

	if err = cmd.Run(config); err != nil {
		log.Fatal(err)
	}
}

func getShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		log.Fatal("shell is not set")
	}
	return shell
}
