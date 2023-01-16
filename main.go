package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"pb.chimid.rocks/ipgeoservice"
	"pb.chimid.rocks/repos"
	"pb.chimid.rocks/visit_public"
)

func main() {
	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// update starred repos with github api
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
		// custom visit api
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

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// custom unique visitor api
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/api/@chimid/visit",
			Handler: func(c echo.Context) error {
				request := c.Request()

				// return c.JSON(http.StatusOK, map[string]interface{}{
				// 	"X-Visitor-Token": request.Header.Get("X-Visitor-Token"),
				// })

				if request.Header.Get("X-Visitor-Token") == "" {
					return c.JSON(http.StatusBadRequest, map[string]interface{}{
						"message": "Missing X-Visitor-Token header",
					})
				}

				var total int
				visitHashExp := dbx.HashExp{"unique_visitor_token": request.Header.Get("X-Visitor-Token")}
				records := []*visit_public.VisitPublic{}

				if err := app.DB().
					Select("count(*)").
					From("visit").
					AndWhere(visitHashExp).
					Row(&total); err != nil {
					panic(err)
				}

				if err := app.DB().
					Select("`visit`.*").
					From("visit").
					AndWhere(visitHashExp).
					OrderBy("created DESC").
					Limit(100).
					All(&records); err != nil {
					panic(err)
				}

				for _, record := range records {
					record.MatchIp(c.RealIP())
				}

				return c.JSON(http.StatusOK, map[string]interface{}{
					"totalItems": total,
					"items":      records,
				})
			},
		})
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
