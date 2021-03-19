package main

import (
	"log"

	"github.com/haguro/termcord/pkg/termcorder"
)

func main() {
	tc, err := termcorder.TermcordingFromFlags()
	if err != nil {
		log.Fatal(err)
	}

	if err := tc.Start(); err != nil {
		log.Fatal(err)
	}
}
