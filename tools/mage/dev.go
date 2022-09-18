package magetool

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Dev mg.Namespace

func RunVCmd(cmd string, args ...string) func(args ...string) error {
	return func(args2 ...string) error {
		return sh.RunV(cmd, append(args, args2...)...)
	}
}

func WithCompose(output bool) (func(args ...string) error, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if output {
		return RunVCmd("docker-compose", "-f", fmt.Sprintf("%s/docker-compose.yaml", wd)), nil
	}
	return sh.RunCmd("docker-compose", "-f", fmt.Sprintf("%s/docker-compose.yaml", wd)), nil
}

// Start Starts or updates the development environment
func (Dev) Start() error {
	compose, err := WithCompose(false)
	if err != nil {
		return err
	}

	fmt.Println("Starting development environment...")
	err = compose("up", "-d")
	if err != nil {
		fmt.Println("Failed to start development environment")
		return err
	}
	fmt.Println("Development environment running")
	return nil
}

// Stop Stops the development environment
func (Dev) Stop() error {
	compose, err := WithCompose(false)
	if err != nil {
		return err
	}

	fmt.Println("Stopping development environment...")
	err = compose("stop")
	if err != nil {
		fmt.Println("Failed to stop development environment")
		return err
	}
	fmt.Println("Development environment stopped")
	return nil
}

// Restart Use to restart the given service
func (Dev) Restart(ctx context.Context, service string) error {
	compose, err := WithCompose(true)
	if err != nil {
		return err
	}

	return compose("restart", service)
}

// Logs Attaches the terminal to the output of the service, use '-' to show all logs
func (Dev) Logs(ctx context.Context, service string) error {
	compose, err := WithCompose(true)
	if err != nil {
		return err
	}

	if service == "-" {
		return compose("logs", "-f")
	}
	return compose("logs", "-f", service)
}

// Openapi Serves an OpenAPI UI with automatic reload
func (Dev) Openapi() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not start openapi docs: %w", err)
	}

	fmt.Println("Starting OpenAPI ui with automatic reload")
	return sh.RunV("npx", "--yes", "swagger-ui-watcher", "-p", "8001", path.Join(wd, "tools/openapi/ref/api.yaml"))
}

func (Dev) Docs() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not start docs: %w", err)
	}

	return sh.RunV("docker", "run", "--rm", "-it", "-p", "8000:8000", "-v", fmt.Sprintf("%s:/docs", wd), "squidfunk/mkdocs-material")
}

func (Dev) Mongo() error {
	return sh.RunV("docker", "run", "--rm", "-it", "-p", "8002:8081",
		"-e", "ME_CONFIG_MONGODB_ADMINUSERNAME=admin",
		"-e", "ME_CONFIG_MONGODB_ADMINPASSWORD=admin",
		"-e", "ME_CONFIG_MONGODB_URL=mongodb://root:root@assetdb:27017/",
		"--network", "sensorbucket_network",
		"mongo-express")
}
