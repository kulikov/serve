package commands

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
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
			cli.StringFlag{Name: "domain"},
			cli.StringFlag{Name: "location"},
			cli.StringFlag{Name: "staging"},
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

					// wait for child process compelete and unregister it from consul
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
								os.Exit(status.ExitStatus())
							}
						}

						if result != nil {
							os.Exit(0)
						} else {
							os.Exit(2)
						}
					}()

					// Register service to consul
					if err := consul.Agent().ServiceRegister(&api.AgentServiceRegistration{
						ID:                serviceId,
						Name:              c.GlobalString("name"),
						Tags:              mapToList(tagsFromFlags(c)),
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

					// Handle shutdown signals and kill child process
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
									ID:                serv.ServiceID,
									Service:           serv.ServiceName,
									Tags:              mapToList(utils.MergeMaps(ParseTags(serv.ServiceTags), tagsFromFlags(c))),
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

func tagsFromFlags(c *cli.Context) map[string]string {
	tags := make(map[string]string, 0)

	if t := c.GlobalString("version"); t != "" {
		tags["version"] = t
	}

	if t := c.GlobalString("domain"); t != "" {
		tags["domain"] = t
	}

	if t := c.GlobalString("location"); t != "" {
		tags["location"] = t
	}

	if t := c.GlobalString("staging"); t != "" {
		tags["staging"] = t
	}

	return tags
}

func mapToList(m map[string]string) []string {
	out := make([]string, 0)
	for k, v := range m {
		out = append(out, k+":"+v)
	}
	return out
}
