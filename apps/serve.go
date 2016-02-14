package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"../github"
	"../manifest"
	"../manifest/alerts"
)

func main() {
	ec := echo.New()

	ec.Use(middleware.Logger())
	ec.Use(middleware.Recover())

	handlers := []github.PushHandler{
		manifest.ManifestHandler{
			Plugins: []manifest.Plugin{
				alerts.ElasticAlertPlugin{},
			},
		},
	}

	ec.Post("/github/events", func(c *echo.Context) error {
		switch c.Request().Header.Get("X-GitHub-Event") {
		case "push":
			event := github.Push{}

			err := c.Bind(&event)
			if err != nil {
				return err
			}

			for _, handler := range handlers {
				go func() {
					err := handler.Handle(event)

					if err != nil {
						log.Printf("%T: %s\n", handler, err)
					}
				}()
			}

			return c.JSON(http.StatusOK, event)

		default:
			return c.String(http.StatusBadRequest, "Only `push` events accepted!")
		}
	})

	log.Print("Starting serve on :9090")

	ec.Run(":9090")
}
