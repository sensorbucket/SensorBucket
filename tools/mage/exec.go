package magetool

import (
	"fmt"
	"strings"
)

// Exec Executes forwards a command to the service CLI
func Exec(service, command string) error {
	compose, err := WithCompose(true)
	if err != nil {
		return err
	}

	composeArgs := []string{
		"exec", service,
		"go", "run", fmt.Sprintf("cmd/%s/main.go", service),
	}
	composeArgs = append(composeArgs, strings.Split(command, " ")...)

	fmt.Println("Executing command:", strings.Join(composeArgs, " "))

	return compose(composeArgs...)
}
