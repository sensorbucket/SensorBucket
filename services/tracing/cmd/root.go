package cmd

import "github.com/urfave/cli/v2"

var App = &cli.App{
	Action: cmdServe,
	Commands: []*cli.Command{
		{
			Name:   "serve",
			Action: cmdServe,
		},
		{
			Name:   "cleanup",
			Action: cmdCleanDatabase,
		},
	},
}
