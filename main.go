// main.go
package main

import (
	"log"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pocketbase/pocketbase"
)

type Repo struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Url         string `json:"url"`
	Homepage    string `json:"homepage"`
	Description string `json:"description"`
}

func repos() []uint8 {
	// repos := []Repo{}

	resp, err := http.Get("https://api.github.com/users/moojigc/starred?sort=updated")

	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	var result []Repo

	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatal(err)
	}

	for _, repo := range result {
		fmt.Println(repo.Name)
	}

	reformatted, _ := json.Marshal(result)

	return reformatted
}

func main() {
	app := pocketbase.New()

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
