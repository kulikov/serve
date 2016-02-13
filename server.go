package main

import (
	"log"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"strings"
	"encoding/base64"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gopkg.in/yaml.v2"

	"./manifest"
	"./github"
	"./utils"
)

func main() {
	ec := echo.New()

	ec.Use(middleware.Logger())
	ec.Use(middleware.Recover())

	ec.Post("/github/events", func(c *echo.Context) error {
		eventType := c.Request().Header.Get("X-GitHub-Event")
		if (eventType != "push") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Only `push` events accepted!"})
		}

		event := &github.Push{}

		err := c.Bind(event)
		if err != nil {
			return err
		}

		log.Println(event)

		modified := false

		for _, commit := range event.Commits {
			log.Println(append(commit.Added, commit.Modified...))

			if utils.Contains("manifest.yml", append(commit.Added, commit.Modified...)) {
				modified = true
			}
		}

		if modified {
			resp, err := http.Get(strings.Replace(event.Repository.ContentUrl, "{+path}", "manifest.yml", 1))
			defer resp.Body.Close()

			if err != nil {
				return err
			}

			fileContent := &github.FileContent{}
			data, _ := ioutil.ReadAll(resp.Body)

			err = json.Unmarshal(data, fileContent)
			if err != nil {
				return err
			}

			data, err = base64.StdEncoding.DecodeString(fileContent.Content)
			if err != nil {
				return err
			}

			manifest := &manifest.Manifest{}
			yaml.Unmarshal(data, manifest)

			log.Println(manifest)
		}

		return c.JSON(http.StatusOK, event)
	})

	log.Print("Starting serve on :9090")
	ec.Run(":9090")
}
