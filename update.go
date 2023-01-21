package main

import (
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/golang-module/carbon/v2"
)

var buildTime string

// TODO finish the timed check (once every 10 days only)
// func checkUpdateWithDelay() {
// 	var err error
// 	var lastUpdate time.Time

// 	content, err := readConfigFile(getConfigDir(), "lastUpdate.txt")
// 	if err != nil {
// 		// cannot get read file -> last check never
// 		lastUpdate = time.Time{}
// 	} else {
// 		lastUpdate, err = time.Parse(string(content), time.RFC3339)
// 		if err != nil {
// 			// cannot get read file -> last check never
// 			lastUpdate = time.Time{}
// 		}
// 	}

// 	// should we check version?
// 	if lastUpdate.Unix() < (time.Now().Unix() - 60*60*24*7) {
// 		// yes we should, last update was more than 7 days ago

// 	}

// }

func CheckUpdateNow() {
	type releaseApiT struct {
		Name        string `json:"name"`
		PublishedAt string `json:"published_at"`
	}

	var err error

	// parse buildTime
	buildTimeInt, err := strconv.Atoi(buildTime)
	if err != nil {
		log.Debug().Msgf("cannot parse buildTime to int: %v", err)
		return
	}
	buildTimeParsed := carbon.CreateFromTimestamp(int64(buildTimeInt))

	// parse GitHub release time
	var lastUpdateOnGithub releaseApiT
	client := resty.New()
	_, err = client.R().
		SetResult(&lastUpdateOnGithub).
		Get("https://api.github.com/repos/wsw70/dly/releases/latest")
	if err != nil {
		// something went wrong, will try another time
		return
	}
	githubReleaseTime := carbon.Parse(lastUpdateOnGithub.PublishedAt)

	// now compare both
	if buildTimeParsed.Lt(githubReleaseTime) {
		log.Warn().Msgf("new version %s available at https://github.com/wsw70/dly/releases/latest", lastUpdateOnGithub.Name)
	} else {
		log.Debug().Msgf("no new version")
	}
	log.Debug().Msgf("releases: local %v, GitHub %v", buildTimeParsed.ToIso8601String(), githubReleaseTime.ToIso8601String())
}
