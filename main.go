package main

import (
	"fmt"
	"log"
	"os"
	"termcord/cmd"
)

func main() {
	filename := "termcording"
	command := os.Getenv("SHELL")
	if command == "" {
		log.Fatal("shell is not set")
	}

	fmt.Println("Starting recording session")
	defer fmt.Printf("Recording session ended. Session saved to %q\n", filename)

	err := cmd.Run(filename, command)
	if err != nil {
		log.Fatal(err)
	}
}
