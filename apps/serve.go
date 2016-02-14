package main

import (
	"log"
	"flag"

	"github.com/spf13/viper"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"../github"
	"../manifest"
	"../manifest/alerts"
)

func main() {
	configFile := flag.String("config", "/etc/serve.yml", "Path to config file")
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
				alerts.ElasticAlertPlugin{},
			},
		},
	))

	log.Print("Starting serve on :9090")

	ec.Run(":9090")
}
