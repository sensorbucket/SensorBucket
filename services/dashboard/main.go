package main

import (
	"fmt"
	"os"

	"sensorbucket.nl/sensorbucket/services/dashboard/dashboard"
)

func main() {
	if err := dashboard.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}
