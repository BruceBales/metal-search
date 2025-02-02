package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/brucebales/metal-search/crawler/helpers"
	"github.com/brucebales/metal-search/crawler/metaldb"
	"github.com/brucebales/metal-search/models"
)

func main() {

	db, err := metaldb.GetDB()
	if err != nil {
		fmt.Printf("Failed to get DB: %v\n", err)
	}
	defer db.Close()

	numBands := 3540558574

	batchSize := 10
	batchCount := numBands / batchSize

	currentIndex := 1

	updateMode := os.Getenv("UPDATE_MODE") == "true"

	for i := 0; i < batchCount; i++ {
		bands := []models.Band{}
		for j := currentIndex; j <= currentIndex+batchSize; j++ {
			if !updateMode {
				exists, err := metaldb.BandExists(db, j)
				if err != nil {
					fmt.Printf("Failed to check if band exists: %v\n", err)
				}
				if exists {
					fmt.Printf("Band with ID %d already exists\n", j)
					continue
				}
			}

			fmt.Print("Scraping band: ", j, "\n")
			url := fmt.Sprintf("https://www.metal-archives.com/bands/scrape/%d", j)
			name, genre, country, location, status, formedIn, themes, yearsActive, label, err := helpers.ExtractData(url)
			if err != nil {
				if err.Error() == "status code 429" {
					fmt.Println("!!! RATE LIMITED !!!")
					fmt.Println("Backing off for 5 seconds")
					time.Sleep(5 * time.Second)
				} else {
					fmt.Printf("Failed to extract data: %v\n", err)
				}
				continue
			}

			spotifyURL, err := helpers.GetSpotifyURL(fmt.Sprintf("https://www.metal-archives.com/link/ajax-list/type/band/id/%d", j))
			if err != nil {
				if err.Error() != "status code 404" {
					fmt.Printf("Failed to get Spotify URL: %v\n", err)
				}
			}
			band := models.Band{}

			// fmt.Println("Spotify URL: ", spotifyURL)

			if spotifyURL != "" {
				band.SpotifyLink = spotifyURL
			}
			band.ID = j
			band.Name = name
			band.Genre = genre
			band.Country = country
			band.Location = location
			band.Status = status
			if formedIn != "" && formedIn != "N/A" {
				band.FormedIn, err = strconv.Atoi(formedIn)
				if err != nil {
					fmt.Printf("Failed to convert formedIn to int: %v\n", err)
				}
			}
			band.YearsActive = yearsActive
			band.Label = label
			band.Themes = themes

			bands = append(bands, band)

			time.Sleep(1 * time.Second)

		}
		currentIndex += batchSize

		err := metaldb.WriteToMysql(db, bands)
		if err != nil {
			fmt.Printf("Failed to write to MySQL: %v\n", err)
		}

	}

}
