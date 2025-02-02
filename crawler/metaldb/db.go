package metaldb

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/brucebales/metal-search/models"
	"github.com/go-sql-driver/mysql"
)

func GetDB() (*sql.DB, error) {
	// Load the CA certificate
	rootCertPool := x509.NewCertPool()

	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user:", err)
		return nil, err
	}

	filePath := filepath.Join(usr.HomeDir, "metal-cert.crt")

	pem, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert file: %v", err)
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		return nil, fmt.Errorf("failed to append CA cert to pool")
	}

	// Register the TLS configuration
	err = mysql.RegisterTLSConfig("custom", &tls.Config{
		RootCAs: rootCertPool,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register TLS config: %v", err)
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
		return nil, fmt.Errorf("failed to open MySQL connection: %v", err)
	}

	return db, nil
}

func WriteToMysql(db *sql.DB, bands []models.Band) error {
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

// Check if a band exists by ID
func BandExists(db *sql.DB, bandID int) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM bands WHERE id = ?)"
	err := db.QueryRow(query, bandID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if band exists: %v", err)
	}
	return exists, nil
}
