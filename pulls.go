package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Comment struct {
	From      User   `json:"user"`
	Body      string `json:"body"`
	UpdatedAt string `json:"updated_at"`
	Line      int    `json:"line"`
	Path      string `json:"path"`
}

type User struct {
	Username string `json:"login"`
}

func GetReviewComments(reviewURL string) ([]Comment, error) {
	req, _ := http.NewRequest(http.MethodGet, reviewURL, nil)

	godotenv.Load()
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("ACCESS_TOKEN")))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, _ := ioutil.ReadAll(res.Body)

	var comments []Comment
	json.Unmarshal(body, &comments)

	return comments, nil
}
