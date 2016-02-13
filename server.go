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
)

type (
	GithubPush struct {
		Ref        string `json:"ref"`
		Repository GithubRepo `json:"repository"`
		Commits    []GithubCommit `json:"commits"`
	}

	GithubRepo struct {
		ContentUrl string `json:"contents_url"`
	}

	GithubCommit struct {
		Modified []string `json:"modified"`
		Added    []string `json:"added"`
		Removed  []string `json:"removed"`
	}

	GithubFile struct {
		Content string `json:"content"`
	}
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

		event := &GithubPush{}

		err := c.Bind(event)
		if err != nil {
			return err
		}

		log.Println(event)

		modified := false

		for _, commit := range event.Commits {
			log.Println(append(commit.Added, commit.Modified...))

			log.Println("index: ", contains("manifest.yml", append(commit.Added, commit.Modified...)))

			if contains("manifest.yml", append(commit.Added, commit.Modified...)) {
				modified = true
			}
		}

		if modified {
			resp, err := http.Get(strings.Replace(event.Repository.ContentUrl, "{+path}", "manifest.yml", 1))
			defer resp.Body.Close()

			if err != nil {
				return err
			}

			manifestFile := &GithubFile{}
			data, _ := ioutil.ReadAll(resp.Body)

			err = json.Unmarshal(data, manifestFile)
			if err != nil {
				return err
			}

			data, err = base64.StdEncoding.DecodeString(manifestFile.Content)
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

func contains(elm string, list []string) bool {
	for _, v := range list {
		if v == elm {
			return true
		}
	}
	return false
}
