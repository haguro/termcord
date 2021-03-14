package main

import (
	"log"
	"termcord/pkg/termcorder"
)

func main() {
	tc, err := termcorder.TermcordingFromFlags()
	if err != nil {
		log.Fatal(err)
	}

	//TODO a better way to handle optional parameters
	if err := tc.Start(nil, nil); err != nil {
		log.Fatal(err)
	}
}
