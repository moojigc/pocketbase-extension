package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

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

func getVisitSelectQuery(query url.Values, app *pocketbase.PocketBase) *dbx.SelectQuery {
	var originHashExp dbx.HashExp
	uniqueVisitorHashExp := dbx.HashExp{"unique_visitor_token": query.Get("unique_visitor_token")}

	selectQuery := app.DB().
		Select("`visit`.*").
		From("visit")

	if query.Get("origin") != "" {
		originHashExp = dbx.HashExp{"origin": query.Get("origin")}
	}

	if uniqueVisitorHashExp != nil {
		selectQuery.AndWhere(uniqueVisitorHashExp)
	}
	if originHashExp != nil {
		selectQuery.AndWhere(originHashExp)
	}

	return selectQuery
}

func getVisitCountQuery(query url.Values, app *pocketbase.PocketBase) *dbx.SelectQuery {
	var uniqueVisitorHashExp dbx.HashExp
	var originHashExp dbx.HashExp

	if query.Get("unique_visitor_token") != "" {
		uniqueVisitorHashExp = dbx.HashExp{"unique_visitor_token": query.Get("unique_visitor_token")}
	}

	if query.Get("origin") != "" {
		originHashExp = dbx.HashExp{"origin": query.Get("origin")}
	}

	countQuery := app.DB().
		Select("count(*)").
		From("visit")

	if uniqueVisitorHashExp != nil {
		countQuery.AndWhere(uniqueVisitorHashExp)
	}
	if originHashExp != nil {
		countQuery.AndWhere(originHashExp)
	}

	return countQuery
}

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
				request := c.Request()

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

				var originToSave string
				originHeader := request.Header.Get("Origin")
				referringOrigin := request.Header.Get("X-Referring-Origin")

				if referringOrigin != "" {
					originToSave = referringOrigin
				} else {
					originToSave = originHeader
				}

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
				record.Set("origin", originToSave)
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
				total := 0
				records := []*visit_public.VisitPublic{}

				countQuery := getVisitCountQuery(request.URL.Query(), app)

				if err := countQuery.Row(&total); err != nil {
					panic(err)
				}

				if request.Header.Get("X-Visitor-Token") != "" {
					selectQuery := getVisitSelectQuery(request.URL.Query(), app)
					if err := selectQuery.
						OrderBy("created DESC").
						Limit(1).
						All(&records); err != nil {
						fmt.Println(err)
					}
					for _, record := range records {
						record.MatchIp(c.RealIP())
					}

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
