package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"

	"../gocd"
	"../manifest"
	"../manifest/alerts"
)

var manifestPlugins = []manifest.Plugin{
	&gocd.DeployPlugin{},
	&alerts.GraphiteAlertPlugin{},
	&alerts.ElasticAlertPlugin{},
}

func main() {
	app := cli.NewApp()
	app.Name = "serve"
	app.Usage = "Automate your infrastructure!"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "config.yml",
			Usage: "Path to config.yaml file",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "github-hook",
			Usage: "Handle github hook event and check manifest changes",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "payload"},
			},
			Action: func(c *cli.Context) {
				conf, err := manifest.InitConfig(c.String("config"))
				if err != nil {
					log.Panic(err)
				}

				if err := manifest.HandleGithubChanges(conf, manifestPlugins, c.String("payload")); err != nil {
					log.Panicln(err)
				}
			},
		},
	}

	app.Run(os.Args)
}
