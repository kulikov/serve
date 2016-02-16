package webhook

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func StartWebHookServer(port int, handlersDir string) {
	ec := echo.New()
	ec.Use(middleware.Logger())
	ec.Use(middleware.Recover())

	ec.Post("/github/events", func(c *echo.Context) error {
		defer c.Request().Body.Close()

		handlers := make([]string, 0)
		err := filepath.Walk(handlersDir, func(path string, f os.FileInfo, err error) error {
			if err == nil && !f.IsDir() {
				handlers = append(handlers, path)
			}
			return err
		})

		if err != nil {
			log.Panicf("Error on find hadnlers: %s\n", err)
		}

		eventType := c.Request().Header.Get("X-GitHub-Event")
		payload, err := ioutil.ReadAll(c.Request().Body)

		log.Printf("Receive github event '%s': \n%s\n", eventType, string(payload))

		for _, handler := range handlers {
			go runHandler(handler, "github", eventType, string(payload))
		}

		return err
	})

	log.Printf("Starting web hook server on :%v\n", port)

	ec.Run(":" + strconv.Itoa(port))
}

func runHandler(handler string, source string, eventType string, payload string) {
	log.Printf("Run %s", handler)

	out, err := exec.Command("/bin/bash", handler, "--source", source, "--event", eventType, "--payload", payload).CombinedOutput()

	if err != nil {
		log.Printf("Error: %s\n\n", err)
	} else {
		log.Printf("Out: \n%s\n\n", string(out))
	}
}
