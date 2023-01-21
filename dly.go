package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// version extracted from tag added during compilation
var compiledVersion string

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
	log.Debug().Msgf("debugging initialized, version %s %s/%s", compiledVersion, runtime.GOOS, runtime.GOARCH)
}

func main() {
	var content []byte

	// if debugging check immediately for new version
	if os.Getenv("DLY_DEBUG") == "yes" {
		CheckUpdateNow()
	}

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
	CheckUpdateNow()
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
	// check for the new parameter AppendHashtag
	if conf.AppendHashtag != "" {
		// new parameter
		newText = fmt.Sprintf("%s #%s", newText, conf.AppendHashtag)
	} else if conf.AddHashtag {
		// legacy, deprecated parameter
		newText = fmt.Sprintf("%s #%s", newText, conf.HashtagToAdd)
		log.Warn().Msgf("the configuration parameter AddHashtag is deprecated and will be removed in a future version. Use AppendHashtag instead")
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
