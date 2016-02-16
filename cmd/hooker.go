package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"

	"../webhook"
)

func main() {
	app := cli.NewApp()
	app.Name = "hooker"
	app.Usage = "Listen web hooks and call action"

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "port",
			Value: 9090,
		},
		cli.StringFlag{
			Name:  "handlers",
			Usage: "Path to dir with hook handlers scripts",
		},
	}

	app.Action = func(c *cli.Context) {
		if !c.IsSet("handlers") {
			log.Fatalf("--handlers is required!")
		}

		webhook.StartWebHookServer(c.Int("port"), c.String("handlers"))
	}

	app.Run(os.Args)
}
