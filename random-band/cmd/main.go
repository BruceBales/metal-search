package main

import (
	"log"
	"net/http"
	"os"

	"github.com/brucebales/metal-search/random-band/handlers"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	http.HandleFunc("/", handlers.Home)
	http.HandleFunc("/randomBand", handlers.RandomBand)
	http.HandleFunc("/poser", handlers.Poser)
	if os.Getenv("TLS_ENABLED") == "true" {
		go func() {
			log.Fatal(http.ListenAndServe(":80", nil))
		}()
		log.Fatal(http.ListenAndServeTLS(":443", os.Getenv("TLS_CERT_PATH"), os.Getenv("TLS_KEY_PATH"), nil))
	} else {
		log.Fatal(http.ListenAndServe(":80", nil))
	}
}
