package main

import (
	"fmt"
	"os"

	"sensorbucket.nl/service/helloworld"
)

func main() {
	if err := helloworld.Serve(":3000"); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v", err)
	}
}
