package main

import (
	"log"
	"termcord/pkg/termcorder"
)

func main() {
	tc, closer, err := termcorder.TermcordingFromFlags()
	if err != nil {
		log.Fatal(err)
	}
	defer closer()

	if err := tc.Start(); err != nil {
		log.Fatal(err)
	}
}
