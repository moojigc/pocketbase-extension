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
	Id          int      `json:"id"`
	Name        string   `json:"name"`
	HtmlUrl     string   `json:"html_url"`
	Homepage    string   `json:"homepage"`
	Description string   `json:"description"`
	Topics      []string `json:"topics"`
}

func (r *Repo) TopicsWereUpdated(topics []string) bool {
	if len(r.Topics) != len(topics) {
		return true
	}

	var shorterArr *[]string

	if len(r.Topics) < len(topics) {
		shorterArr = &r.Topics
	} else {
		shorterArr = &topics
	}

	for index, topic := range *shorterArr {
		if topic != (*shorterArr)[index] {
			return true
		}
	}

	return false
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
				Name:     "github_id",
				Type:     schema.FieldTypeNumber,
				Required: true,
				Unique:   true,
			},
			&schema.SchemaField{
				Name:     "name",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   true,
			},
			&schema.SchemaField{
				Name:     "html_url",
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

func LoadOrUpdateRepos(app *pocketbase.PocketBase, retries int) ([]*models.Record, bool) {

	initRepoCollection(app)
	collection, _ := (*app).Dao().FindCollectionByNameOrId("repositories")
	repos_from_github := GetRepos()
	changeDetected := false

	var repoRecords []*models.Record

	for _, repo := range *repos_from_github {
		record, err := (*app).Dao().FindFirstRecordByData(collection.Id, "github_id", repo.Id)
		logger.Print(record)
		logger.Print(err)

		if err != nil && record == nil {
			logger.Print("Creating new record")
			record = models.NewRecord(collection)
		}

		if record.Get("name") != repo.Name {
			changeDetected = true
		}
		if record.Get("html_url") != repo.HtmlUrl {
			changeDetected = true
		}
		if record.Get("homepage") != repo.Homepage {
			changeDetected = true
		}
		if record.Get("description") != repo.Description {
			changeDetected = true
		}
		if repo.TopicsWereUpdated(record.Get("topics").([]string)) {
			changeDetected = true
		}

		record.Set("github_id", repo.Id)
		record.Set("name", repo.Name)
		record.Set("html_url", repo.HtmlUrl)
		record.Set("homepage", repo.Homepage)
		record.Set("description", repo.Description)
		record.Set("topics", repo.Topics)

		if err := (*app).Dao().SaveRecord(record); err != nil {
			logger.Println(err)
		}
		repoRecords = append(repoRecords, record)

	}

	return repoRecords, changeDetected
}
