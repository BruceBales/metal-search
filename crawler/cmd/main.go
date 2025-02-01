package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/brucebales/metal-search/models"
	"golang.org/x/net/html"
)

// Extract data from a given URL
func extractData(url string) (string, string, string, string, string, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", "", "", "", "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "MyCustomUserAgent/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", "", "", "", fmt.Errorf("failed to send request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", "", "", "", fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	var name, genre, location, status, formedIn, themes string
	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return name, genre, location, status, formedIn, themes, nil
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

		}
	}
}

func main() {

	numBands := 20

	bands := []models.Band{}

	for i := 1; i <= numBands; i++ {
		url := fmt.Sprintf("https://www.metal-archives.com/bands/scrape/%d", i)
		name, genre, location, status, formedIn, themes, err := extractData(url)
		if err != nil {
			fmt.Printf("Failed to extract data: %v\n", err)
			continue
		}

		band := models.Band{}

		band.ID = i
		band.Name = name
		band.Genre = genre
		band.Location = location
		band.Status = status
		band.FormedIn, err = strconv.Atoi(formedIn)
		if err != nil {
			fmt.Printf("Failed to convert formedIn to int: %v\n", err)
			continue
		}
		band.Themes = themes

		bands = append(bands, band)

	}

	bandsJSON, err := json.Marshal(bands)
	if err != nil {
		fmt.Printf("Failed to marshal bands: %v\n", err)
	}

	fmt.Println(string(bandsJSON))

}
