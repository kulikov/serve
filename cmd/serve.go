package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/kulikov/serve/commands"
	"github.com/kulikov/serve/gocd"
	"github.com/kulikov/serve/manifest"
	"github.com/kulikov/serve/manifest/alerts"
)

var manifestPlugins = []manifest.Plugin{
	gocd.DeployPlugin{},
	alerts.GraphiteAlertPlugin{},
	alerts.ElasticAlertPlugin{},
}

func main() {
	app := cli.NewApp()
	app.Name = "serve"
	app.Usage = "Automate your infrastructure!"

	app.Commands = []cli.Command{
		commands.ServiceCommand(),
		commands.ConsulCommand(),
		commands.WebhookServerCommand(manifestPlugins),
	}

	app.Run(os.Args)
}
