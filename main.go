package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"pb.chimid.rocks/repos"
)

func main() {
	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// add new "GET /hello" route to the app router (echo)
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/api/@chimid/repos",
			Handler: func(c echo.Context) error {
				records := repos.LoadOrUpdateRepos(app, 0)
				return c.JSON(http.StatusOK, map[string]interface{}{
					"items": records,
				})
			},
			Middlewares: []echo.MiddlewareFunc{apis.ActivityLogger(app)},
			Name:        "LoadRepos",
		})

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
