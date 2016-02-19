package main

import (
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/kulikov/hookup"

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
		{
			Name: "service",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "manifest"},
				cli.StringFlag{Name: "port"},
			},
			Subcommands: []cli.Command{
				{
					Name:            "start",
					SkipFlagParsing: true,
					Action: func(c *cli.Context) {
						err := syscall.Exec("/bin/bash", append([]string{"bash", "-c"}, strings.Join(c.Args(), " ")), os.Environ())

						if err != nil {
							log.Println(err)
						}
					},
				},
			},
		},
		{
			Name:  "webhook-server",
			Usage: "Start webhook http sever and handle github hook event for check manifest changes",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "port",
					Value: "9090",
				},
				cli.StringFlag{
					Name:  "config",
					Value: "config.yml",
					Usage: "Path to config.yml file",
				},
			},
			Action: func(c *cli.Context) {
				conf, err := manifest.InitConfig(c.String("config"))
				if err != nil {
					log.Fatalf("Error load config: %v", err)
				}

				hookup.StartWebhookServer(c.Int("port"), func(source string, eventType string, payload string) {
					if source == "github" && eventType == "push" {
						if err := manifest.HandleGithubChanges(conf, manifestPlugins, payload); err != nil {
							log.Printf("Error %v", err)
						}
					}
				})
			},
		},
	}

	app.Run(os.Args)
}
