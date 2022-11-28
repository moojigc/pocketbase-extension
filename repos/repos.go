package repos

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

type Repo struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Url         string `json:"url"`
	Homepage    string `json:"homepage"`
	Description string `json:"description"`
}

var logger *log.Logger = log.Default()

func GetRepos() *[]Repo {

	resp, err := http.Get("https://api.github.com/users/moojigc/starred?sort=updated")

	logger.Println("Getting repos")

	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	var repos []Repo

	if err := json.Unmarshal(body, &repos); err != nil {
		log.Fatal(err)
	}

	logger.Printf("Got %d repos", len(repos))

	return &repos
}

func initRepoCollection(app *pocketbase.PocketBase) *models.Collection {
	collection := &models.Collection{
		Name:       "repositories",
		Type:       models.CollectionTypeBase,
		ListRule:   nil,
		ViewRule:   nil,
		CreateRule: types.Pointer("@request.auth.id != ''"),
		UpdateRule: types.Pointer("@request.auth.id != ''"),
		DeleteRule: types.Pointer("@request.auth.id != ''"),
		Schema: schema.NewSchema(
			&schema.SchemaField{
				Name:     "name",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   true,
			},
			&schema.SchemaField{
				Name:     "url",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   true,
			},
			&schema.SchemaField{
				Name:     "homepage",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "description",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
		),
	}
	app.Dao().SaveCollection(collection)
	return collection
}

func LoadOrUpdateRepos(app *pocketbase.PocketBase, retries int) []*models.Record {

	initRepoCollection(app)
	collection, _ := (*app).Dao().FindCollectionByNameOrId("repositories")

	repos := GetRepos()

	var repoRecords []*models.Record

	for _, repo := range *repos {
		record, err := (*app).Dao().FindFirstRecordByData(collection.Id, "name", repo.Name)
		logger.Print(record)
		logger.Print(err)

		if err != nil || record == nil {
			record = models.NewRecord(collection)
		}
		name := repo.Name
		record.Set("name", name)
		record.Set("url", repo.Url)
		record.Set("homepage", repo.Homepage)
		record.Set("description", repo.Description)
		if err := (*app).Dao().SaveRecord(record); err != nil {
			logger.Println(err)
		}
		repoRecords = append(repoRecords, record)
	}

	return repoRecords
}
