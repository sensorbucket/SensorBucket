# Getting started

## Project structure
The project is structured as shown below. Each folder contains a `readme.md` with more details.
```sh
.
├── cmd
│   └─ helloworld           # Entry point for binary "helloworld"
├── docs                    # MKDocs documentation (You're reading it now!)
├── internal                # Internal packages used in this codebase
├── pkg                     # Packages that can be used outside this codebase
├── service
│   └── helloworld          # The service "helloworld" 
├── tools           
│   ├── docker-compose.yaml # Local deployment as Docker Compose
│   ├── mage                # Magefile commands
│   ├── openapi             # OpenAPI specification
│   └── seed                # Seeding data for databases and services
└── magefile.go             # Tooling system, somewhat like Makefiles
```

## Commands

This project uses Magefiles to simplify interacting with this project and the system. 

```sh
$> mage
Targets:
  dev:docs       
  dev:logs       Attaches the terminal to the output of the service
  dev:openapi    Serves an OpenAPI UI with automatic reload
  dev:restart    Use to restart the given service
  dev:start      Starts or updates the development environment
  dev:stop       Stops the development environment
  exec           Executes forwards a command to the service CLI
  serve          Starts a service and automatically restarts when a file changes
```

## Running services

### Mage dev:start

Running `mage dev:start` will start the full docker-compose file, which should result in a fully configured Sensorbucket system running on your workstation.

If you make updates to `docker-compose.yml`, call `mage dev:start` again to apply the changes.

### Mage dev:logs

Call `mage dev:logs <service>` to start redirecting the service's output to your terminal. To view all logs substitute `<service>` with `-` (dash character).

### Mage serve

!!! note  
    Services often depend on other services (such as a database or message queue). Using this command will run the service on your machine and **not** in the docker-compose. See `mage dev:start` above for running the docker-compose file.

The command `mage serve <service>` will automatically run: `go run cmd/<service>/main.go serve`. Notice the `serve` subcommand, this can be ignored if there are no other subcommands. Beside running the service, it will also restart the service in case a file changes in one of the following folders:

 - `internal`
 - `pkg`
 - `cmd/<service>`
 - `service/<service>`

## OpenAPI and Documentation

To get a live-reload version of the OpenAPI specification or the Documentation use the following commands:

- `mage dev:openapi`
- `mage dev:docs`

Documentation will be automically build using the Gitlab CI/CD and served on [https://sensorbucket.gitlab.io/sensorbucket](https://sensorbucket.gitlab.io/sensorbucket).