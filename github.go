package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// GitHubObject represents response github object
type GitHubObject struct {
	Sha  string `json:"sha"`
	Type string `json:"type"`
	Url  string `json:"url"`
}

// GitHubResponse represents github repsonse
type GitHubResponse struct {
	Ref    string       `json:"ref"`
	NodeId string       `json:"node_id"`
	Url    string       `json:"url"`
	Object GitHubObject `json:"object"`
}

// helper function to get commit of the request
func getCommit(r Request) (string, error) {
	// https://api.github.com/repos/<repo>/git/refs/tags/<tag>
	rurl := fmt.Sprintf("https://api.github.com/repos/%s/git/refs/tags/%s", r.Repository, r.Tag)
	resp, err := http.Get(rurl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if Config.Verbose > 0 {
		log.Println("request", rurl, "response", string(body))
	}
	if err != nil {
		return "", err
	}
	var rec GitHubResponse
	err = json.Unmarshal(body, &rec)
	if err != nil {
		return "", err
	}
	if rec.Url != rurl {
		msg := fmt.Sprintf("github url does not match %s!=%s", rec.Url, rurl)
		return "", errors.New(msg)
	}
	return rec.Object.Sha, nil
}
