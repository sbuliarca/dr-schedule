package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"

	"github.com/dr-schedule/internal/ds"
	"github.com/dr-schedule/internal/ds/wpamelia"
)

func main() {
	// todo add command parameters
	configureLogger("info", "text")
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	pc := wpamelia.NewClient("https://lavitamed.ro/")

	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		logrus.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		logrus.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.New(client)
	if err != nil {
		logrus.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	// todo: parameterize the calendar id
	cal := ds.NewCal(srv, "6j9vbq93sa5c0rj7jj4jjfj1ek@group.calendar.google.com")

	service := ds.NewService(cal, pc)

	foundDuplicate := true
	for foundDuplicate {
		if foundDuplicate, err = cal.CleanDuplicateEvents(time.Now(), 31); err != nil {
			logrus.Fatalf("failed cleaning up duplicate calendar events on startup: %v", err)
		}
	}

	c := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.VerbosePrintfLogger(logrus.New()))))

	_, err = c.AddFunc("@every 1m", func() {
		logrus.Infof("Started checking for schedule")
		if err := service.SyncSlots(time.Now()); err != nil {
			logrus.Errorf("failed syncing calendar: %v", err)
		}
	})
	if err != nil {
		logrus.Fatalf("could not add scheduling function, err: %v", err)
	}

	c.Start()

	logrus.Info("Started scheduling app")

	waitForShutdown()
	logrus.Infof("Shutdown triggered, waiting for scheduler to finish")

	ctx := c.Stop()
	<-ctx.Done()
}

func configureLogger(level, format string) {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.WithFields(logrus.Fields{"log_level": level}).
			WithError(err).
			Panic("invalid log level")
	}
	logrus.SetLevel(l)

	format = strings.ToLower(format)
	if format != "text" && format != "json" {
		logrus.Panicf("invalid log format: %s", format)
	}
	if format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
}

func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
}

func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		logrus.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		logrus.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logrus.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
