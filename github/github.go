package github

import (
	"github.com/labstack/echo"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

type Push struct {
	Ref        string         `json:"ref"`
	Repository GithubRepo     `json:"repository"`
	Commits    []GithubCommit `json:"commits"`
}

type GithubRepo struct {
	SshUrl     string `json:"ssh_url"`
	ContentUrl string `json:"contents_url"`
}

type GithubCommit struct {
	Modified []string `json:"modified"`
	Added    []string `json:"added"`
	Removed  []string `json:"removed"`
}

type FileContent struct {
	Sha     string `json:"sha"`
	Content string `json:"content"`
}

type PushHandler interface {
	Handle(conf *viper.Viper, event Push) error
}

func WebhookHandler(conf *viper.Viper, handlers ...PushHandler) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		switch c.Request().Header.Get("X-GitHub-Event") {
		case "push":
			event := Push{}

			err := c.Bind(&event)
			if err != nil {
				return err
			}

			for _, handler := range handlers {
				go func(h PushHandler) {
					err := h.Handle(conf, event)

					if err != nil {
						log.Printf("%T: %s\n", h, err)
					}
				}(handler)
			}

			return c.JSON(http.StatusOK, event)

		default:
			return c.String(http.StatusBadRequest, "Only `push` events accepted!")
		}
	}
}
