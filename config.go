package main

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	DailyNotesPath               string `yaml:"DailyNotesPath"`
	FilenameFormat               string `yaml:"FilenameFormat"`
	AddTimestamp                 bool   `yaml:"AddTimestamp"`
	AppendHashtag                string `yaml:"AppendHashtag"`
	ShowNotificationOnSuccess    bool   `yaml:"ShowNotificationOnSuccess"`
	ShowNotificationOnNewVersion bool   `yaml:"ShowNotificationOnNewVersion"`
}

// getConfigDir queries the environment to recover the users's home and builds the path to the config directory
func getConfigDir() (configDir string) {
	// check if a "home" variable is set
	var homeDir string
	var homeExists bool
	for _, homeDirEnv := range []string{"USERPROFILE", "HOME"} {
		homeDir, homeExists = os.LookupEnv(homeDirEnv)
		if homeExists {
			// we have a match for home dir
			log.Debug().Msgf("found home directory: %v", homeDir)
			break
		}
	}
	// check if we found a home dir
	if homeDir == "" {
		log.Fatal().Msgf("home directory is not set, I cannot locate your config. Please raise an Issue with the --version string")
	}

	// get the name of the program
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		log.Fatal().Msgf("cannot read build info")
	}
	programName := strings.Split(buildInfo.Path, "/")[2]
	return path.Join(homeDir, ".config", programName)
}

// readConfigFile retrieves the specific configuration file from a given directory
func readConfigFile(programDir string, configFile string) (content []byte, err error) {
	fullPathToConfigFile := filepath.Join(getConfigDir(), configFile)
	handler, err := os.Open(fullPathToConfigFile)
	if err != nil {
		log.Debug().Msgf("cannot open config file %s: %v", fullPathToConfigFile, err)
		return nil, err
	}
	content, err = io.ReadAll(handler)
	if err != nil {
		log.Debug().Msgf("cannot read config file at %s: %v", fullPathToConfigFile, err)
	}
	handler.Close()
	return content, nil
}

// getConfiguration retreives and parses dly.yml
func getConfiguration() (conf Configuration) {
	var err error

	configDir := getConfigDir()
	content, err := readConfigFile(configDir, "dly.yml")
	if err != nil {
		// if the file does not exist, create it
		if errors.Is(err, fs.ErrNotExist) {
			log.Info().Msgf("no config file present")
			// create a minimal config file
			err = os.MkdirAll(configDir, 0770)
			if err != nil {
				log.Fatal().Msgf("cannot create path %s for config file: %v", configDir, err)
			}
			configFile := filepath.Join(configDir, "dly.yml")
			f, err := os.Create(configFile)
			if err != nil {
				log.Fatal().Msgf("cannot create empty config file at %s for config file: %v", configFile, err)
			}
			// initialize content with defaults
			defaultValues := Configuration{
				DailyNotesPath:               "YOU MUST SET THIS to your journal folder",
				FilenameFormat:               "2006_01_02",
				AddTimestamp:                 true,
				AppendHashtag:                "from-cli",
				ShowNotificationOnSuccess:    true,
				ShowNotificationOnNewVersion: true,
			}
			defaultValuesB, _ := yaml.Marshal(defaultValues)
			_, err = f.Write(defaultValuesB)
			if err != nil {
				log.Fatal().Msgf("cannot add line to config file at %s for config file: %v", configFile, err)
			}
			log.Info().Msgf("minimal config file created at %s, you MUST now edit it to at least set the path to daily notes", configFile)
			os.Exit(2)
		} else {
			// the error was not a missing file
			log.Fatal().Msgf("cannot check for presence of the config file: %v", err)
		}
	}

	// extract the configuration
	err = yaml.Unmarshal(content, &conf)
	if err != nil {
		log.Fatal().Msgf("cannot unmarshal config file dly.yml: %v", err)
	}

	return conf
}
