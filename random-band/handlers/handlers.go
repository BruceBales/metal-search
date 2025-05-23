package handlers

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

	"github.com/brucebales/metal-search/random-band/helpers"
	"github.com/go-sql-driver/mysql"
)

func Home(w http.ResponseWriter, r *http.Request) {

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

	tlsEnabled := os.Getenv("TLS_ENABLED") == "true"

	if tlsEnabled {
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
	}

	// Construct the DSN
	var dsn string
	if tlsEnabled {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/defaultdb?tls=custom", dbUser, dbPassword, dbHost, dbPort)
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/defaultdb", dbUser, dbPassword, dbHost, dbPort)
	}

	// Open a connection to the MySQL database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open MySQL connection: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = helpers.ReportHit(db, r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to record hit: %v", err), http.StatusInternalServerError)
		return
	}

	// Query for distinct countries
	countries, err := helpers.GetDistinctValues(db, "country")
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
					<option value="%%Thrash Metal%%">Thrash Metal</option>
                    <option value="%%Death Metal%%">Death Metal</option>
					<option value="%%Melodic Death Metal%%">Melodic Death Metal</option>
					<option value="%%Technical Death Metal%%">Technical Death Metal</option>
                    <option value="%%Black Metal%%">Black Metal</option>
					<option value="%%Melodic Black Metal%%">Melodic Black Metal</option>
					<option value="%%Folk Metal%%">Folk Metal</option>
                    <option value="%%Power Metal%%">Power Metal</option>
                    <option value="%%Doom Metal%%">Doom Metal</option>
					<option value="Metalcore">Metalcore</option>
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

func Poser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(666)
	fmt.Fprintln(w, `
	<html>
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
			.links {
				display: flex;
				justify-content: space-between;
			}

			a {
				color: grey;
			}
		</style>
	<head><title>666 Posers Not Allowed</title></head>
	<body>
	<container>
		HTTP Error Code 666
		<br>
		<h1>Posers Not Allowed</h1>
		</container>
	</body>
	</html>
	`)
}

func RandomBand(w http.ResponseWriter, r *http.Request) {
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

	tlsEnabled := os.Getenv("TLS_ENABLED") == "true"

	if tlsEnabled {
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
	}

	// Construct the DSN
	var dsn string
	if tlsEnabled {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/defaultdb?tls=custom", dbUser, dbPassword, dbHost, dbPort)
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/defaultdb", dbUser, dbPassword, dbHost, dbPort)
	}

	// Open a connection to the MySQL database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open MySQL connection: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = helpers.ReportHit(db, r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to record hit: %v", err), http.StatusInternalServerError)
		return
	}

	// Get parameters from the request
	genre := r.URL.Query().Get("genre")
	country := r.URL.Query().Get("country")

	if genre == "Metalcore" {
		Poser(w, r)
		return
	}

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
			fmt.Fprintln(w, `
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
                    .container {
                        text-align: left;
                        width: 400px; /* Fixed width for the container */
                    }
					.links {
                        display: flex;
                        justify-content: space-between;
                    }

                    a {
                        color: grey;
                    }
                </style>
			</head>`)
			fmt.Fprintln(w, `
			<div class="container">
				<a href="/">Home</a><br>
				<br>No bands found.
			</div>
			`)
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
					.links {
                        display: flex;
                        justify-content: space-between;
                    }

                    a {
                        color: grey;
                    }
                </style>
            </head>
            <body>
                <div class="container">
					<div class="links">
                        <a href="/">Home</a>
                        <a href="javascript:location.reload()">Refresh</a>
                    </div>
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
