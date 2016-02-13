package main

import (
	"log"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"

	"github.com/bmatsuo/go-jsontree"
)

type Error struct {
	error string
}

func main() {
	e := echo.New()

	e.Use(mw.Logger())
	e.Use(mw.Recover())

	e.Post("/github/payload", func(c *echo.Context) error {
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		json := jsontree.New()
		if err := json.UnmarshalJSON(body); err != nil {
			return c.JSON(http.StatusBadRequest, Error{"JSON expected!"})
		}

		s, _ := json.MarshalJSON()
		log.Println(string(s))

		return c.JSON(http.StatusOK, json)
	})

	log.Print("Starting serve...")
	e.Run(":1323")
}
