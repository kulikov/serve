package commands

import (
	"log"

	"github.com/codegangsta/cli"

	"github.com/kulikov/hookup"
	"github.com/kulikov/serve/manifest"
)

func WebhookServerCommand(manifestPlugins []manifest.Plugin) cli.Command {
	return cli.Command{
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
	}
}
