package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"./manifest"
	"./github"
)

func main() {
	ec := echo.New()

	ec.Use(middleware.Logger())
	ec.Use(middleware.Recover())

	handlers := []github.PushHandler{
		manifest.ManifestHandler{},
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
					handler.Handle(event)
				}()
			}

			return c.JSON(http.StatusOK, event)

		default:
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Only `push` events accepted!"})
		}
	})

	log.Print("Starting serve on :9090")

	ec.Run(":9090")
}
