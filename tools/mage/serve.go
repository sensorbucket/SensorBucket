package magetool

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/magefile/mage/sh"
)

// Serve Starts a service and automatically restarts when a file changes
func Serve(ctx context.Context, service string) error {
	path := fmt.Sprintf("cmd/%s/main.go", service)
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("no such service exists, does the file (%s) exist?", path)
		}
		return err
	}

	fmt.Printf("Watching service: %s\n", service)
	return sh.RunV("nodemon",
		"--watch", "pkg",
		"--watch", "internal",
		"--watch", fmt.Sprintf("service/%s", service),
		"--watch", fmt.Sprintf("cmd/%s", service),
		"--ext", "go",
		"--exec", fmt.Sprintf("go run %s serve || exit 1", path),
		"--signal", "SIGINT",
		"--delay", "0.2",
	)
}
