package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
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

func main() {
	// Load environment variables from .env file (optional)
	if err := godotenv.Load(); err != nil {
		fmt.Println("Could not load .env file (that's okay if running in production or with env vars set)")
	}

	plexToken := os.Getenv("PLEX_TOKEN")
	plexIP := os.Getenv("PLEX_IP")
	sectionKey := os.Getenv("PLEX_SECTION_KEY")

	if plexToken == "" || plexIP == "" || sectionKey == "" {
		log.Fatal("Missing required environment variables. Please check your .env file.")
	}

	url := fmt.Sprintf("http://%s:32400/library/sections/%s/all?X-Plex-Token=%s", plexIP, sectionKey, plexToken)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to fetch movies: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Non-200 response: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var container PlexMediaContainer
	err = xml.Unmarshal(body, &container)
	if err != nil {
		log.Fatalf("Failed to parse XML: %v", err)
	}

	fmt.Printf("Found %d movies:\n", len(container.Movies))
	listOfMovies := make([]string, 0)
	for _, movie := range container.Movies {
		listOfMovies = append(listOfMovies, fmt.Sprintf("%s (%s)", movie.Title, movie.Year))
	}

	suggestions, err := GetSuggestionsForPlexMovies(listOfMovies)
	if err != nil {
		log.Fatalf("Error getting suggestions: %v", err)
	}

	suggestionsList := strings.Split(suggestions, ",")

	fmt.Println(suggestionsList)

	for _, suggestion := range suggestionsList {
		RequestSpecificMovie(suggestion)
	}
}
