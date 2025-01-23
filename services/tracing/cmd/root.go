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
			Flags: []cli.Flag{
				&cli.Float64Flag{
					Name:     "days",
					Usage:    "Keep only data since <value> ago. Fractions allowed, a day is specified as 24 hours",
					Required: true,
				},
			},
		},
	},
}
