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
	marathon "github.com/gambol99/go-marathon"

	"github.com/kulikov/serve/utils"
)

func MarathonCommand() cli.Command {
	return cli.Command{
		Name:  "marathon",
		Usage: "Deploy serivce into marathon",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "name"},
			cli.StringFlag{Name: "version"},
		},
		Subcommands: []cli.Command{
			{
				Name:            "start-site",
				Action: func(c *cli.Context) {
					marathonConf := marathon.NewDefaultConfig()
					marathonConf.URL = "http://" + c.GlobalString("marathon-host") + ":8080"
					marathonApi, _ := marathon.NewClient(marathonConf)

					app := &marathon.Application{
						ID: c.GlobalString("name") + "-v" + c.GlobalString("version"),
						Cmd: "bin/serve service --name $(echo '#{project}' | sed 's/[^a-z0-9]/-/gI') --version #{version}.${GO_PIPELINE_LABEL} --qa-domain '#{domain}' --location '#{location}' --port \\$PORT0",
					}

					depId, err := marathonApi.UpdateApplication(app, false)
				},
			},
		},
	}
}
