package main

import (
	"log"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"

	"github.com/bmatsuo/go-jsontree"
)

func main() {
	ec := echo.New()

	ec.Use(mw.Logger())
	ec.Use(mw.Recover())

	ec.Post("/github/events", func(c *echo.Context) error {
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		json := jsontree.New()
		if err := json.UnmarshalJSON(body); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "JSON expected!"})
		}

		s, _ := json.MarshalJSON()
		log.Println(string(s))

		return c.JSON(http.StatusOK, json)
	})

	log.Print("Starting serve on :9090")
	ec.Run(":9090")
}
