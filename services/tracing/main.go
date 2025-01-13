package main

import (
	"log"
	"os"

	"sensorbucket.nl/sensorbucket/services/tracing/cmd"
)

func main() {
	if err := cmd.App.Run(os.Args); err != nil {
		log.Fatalf("Error: %s\n", err.Error())
	}
}
