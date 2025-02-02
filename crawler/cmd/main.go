package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"

	"github.com/brucebales/metal-search/crawler/helpers"
	"github.com/brucebales/metal-search/models"
)

func main() {

	numBands := 3540558574

	batchSize := 10
	batchCount := numBands / batchSize

	currentIndex := 1

	for i := 0; i < batchCount; i++ {
		bands := []models.Band{}
		for j := currentIndex; j <= currentIndex+batchSize; j++ {
			fmt.Print("Scraping band: ", j, "\n")
			url := fmt.Sprintf("https://www.metal-archives.com/bands/scrape/%d", j)
			name, genre, country, location, status, formedIn, themes, yearsActive, label, err := helpers.ExtractData(url)
			if err != nil {
				fmt.Printf("Failed to extract data: %v\n", err)
				continue
			}

			spotifyURL, err := helpers.GetSpotifyURL(fmt.Sprintf("https://www.metal-archives.com/link/ajax-list/type/band/id/%d", j))
			if err != nil {
				fmt.Printf("Failed to get Spotify URL: %v\n", err)
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
			band.FormedIn, err = strconv.Atoi(formedIn)
			band.YearsActive = yearsActive
			band.Label = label
			if err != nil {
				fmt.Printf("Failed to convert formedIn to int: %v\n", err)
				continue
			}
			band.Themes = themes

			bands = append(bands, band)

			time.Sleep(1 * time.Second)

		}
		currentIndex += batchSize

		err := writeToMysql(bands)
		if err != nil {
			fmt.Printf("Failed to write to MySQL: %v\n", err)
		}

	}

}

func writeToMysql(bands []models.Band) error {
	// Load the CA certificate
	rootCertPool := x509.NewCertPool()

	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user:", err)
		return err
	}

	filePath := filepath.Join(usr.HomeDir, "metal-cert.crt")

	pem, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read CA cert file: %v", err)
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		return fmt.Errorf("failed to append CA cert to pool")
	}

	// Register the TLS configuration
	err = mysql.RegisterTLSConfig("custom", &tls.Config{
		RootCAs: rootCertPool,
	})
	if err != nil {
		return fmt.Errorf("failed to register TLS config: %v", err)
	}

	// Read environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	// Construct the DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/defaultdb?tls=custom", dbUser, dbPassword, dbHost, dbPort)

	// Open a connection to the MySQL database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %v", err)
	}
	defer db.Close()

	// Insert the data into the bands table
	for _, band := range bands {
		_, err := db.Exec(`
            INSERT INTO bands (id, name, country, location, formed_in, status, years_active, genre, themes, label, band_cover, spotify_link)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            ON DUPLICATE KEY UPDATE
            name = VALUES(name),
            country = VALUES(country),
            location = VALUES(location),
            formed_in = VALUES(formed_in),
            status = VALUES(status),
            years_active = VALUES(years_active),
            genre = VALUES(genre),
            themes = VALUES(themes),
            label = VALUES(label),
            band_cover = VALUES(band_cover),
            spotify_link = VALUES(spotify_link)
        `, band.ID, band.Name, band.Country, band.Location, band.FormedIn, band.Status, band.YearsActive, band.Genre, band.Themes, band.Label, band.BandCover, band.SpotifyLink)
		if err != nil {
			return fmt.Errorf("failed to insert band into MySQL: %v", err)
		}
	}

	return nil
}
