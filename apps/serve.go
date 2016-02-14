package main

import (
	"flag"
	"log"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"

	"../github"
	"../gocd"
	"../manifest"
	"../manifest/alerts"
)

func main() {
	configFile := flag.String("config", "config.yml", "Path to config file")
	flag.Parse()

	conf := viper.New()
	conf.SetConfigFile(*configFile)
	conf.SetConfigType("yml")
	err := conf.ReadInConfig()
	if err != nil {
		log.Panicf("Fatal error config file: %s \n", err)
	}

	ec := echo.New()
	ec.Use(middleware.Logger())
	ec.Use(middleware.Recover())

	ec.Post("/github/events", github.WebhookHandler(
		conf,
		manifest.ManifestHandler{
			Plugins: []manifest.Plugin{
				gocd.DeployPlugin{},
				alerts.GraphiteAlertPlugin{},
				alerts.ElasticAlertPlugin{},
			},
		},
	))

	log.Print("Starting serve on :9090")

	ec.Run(":9090")
}
