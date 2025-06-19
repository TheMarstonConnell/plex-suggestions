package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// PlexMovie represents a single movie in the Plex XML response
type PlexMovie struct {
	Title string `xml:"title,attr"`
	Year  string `xml:"year,attr"`
}

// PlexMediaContainer represents the root XML structure
type PlexMediaContainer struct {
	Movies []PlexMovie `xml:"Video"`
}

func runApp() error {
	plexToken := os.Getenv("PLEX_TOKEN")
	plexIP := os.Getenv("PLEX_IP")
	sectionKey := os.Getenv("PLEX_SECTION_KEY")

	if plexToken == "" || plexIP == "" || sectionKey == "" {
		return fmt.Errorf("missing required environment variables. Please check your .env file")
	}

	url := fmt.Sprintf("http://%s:32400/library/sections/%s/all?X-Plex-Token=%s", plexIP, sectionKey, plexToken)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch movies: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	var container PlexMediaContainer
	err = xml.Unmarshal(body, &container)
	if err != nil {
		return fmt.Errorf("failed to parse XML: %v", err)
	}

	log.Info().Int("movieCount", len(container.Movies)).Msg("Retrieved movies from Plex")

	listOfMovies := make([]string, 0)
	for _, movie := range container.Movies {
		listOfMovies = append(listOfMovies, fmt.Sprintf("%s (%s)", movie.Title, movie.Year))
	}

	suggestions, err := GetSuggestionsForPlexMovies(listOfMovies)
	if err != nil {
		return fmt.Errorf("error getting suggestions: %v", err)
	}

	suggestionsList := strings.Split(suggestions, ",")
	log.Info().Int("suggestionCount", len(suggestionsList)).Msg("Received AI movie suggestions")

	for _, suggestion := range suggestionsList {
		RequestSpecificMovie(suggestion)
	}

	return nil
}

func main() {
	// Configure zerolog for pretty console output
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load environment variables from .env file (optional)
	if err := godotenv.Load(); err != nil {
		log.Info().Msg("Could not load .env file (that's okay if running in production or with env vars set)")
	}

	// Get run interval from environment
	runIntervalStr := os.Getenv("RUN_INTERVAL_DAYS")
	if runIntervalStr == "" {
		runIntervalStr = "0" // Default to run once
	}

	runIntervalDays, err := strconv.Atoi(runIntervalStr)
	if err != nil {
		log.Fatal().Str("value", runIntervalStr).Msg("Invalid RUN_INTERVAL_DAYS value")
	}

	if runIntervalDays == 0 {
		// Run once and exit
		log.Info().Msg("Running once and exiting...")
		if err := runApp(); err != nil {
			log.Fatal().Err(err).Msg("Error running app")
		}
		log.Info().Msg("Completed successfully!")
		return
	}

	// Run in a loop
	interval := time.Duration(runIntervalDays) * 24 * time.Hour
	log.Info().Int("intervalDays", runIntervalDays).Dur("interval", interval).Msg("Starting continuous mode")

	for {
		log.Info().Msg("Starting movie suggestion run...")

		if err := runApp(); err != nil {
			log.Error().Err(err).Msg("Error running app")
		} else {
			log.Info().Msg("Completed successfully!")
		}

		log.Info().Dur("sleepDuration", interval).Msg("Sleeping until next run...")
		time.Sleep(interval)
	}
}
