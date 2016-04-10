package main

import (
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/consul/api"
	"github.com/kulikov/hookup"

	"github.com/kulikov/serve/utils"
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
				cli.StringFlag{Name: "name"},
				cli.StringFlag{Name: "version", Value: "0.0"},
				cli.StringFlag{Name: "host"},
				cli.StringFlag{Name: "location", Value: "/"},
				cli.StringFlag{Name: "staging", Value: "live"},
				cli.StringFlag{Name: "port"},
			},
			Subcommands: []cli.Command{
				{
					Name:            "start",
					SkipFlagParsing: true,
					Action: func(c *cli.Context) {
						consul, _ := api.NewClient(api.DefaultConfig())

						serviceId := c.GlobalString("name") + "-" + c.GlobalString("version") + "-" + c.GlobalString("port")

						if err := consul.Agent().ServiceRegister(&api.AgentServiceRegistration{
							ID:   serviceId,
							Name: c.GlobalString("name"),
							Tags: []string{
								"version:" + c.GlobalString("version"),
								"host:" + c.GlobalString("host"),
								"location:" + c.GlobalString("location"),
								"staging:" + c.GlobalString("staging"),
							},
							Port:              c.GlobalInt("port"),
							EnableTagOverride: true,
							Check: &api.AgentServiceCheck{
								TCP:      "localhost:" + c.GlobalString("port"),
								Interval: "5s",
							},
						}); err != nil {
							log.Fatal(err)
						}

						if outdated, _, err := consul.Catalog().Service(c.GlobalString("name"), "staging:"+c.GlobalString("staging"), &api.QueryOptions{}); err == nil {
							for _, serv := range outdated {
								if serv.ServiceID != serviceId {
									log.Println("Deregister outdated", serv.ServiceID)

									if _, err := consul.Catalog().Deregister(&api.CatalogDeregistration{
										Node:      serv.Node,
										ServiceID: serv.ServiceID,
									}, &api.WriteOptions{}); err != nil {
										log.Println("Service deregistration error:", err)
									}
								}
							}
						}

						if err := syscall.Exec("/bin/bash", append([]string{"bash", "-c"}, strings.Join(c.Args(), " ")), os.Environ()); err != nil {
							log.Fatal(err)
						}
					},
				},
				{
					Name:            "release",
					Action: func(c *cli.Context) {
						consul, _ := api.NewClient(api.DefaultConfig())

						version := "version:"+c.GlobalString("version")

						if staged, _, err := consul.Catalog().Service(c.GlobalString("name"), version, &api.QueryOptions{}); err == nil {
							for _, serv := range staged {
								if _, err := consul.Catalog().Register(&api.CatalogRegistration{
									Node: serv.Node,
									Address: serv.Address,
									Service: &api.AgentService{
										ID: serv.ServiceID,
										Service: serv.ServiceName,
										Tags: append(utils.Filter(serv.ServiceTags, func(t string) bool { return !strings.HasPrefix(t, "staging:") }), "staging:live"),
										Port: serv.ServicePort,
										Address: serv.ServiceAddress,
										EnableTagOverride: serv.ServiceEnableTagOverride,
									},
									Check: &api.AgentCheck{
										Node: serv.Node,
										CheckID: "service:" + serv.ServiceID,
										ServiceID: serv.ServiceID,
										ServiceName: serv.ServiceName,
										Status: api.HealthPassing,
									},
								}, &api.WriteOptions{}); err != nil {
									log.Println("Register:", err)
								}
							}
						}

						if outdated, _, err := consul.Catalog().Service(c.GlobalString("name"), "staging:live", &api.QueryOptions{}); err == nil {
							for _, serv := range outdated {
								if !utils.Contains(version, serv.ServiceTags) {
									log.Println("Deregister outdated live", serv.ServiceID)

									if _, err := consul.Catalog().Deregister(&api.CatalogDeregistration{
										Node:      serv.Node,
										ServiceID: serv.ServiceID,
									}, &api.WriteOptions{}); err != nil {
										log.Println("Service deregistration error:", err)
									}
								}
							}
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
