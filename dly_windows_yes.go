//go:build windows

package main

import (
	"fmt"

	"gopkg.in/toast.v1"
)

func notifyAboutNote() {
	notification := toast.Notification{
		AppID: "dly - your daily note from CLI",
		Title: "Note added",
	}
	err := notification.Push()
	if err != nil {
		log.Error().Msgf("cannot push notification: %v", err)
	}
}

func notifyAboutNewVersion(name string, conf Configuration) {
	notification := toast.Notification{
		AppID:   "dly",
		Title:   "new version",
		Message: fmt.Sprintf("new version %s available at https://github.com/wsw70/dly/releases/latest", name),
	}
	err := notification.Push()
	if err != nil {
		log.Error().Msgf("cannot push notification: %v", err)
		log.Warn().Msgf("new version %s available at https://github.com/wsw70/dly/releases/latest", name)
	}
}
