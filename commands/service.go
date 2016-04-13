package commands

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/codegangsta/cli"

	"github.com/hashicorp/consul/api"
	"github.com/kulikov/serve/utils"
)

func ServiceCommand() cli.Command {
	return cli.Command{
		Name:  "service",
		Usage: "Run and release services",
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
					log.Println("Starting", c.Args())

					cmd := exec.Command(c.Args().First(), c.Args().Tail()...)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr

					if err := cmd.Start(); err != nil {
						log.Fatal("Error on process staring", err)
					}

					consul, _ := api.NewClient(api.DefaultConfig())

					serviceId := c.GlobalString("name") + "-" + c.GlobalString("version") + "-" + c.GlobalString("port")

					go func() {
						result := cmd.Wait()
						log.Printf("Command finished with: %v", result)

						log.Println("Deregister service", serviceId, "...")
						if err := consul.Agent().ServiceDeregister(serviceId); err != nil {
							log.Fatal(err)
						}

						log.Println("Deregistered.")

						if exiterr, ok := result.(*exec.ExitError); ok {
							if status, ok := exiterr.Sys().(syscall.WaitStatus); ok && status.Exited() {
								log.Printf("Exit Status: %d", status.ExitStatus())
								os.Exit(status.ExitStatus())
							}
						}

						if result != nil {
							os.Exit(0)
						} else {
							os.Exit(2)
						}
					}()

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
						cmd.Process.Kill()
						log.Fatal(err)
					}

					// Handle SIGINT and SIGTERM.
					ch := make(chan os.Signal)
					signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
					log.Println(<-ch)

					cmd.Process.Kill()

					time.Sleep(time.Second)
					log.Println("Stopped.")
				},
			},
			{
				Name: "release",
				Action: func(c *cli.Context) {
					consul, _ := api.NewClient(api.DefaultConfig())

					if staged, _, err := consul.Catalog().Service(c.GlobalString("name"), "version:"+c.GlobalString("version"), &api.QueryOptions{}); err == nil {
						for _, serv := range staged {
							if _, err := consul.Catalog().Register(&api.CatalogRegistration{
								Node:    serv.Node,
								Address: serv.Address,
								Service: &api.AgentService{
									ID:      serv.ServiceID,
									Service: serv.ServiceName,
									Tags: append(utils.Filter(serv.ServiceTags, func(t string) bool {
										return !strings.HasPrefix(t, "staging:")
									}), "staging:live"),
									Port:              serv.ServicePort,
									Address:           serv.ServiceAddress,
									EnableTagOverride: serv.ServiceEnableTagOverride,
								},
								Check: &api.AgentCheck{
									Node:        serv.Node,
									CheckID:     "service:" + serv.ServiceID,
									ServiceID:   serv.ServiceID,
									ServiceName: serv.ServiceName,
									Status:      api.HealthPassing,
								},
							}, &api.WriteOptions{}); err != nil {
								log.Println("Register:", err)
							}
						}
					}
				},
			},
		},
	}
}
