package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// RadarrMovie represents a movie in Radarr
type RadarrMovie struct {
	Title       string `json:"title"`
	Year        int    `json:"year"`
	TMDBID      int    `json:"tmdbId"`
	TitleSlug   string `json:"titleSlug"`
	Monitored   bool   `json:"monitored"`
	HasFile     bool   `json:"hasFile"`
	IsAvailable bool   `json:"isAvailable"`
}

// RadarrSearchResult represents search results from Radarr
type RadarrSearchResult struct {
	Title       string `json:"title"`
	Year        int    `json:"year"`
	TMDBID      int    `json:"tmdbId"`
	TitleSlug   string `json:"titleSlug"`
	Monitored   bool   `json:"monitored"`
	HasFile     bool   `json:"hasFile"`
	IsAvailable bool   `json:"isAvailable"`
}

// RadarrClient represents a Radarr API client
type RadarrClient struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

// NewRadarrClient creates a new Radarr API client
func NewRadarrClient(baseURL, apiKey string) *RadarrClient {
	return &RadarrClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Client:  &http.Client{},
	}
}

// SearchMovies searches for movies in Radarr
func (r *RadarrClient) SearchMovies(query string) ([]RadarrSearchResult, error) {
	url := fmt.Sprintf("%s/api/v3/movie/lookup?term=%s", r.BaseURL, url.QueryEscape(query))

	// Debug: Print the URL being called
	fmt.Printf("🔍 Searching Radarr URL: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("X-Api-Key", r.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("non-200 response: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var results []RadarrSearchResult
	err = json.Unmarshal(body, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return results, nil
}

// AddMovie adds a movie to Radarr for monitoring and downloading
func (r *RadarrClient) AddMovie(movie RadarrSearchResult, qualityProfileID int, rootFolderPath string) error {
	addMovieRequest := map[string]interface{}{
		"title":            movie.Title,
		"titleSlug":        movie.TitleSlug,
		"tmdbId":           movie.TMDBID,
		"year":             movie.Year,
		"monitored":        true,
		"qualityProfileId": qualityProfileID,
		"rootFolderPath":   rootFolderPath,
		"addOptions": map[string]interface{}{
			"searchForMovie": true,
		},
	}

	jsonData, err := json.Marshal(addMovieRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	url := fmt.Sprintf("%s/api/v3/movie", r.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("X-Api-Key", r.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("non-201 response: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetMovies gets all movies from Radarr
func (r *RadarrClient) GetMovies() ([]RadarrMovie, error) {
	url := fmt.Sprintf("%s/api/v3/movie", r.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("X-Api-Key", r.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("non-200 response: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var movies []RadarrMovie
	err = json.Unmarshal(body, &movies)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return movies, nil
}

// getQualityProfileIDByName returns the ID of the quality profile matching the given name (case-insensitive)
func (r *RadarrClient) getQualityProfileIDByName(name string) (int, error) {
	profiles, err := r.GetQualityProfiles()
	if err != nil {
		return 0, err
	}
	for _, profile := range profiles {
		if pname, ok := profile["name"].(string); ok {
			if strings.EqualFold(pname, name) {
				if id, ok := profile["id"].(float64); ok {
					return int(id), nil
				}
			}
		}
	}
	return 0, fmt.Errorf("quality profile '%s' not found", name)
}

// RequestMovieByName searches for a movie by name and requests it to be added to Radarr
func (r *RadarrClient) RequestMovieByName(movieName string, qualityProfileID int, rootFolderPath string) error {
	// First, search for the movie
	results, err := r.SearchMovies(movieName)
	if err != nil {
		return fmt.Errorf("failed to search for movie '%s': %v", movieName, err)
	}

	if len(results) == 0 {
		return fmt.Errorf("no movies found for '%s'", movieName)
	}

	// Use the first (best) result
	bestMatch := results[0]
	fmt.Printf("Found movie: %s (%d) - TMDB ID: %d\n", bestMatch.Title, bestMatch.Year, bestMatch.TMDBID)

	// Check if the movie is already in Radarr
	existingMovies, err := r.GetMovies()
	if err != nil {
		return fmt.Errorf("failed to check existing movies: %v", err)
	}

	// Check if movie already exists
	for _, movie := range existingMovies {
		if movie.TMDBID == bestMatch.TMDBID {
			return fmt.Errorf("movie '%s' is already in your Radarr library", bestMatch.Title)
		}
	}

	// Add the movie to Radarr
	err = r.AddMovie(bestMatch, qualityProfileID, rootFolderPath)
	if err != nil {
		return fmt.Errorf("failed to add movie '%s' to Radarr: %v", bestMatch.Title, err)
	}

	fmt.Printf("Successfully requested '%s' (%d) to be added to Radarr\n", bestMatch.Title, bestMatch.Year)
	return nil
}

// RequestMovieByNameWithDefaults is a convenience function that uses env or default settings
func (r *RadarrClient) RequestMovieByNameWithDefaults(movieName string) error {
	// Check for env override
	profileName := os.Getenv("RADARR_QUALITY_PROFILE")
	var qualityProfileID int
	var err error
	if profileName != "" {
		qualityProfileID, err = r.getQualityProfileIDByName(profileName)
		if err != nil {
			fmt.Printf("Warning: %v. Falling back to default profile ID 1.\n", err)
			qualityProfileID = 1
		}
	} else {
		qualityProfileID = 1
	}
	// Default root folder path (you'll need to set this to your actual movies folder)
	rootFolderPath := "/movies" // Change this to your actual movies folder path

	return r.RequestMovieByName(movieName, qualityProfileID, rootFolderPath)
}

// GetQualityProfiles gets all quality profiles from Radarr
func (r *RadarrClient) GetQualityProfiles() ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v3/qualityprofile", r.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("X-Api-Key", r.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("non-200 response: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var profiles []map[string]interface{}
	err = json.Unmarshal(body, &profiles)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return profiles, nil
}

// GetRootFolders gets all root folders from Radarr
func (r *RadarrClient) GetRootFolders() ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v3/rootfolder", r.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("X-Api-Key", r.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("non-200 response: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var folders []map[string]interface{}
	err = json.Unmarshal(body, &folders)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return folders, nil
}

// RequestSpecificMovie is a standalone function to request a movie by name
func RequestSpecificMovie(movieName string) {
	// Get configuration from environment variables
	radarrURL := os.Getenv("RADARR_URL")
	radarrAPIKey := os.Getenv("RADARR_API_KEY")

	// Validate required environment variables
	if radarrURL == "" || radarrAPIKey == "" {
		fmt.Println("Missing required Radarr environment variables. Please check your .env file.")
		return
	}

	client := NewRadarrClient(radarrURL, radarrAPIKey)

	fmt.Printf("🎬 Requesting movie: %s\n", movieName)

	err := client.RequestMovieByNameWithDefaults(movieName)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Successfully requested '%s' to be added to Radarr\n", movieName)
	}
}
