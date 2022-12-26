package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"pb.chimid.rocks/ipgeoservice"
	"pb.chimid.rocks/repos"
)

func main() {
	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// add new "GET /hello" route to the app router (echo)
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/api/@chimid/repos",
			Handler: func(c echo.Context) error {
				records, changeDetected := repos.LoadOrUpdateRepos(app, 0)
				return c.JSON(http.StatusOK, map[string]interface{}{
					"change_detected": changeDetected,
					"items":           records,
				})
			},
			Middlewares: []echo.MiddlewareFunc{apis.ActivityLogger(app)},
			Name:        "LoadRepos",
		})

		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/api/@chimid/visit",
			Handler: func(c echo.Context) error {
				ipGeo := ipgeoservice.GetIpGeo(c.RealIP())

				collection, err := app.Dao().FindCollectionByNameOrId("visit")
				if err != nil {
					log.Println(err)
					return c.JSON(http.StatusInternalServerError, map[string]interface{}{
						"message": "Internal Server Error",
					})
				}
				record := models.NewRecord(
					collection,
				)

				record.Set("unique_visitor_token", c.Request().Header.Get("X-Visitor-Token"))
				record.Set("ip_address", ipGeo.Ip)
				record.Set("country", ipGeo.Country)
				record.Set("city", ipGeo.City)
				record.Set("region", ipGeo.Region)
				record.Set("postal_code", ipGeo.PostalCode)
				record.Set("latitude", ipGeo.Latitude)
				record.Set("longitude", ipGeo.Longitude)
				record.Set("user_agent", c.Request().UserAgent())
				record.Set("referred_by", c.Request().Referer())
				record.Set("origin", c.Request().Header.Get("Origin"))
				record.Set("isp", ipGeo.Connection.Isp)
				record.Set("org", ipGeo.Connection.Org)

				app.Dao().SaveRecord(record)

				return c.NoContent(http.StatusCreated)
			},
			Middlewares: []echo.MiddlewareFunc{apis.ActivityLogger(app)},
			Name:        "CreateVisit",
		})

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
