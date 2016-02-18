package main

import (
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/codegangsta/cli"

	"../gocd"
	"../manifest"
	"../manifest/alerts"
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
					Name: "start",
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
			Name: "manifest",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config",
					Value: "config.yml",
					Usage: "Path to config.yaml file",
				},
			},
			Subcommands: []cli.Command{
				{
					Name:  "github-hook",
					Usage: "Handle github hook event and check manifest changes",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "payload"},
					},
					Action: func(c *cli.Context) {
						conf, err := manifest.InitConfig(c.GlobalString("config"))
						if err != nil {
							log.Fatal(err)
						}

						if err := manifest.HandleGithubChanges(conf, manifestPlugins, c.String("payload")); err != nil {
							log.Fatal(err)
						}
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
