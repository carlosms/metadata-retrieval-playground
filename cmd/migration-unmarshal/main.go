package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/carlosms/metadata-retrieval-playground/internal/store"
	"github.com/google/go-github/github"
)

func main() {
	// As a test, read the first json file for PRs, and save as google/go-github objects
	b, err := ioutil.ReadFile(filepath.FromSlash("./downloads/carlosms-test-org-133497/pull_requests_000001.json"))
	if err != nil {
		panic(err)
	}

	prs := []*github.PullRequest{}

	err = json.Unmarshal(b, &prs)
	if err != nil {
		panic(err)
	}

	storer := store.StdoutStorer{}
	for _, pr := range prs {
		storer.SavePullRequest(pr)

	}
}
