package main

import (
	"time"

	"github.com/go-resty/resty/v2"
)

type releaseApiT struct {
	TagName     string `json:"tag_name"`
	PublishedAt string `json:"published_at"`
}

var buildTime string

func checkUpdate() {
	var err error
	var lastUpdate time.Time

	content, err := readConfigFile(getConfigDir(), "lastUpdate.txt")
	if err != nil {
		// cannot get read file -> last check never
		lastUpdate = time.Time{}
	} else {
		lastUpdate, err = time.Parse(string(content), time.RFC3339)
		if err != nil {
			// cannot get read file -> last check never
			lastUpdate = time.Time{}
		}
	}

	// should we check version?
	if lastUpdate.Unix() < (time.Now().Unix() - 60*60*24*7) {
		// yes we should, last update was more than 7 days
		var lastUpdateOnGithub releaseApiT
		client := resty.New()
		_, err := client.R().
			SetResult(lastUpdateOnGithub).
			Get("https://api.github.com/repos/wsw70/dly/releases/latest")
		if err != nil {
			// something went wrong, will try another time
			return
		}
		githubReleaseTime, err := time.Parse(lastUpdateOnGithub.PublishedAt, time.RFC3339)
		if err != nil {

		}

	}

}
