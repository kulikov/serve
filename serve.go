package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/kulikov/serve/consul"
	"github.com/kulikov/serve/deploy"
	"github.com/kulikov/serve/github"
	"github.com/kulikov/serve/marathon"
)

func main() {
	app := cli.NewApp()
	app.Name = "serve"
	app.Version = "0.2"
	app.Usage = "Automate your infrastructure!"

	app.Commands = []cli.Command{
		consul.ConsulCommand(),
		marathon.MarathonCommand(),
		deploy.BuildCommand(),
		deploy.DeployCommand(),
		deploy.ReleaseCommand(),
		github.WebhookServerCommand(),
	}

	app.Run(os.Args)
}
