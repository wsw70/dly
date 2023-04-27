//go:build !windows

package main

func notifyAboutNote() {
	log.Warn().Msgf("notification is not supported on this platform")
}

func notifyAboutNewVersion(name string, conf Configuration) {
	log.Warn().Msgf("new version %s available at https://github.com/wsw70/dly/releases/latest", name)
	if conf.ShowNotificationOnSuccess {
		log.Warn().Msgf("notification popup is not supported on this platform")
	}
}
