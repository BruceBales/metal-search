package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/randomBand", randomBandHandler)
	if os.Getenv("TLS_ENABLED") == "true" {
		go func() {
			log.Fatal(http.ListenAndServe(":80", nil))
		}()
		log.Fatal(http.ListenAndServeTLS(":443", os.Getenv("TLS_CERT_PATH"), os.Getenv("TLS_KEY_PATH"), nil))
	} else {
		log.Fatal(http.ListenAndServe(":80", nil))
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Read environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	// Load the CA certificate
	rootCertPool := x509.NewCertPool()

	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user:", err)
		return
	}

	filePath := filepath.Join(usr.HomeDir, "metal-cert.crt")
	pem, err := ioutil.ReadFile(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read CA cert file: %v", err), http.StatusInternalServerError)
		return
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		http.Error(w, "Failed to append CA cert to pool", http.StatusInternalServerError)
		return
	}

	// Register the TLS configuration
	err = mysql.RegisterTLSConfig("custom", &tls.Config{
		RootCAs: rootCertPool,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to register TLS config: %v", err), http.StatusInternalServerError)
		return
	}

	// Construct the DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/defaultdb?tls=custom", dbUser, dbPassword, dbHost, dbPort)

	// Open a connection to the MySQL database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open MySQL connection: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query for distinct countries
	countries, err := getDistinctValues(db, "country")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to query countries: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate the HTML form with dropdown options
	fmt.Fprintf(w, `
        <!DOCTYPE html>
        <html>
        <head>
            <title>Random Metal Band</title>
			<style>
				body {
					background-color: black;
					color: lightgrey;
					font-family: Arial, sans-serif;
					display: flex;
					justify-content: center;
					align-items: center;
					height: 100vh;
					margin: 0;
				}
                select, input[type="submit"] {
                    background-color: #333;
                    color: lightgrey;
                    border: 1px solid lightgrey;
                    padding: 5px;
                    margin: 5px;
                }
                select {
                    width: 200px;
                }
            </style>
        </head>
        <body>
			<div class="container">
            <h1>Random Metal Band</h1>
			<br>
            <form action="/randomBand" method="get">
                <label for="genre">Genre:</label>
                <select id="genre" name="genre">
                    <option value="">Any</option>
                    <option value="%%Progressive%%">Progressive Metal</option>
                    <option value="%%Death Metal%%">Death Metal</option>
					<option value="%%Melodic Death Metal%%">Melodic Death Metal</option>
                    <option value="%%Black Metal%%">Black Metal</option>
					<option value="%%Folk Metal%%">Folk Metal</option>
                    <option value="%%Power Metal%%">Power Metal</option>
                    <!-- Add more genres as needed -->
                </select><br><br>
                <label for="country">Country:</label>
                <select id="country" name="country">
                    <option value="">Any</option>
    `)
	for _, country := range countries {
		fmt.Fprintf(w, `<option value="%s">%s</option>`, country, country)
	}
	fmt.Fprintf(w, `
                </select><br><br>
                <input type="submit" value="Get Random Band">
            </form>
			</div>
        </body>
        </html>
    `)
}

func randomBandHandler(w http.ResponseWriter, r *http.Request) {
	// Read environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	// Load the CA certificate
	rootCertPool := x509.NewCertPool()

	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user:", err)
		return
	}

	filePath := filepath.Join(usr.HomeDir, "metal-cert.crt")

	pem, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("failed to read CA cert file: %v", err)
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		fmt.Printf("failed to append CA cert to pool")
	}

	// Register the TLS configuration
	err = mysql.RegisterTLSConfig("custom", &tls.Config{
		RootCAs: rootCertPool,
	})
	if err != nil {
		log.Fatalf("Failed to register TLS config: %v", err)
	}

	// Construct the DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/defaultdb?tls=custom", dbUser, dbPassword, dbHost, dbPort)

	// Open a connection to the MySQL database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open MySQL connection: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Get parameters from the request
	genre := r.URL.Query().Get("genre")
	country := r.URL.Query().Get("country")

	args := []any{}

	// Construct the query
	query := "SELECT id, spotify_link, name, country, location, genre FROM bands WHERE spotify_link != ''"
	if genre != "" {
		query += " AND genre like ?"
		args = append(args, genre)
	}
	if country != "" {
		query += " AND country = ?"
		args = append(args, country)
	}
	query += " ORDER BY RAND() LIMIT 1"

	var spotifyLink string

	var id int
	var name string
	var countryResult string
	var location string
	var genreResult string

	err = db.QueryRow(query, args...).Scan(&id, &spotifyLink, &name, &countryResult, &location, &genreResult)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Fprintln(w, `<a href="/">Home</a><br>`)
			fmt.Fprintln(w, "No bands found.")
			return
		}
		http.Error(w, fmt.Sprintf("Failed to query database: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	metalArchivesLink := fmt.Sprintf("https://www.metal-archives.com/bands/%s/%d", name, id)

	if spotifyLink == "" {
		fmt.Fprintln(w, "No bands found.")
	} else {
		fmt.Fprintf(w, `
            <!DOCTYPE html>
            <html>
            <head>
                <meta charset="UTF-8">
                <title>Random Band</title>
                <style>
                    body {
                        background-color: black;
                        color: lightgrey;
                        font-family: Arial, sans-serif;
                        display: flex;
                        justify-content: center;
                        align-items: center;
                        height: 100vh;
                        margin: 0;
                    }
                    .container {
                        text-align: left;
                        width: 400px; /* Fixed width for the container */
                    }
                    a {
                        color: grey;
                    }
                </style>
            </head>
            <body>
                <div class="container">
					<a href="/">Home</a><br>
                    <p><strong>Band Name:</strong> %s</p>
                    <p><strong>Country:</strong> %s</p>
                    <p><strong>Location:</strong> %s</p>
                    <p><strong>Genre:</strong> %s</p>
					<p><string>Links:</strong></p>
                    <a href="%s">Spotify</a></p>
					<a href="%s">Metal Archives</a></p>
                </div>
            </body>
            </html>
        `, name, countryResult, location, genreResult, spotifyLink, metalArchivesLink)

	}
}

func getDistinctValues(db *sql.DB, column string) ([]string, error) {
	query := fmt.Sprintf("SELECT DISTINCT %s FROM bands WHERE %s IS NOT NULL AND %s != '' ORDER BY %s", column, column, column, column)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return values, nil
}
