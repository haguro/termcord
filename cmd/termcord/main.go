package main

import (
	"log"
	"os"

	"github.com/haguro/termcord/pkg/termcorder"
)

func main() {
	tc, err := termcorder.FromFlags(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	if err := tc.Start(); err != nil {
		log.Fatal(err)
	}
}
