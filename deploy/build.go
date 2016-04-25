package deploy

import (
	"log"

	"github.com/codegangsta/cli"
)

func BuildCommand() cli.Command {
	return cli.Command{
		Name:  "build",
		Usage: "Duild package",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "branch"},
			cli.StringFlag{Name: "build-number"},
		},
		Action: func(c *cli.Context) {
			log.Println("build")
		},
	}
}
