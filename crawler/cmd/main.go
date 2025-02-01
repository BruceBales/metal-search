package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/brucebales/metal-search/models"
	"golang.org/x/net/html"
)

// Extract data from a given URL
func extractData(url string) (string, string, string, string, string, string, string, string, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", "", "", "", "", "", "", "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "MyCustomUserAgent/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", "", "", "", "", "", "", fmt.Errorf("failed to send request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", "", "", "", "", "", "", fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	var name, genre, country, location, status, formedIn, themes, yearsActive, label string
	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return name, genre, country, location, status, formedIn, themes, yearsActive, label, nil
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()

			if t.Data == "h1" {
				for _, a := range t.Attr {
					if a.Key == "class" && a.Val == "band_name" {
						z.Next()
						z.Next()
						name = z.Token().Data
					}
				}
			}
			if t.Data == "dt" {
				z.Next()
				if string(z.Raw()) == "Location:" {
					for i := 0; i < 4; i++ {
						z.Next()
					}
					location = z.Token().Data
				}
			}
			if t.Data == "dt" {
				if string(z.Raw()) == "Genre:" {
					for i := 0; i < 4; i++ {
						z.Next()
					}
					genre = z.Token().Data
				}
			}
			if t.Data == "dt" {
				if string(z.Raw()) == "Status:" {
					for i := 0; i < 4; i++ {
						z.Next()
					}
					status = z.Token().Data
				}
			}
			if t.Data == "dt" {
				if string(z.Raw()) == "Formed in:" {
					for i := 0; i < 4; i++ {
						z.Next()
					}
					formedIn = z.Token().Data
				}
			}
			if t.Data == "dt" {
				if string(z.Raw()) == "Themes:" {
					for i := 0; i < 4; i++ {
						z.Next()
					}
					themes = z.Token().Data
				}
			}
			if t.Data == "dt" {
				if string(z.Raw()) == "Country of origin:" {
					for i := 0; i < 5; i++ {
						z.Next()
					}
					country = z.Token().Data
				}
			}
			if t.Data == "dt" {
				if string(z.Raw()) == "Years active:" {
					for i := 0; i < 4; i++ {
						z.Next()
					}
					// Years active is currently not readable. I will add a parsing function for it later
					yearsActive = "N/A"
				}
			}
			if t.Data == "dt" {
				if string(z.Raw()) == "Current label:" {
					for i := 0; i < 4; i++ {
						z.Next()
					}
					label = z.Token().Data
				}
			}
		}
	}
}

func getSpotifyURL(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "MyCustomUserAgent/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	var spotifyURL string
	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return spotifyURL, err
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			if t.Data == "a" {
				for _, a := range t.Attr {
					if a.Key == "title" && a.Val == "Go to: Spotify" {
						fmt.Println(t)
						for _, a := range t.Attr {
							if a.Key == "href" {
								fmt.Println("Spotify URL: ", a.Val)
								spotifyURL = a.Val
								return spotifyURL, nil
							}
						}
					}
				}
			}
		}
	}
}

func main() {

	numBands := 1000

	batchSize := 100
	batchCount := numBands / batchSize

	currentIndex := 1

	for i := 0; i < batchCount; i++ {
		bands := []models.Band{}
		for j := currentIndex; j <= currentIndex+batchSize; j++ {
			fmt.Print("Scraping band: ", j, "\n")
			url := fmt.Sprintf("https://www.metal-archives.com/bands/scrape/%d", j)
			name, genre, country, location, status, formedIn, themes, yearsActive, label, err := extractData(url)
			if err != nil {
				fmt.Printf("Failed to extract data: %v\n", err)
				continue
			}

			spotifyURL, err := getSpotifyURL(fmt.Sprintf("https://www.metal-archives.com/link/ajax-list/type/band/id/%d", j))
			if err != nil {
				fmt.Printf("Failed to get Spotify URL: %v\n", err)
			}
			band := models.Band{}

			// fmt.Println("Spotify URL: ", spotifyURL)

			if spotifyURL != "" {
				band.Links.Spotify = spotifyURL
			}
			band.ID = j
			band.Name = name
			band.Genre = genre
			band.Country = country
			band.Location = location
			band.Status = status
			band.FormedIn, err = strconv.Atoi(formedIn)
			band.YearsActive = yearsActive
			band.Label = label
			if err != nil {
				fmt.Printf("Failed to convert formedIn to int: %v\n", err)
				continue
			}
			band.Themes = themes

			fmt.Printf("Band: %v\n", band)

			bands = append(bands, band)

			time.Sleep(1 * time.Second)

		}
		currentIndex += batchSize
		fmt.Println("Writing bands to file")
		bandsJSON, err := json.Marshal(bands)
		if err != nil {
			fmt.Printf("Failed to marshal bands: %v\n", err)
		}

		os.WriteFile(fmt.Sprintf("data/bands-%d.json", i), bandsJSON, 0644)

	}

}
