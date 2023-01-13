package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	DailyNotesPath string `yaml:"DailyNotesPath"`
	FilenameFormat string `yaml:"FilenameFormat"`
	AddTimestamp   bool   `yaml:"AddTimestamp,omitempty"`
	AddHashtag     bool   `yaml:"AddHashtag,omitempty"`
	HashtagToAdd   string `yaml:"HashtagToAdd,omitempty"`
}

var log zerolog.Logger

// initialize logging
func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("%s", i))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}
	// set level
	if os.Getenv("DLY_DEBUG") == "yes" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log = zerolog.New(output).With().Timestamp().Logger()
}

func main() {
	var content []byte

	// get user's configuration
	conf := getConfiguration()

	// get text to add to daily note
	textToAdd := getTextToAdd()

	// build today's note full path & filename
	todayFile := filepath.Join(conf.DailyNotesPath, fmt.Sprintf("%s.md", time.Now().Format(conf.FilenameFormat)))
	// get the content of today's note (can be empty if there is no note today)
	content = getTodayNote(todayFile)
	// add to that content what we have on the app input, and place it cleverly in the note, and get that back
	content = addToTodayNote(content, fmt.Sprintf("%s ", textToAdd), conf)
	// write the note back to file
	writeTodayNote(content, todayFile, conf)
}

func getTextToAdd() (textToAdd string) {
	if len(os.Args) == 1 {
		// interactive mode
		fmt.Print("⤑ ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		textToAdd = scanner.Text()
	} else if len(os.Args) == 3 && os.Args[1] == "--quotedText" {
		// text to add is provided quoted
		textToAdd = os.Args[2]
	} else {
		textToAdd = strings.Join(os.Args[1:], " ")
	}
	return textToAdd
}

func addToTodayNote(note []byte, newText string, conf Configuration) (content []byte) {
	// should we prefix with a timestamp?
	if conf.AddTimestamp {
		newText = fmt.Sprintf("**%s** %s", time.Now().Format("15:04"), newText)
	}
	// should we postfix it with an hashtag?
	if conf.AddHashtag {
		newText = fmt.Sprintf("%s #%s", newText, conf.HashtagToAdd)
	}
	// check if the note does not exist
	if len(note) == 0 {
		log.Debug().Msgf("empty note for today")
		return append(note, []byte(fmt.Sprintf("- %s\n", newText))...)
	}
	// check if the note ends with a newline
	if strings.HasSuffix(string(note), "\n") {
		// add a markdown bullet point and send back
		log.Debug().Msgf("text starts on a new line")
		return append(note, []byte(fmt.Sprintf("- %s\n", newText))...)
	}
	// check if the note ends with a -
	if strings.HasSuffix(string(note), "-") {
		// add a space and send back
		log.Debug().Msgf("need to add a newline after -")
		return append(note, []byte(fmt.Sprintf(" %s\n", newText))...)
	}
	// prefix with a newline and send a bullt point back
	log.Debug().Msgf("need to add a newline and add -")
	return append(note, []byte(fmt.Sprintf("\n- %s\n", newText))...)
}

func writeTodayNote(content []byte, todayFile string, conf Configuration) {
	var err error

	// create backup note in temp folder by searching for typical temp folder
	var tmpHandler *os.File
	for _, tempFilePath := range []string{
		filepath.Join("C:", "TEMP", filepath.Base(todayFile)),
		filepath.Join("/", "tmp", filepath.Base(todayFile)),
	} {
		tmpHandler, err = os.Create(tempFilePath)
		if err != nil {
			log.Debug().Msgf("cannot create temporary file %s: %v", tempFilePath, err)
		} else {
			// we gave a good file
			log.Debug().Msgf("backup note is %s", tempFilePath)
			break
		}
	}
	// can we write the backup?
	if tmpHandler == nil {
		log.Error().Msg("cannot find a place to place the backup, skipping it 😐")
	} else {
		_, err := tmpHandler.Write(content)
		if err != nil {
			log.Error().Msgf("could not write the backup file: %v", err)
		}
	}
	tmpHandler.Close()

	// write the updated note
	var noteHandler *os.File
	noteHandler, err = os.Create(todayFile)
	if err != nil {
		log.Fatal().Msgf("could not open the note %s: %v", todayFile, err)
	}
	_, err = noteHandler.Write(content)
	if err != nil {
		log.Fatal().Msgf("could not write the note %s: %v", todayFile, err)
	}
	log.Info().Msgf("note %s updated", todayFile)
}

func getTodayNote(todayFile string) (content []byte) {
	var err error

	// check if today's file exists at all
	_, err = os.Stat(todayFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// it doe snot exist, send back empty content
			log.Info().Msgf("no note for today, will create one")
			return []byte{}
		} else {
			// there is a more serious problem
			log.Fatal().Msgf("cannot check for today's note %s: %v", todayFile, err)
		}
	}
	// the file exists, read it
	handler, err := os.Open(todayFile)
	if err != nil {
		log.Fatal().Msgf("cannot open today's note %s: %v", todayFile, err)
	}
	content, err = io.ReadAll(handler)
	if err != nil {
		log.Fatal().Msgf("cannot read today's note %s: %v", todayFile, err)
	}
	handler.Close()
	return content
}

func getConfiguration() (conf Configuration) {
	var err error

	// check if home variable exists at all
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
	// make sure we actually have homeDir set
	if homeDir == "" {
		log.Fatal().Msgf("home directory is not set, I cannot locate your config. Please report this at https://github.com/wsw70/dly/issues/new with your OS")
	}
	configFileDir := filepath.Join(homeDir, ".config", "dly")
	configFilePath := filepath.Join(configFileDir, "dly.yml")
	_, err = os.Stat(configFilePath)
	// checking if the config file is there or not
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			log.Info().Msgf("no config file present")
			// create a minimal config file
			err = os.MkdirAll(configFileDir, 0770)
			if err != nil {
				log.Fatal().Msgf("cannot create path %s for config file: %v", configFileDir, err)
			}
			f, err := os.Create(configFilePath)
			if err != nil {
				log.Fatal().Msgf("cannot create empty config file at %s for config file: %v", configFilePath, err)
			}
			// initialize content with defaults
			defaultValues := Configuration{
				DailyNotesPath: "YOU MUST SET THIS to your journal folder",
				FilenameFormat: "2006_01_02",
				AddTimestamp:   true,
				AddHashtag:     true,
				HashtagToAdd:   "from-cli",
			}
			defaultValuesB, _ := yaml.Marshal(defaultValues)
			_, err = f.Write(defaultValuesB)
			if err != nil {
				log.Fatal().Msgf("cannot add line to config file at %s for config file: %v", configFilePath, err)
			}
			log.Info().Msgf("minimal config file created at %s, you MUST now edit it to at least set the path to daily notes", configFilePath)
			os.Exit(2)
		} else {
			log.Fatal().Msgf("cannot check for presence of the config file: %v", err)
		}
	}

	// get the contents of the config file
	handler, err := os.Open(configFilePath)
	if err != nil {
		log.Fatal().Msgf("cannot open config file at %s: %v", configFilePath, err)
	}
	content, err := io.ReadAll(handler)
	if err != nil {
		log.Fatal().Msgf("cannot read config file at %s: %v", configFilePath, err)
	}
	handler.Close()
	err = yaml.Unmarshal(content, &conf)
	if err != nil {
		log.Fatal().Msgf("cannot unmarshal config file at %s: %v", configFilePath, err)
	}
	return conf
}
