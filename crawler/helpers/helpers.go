package helpers

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/html"
)

func ExtractData(url string) (string, string, string, string, string, string, string, string, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", "", "", "", "", "", "", "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "MyCustomUserAgent/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode == http.StatusTooManyRequests {
		fmt.Println("!!! RATE LIMITED !!!")
		fmt.Println("Backing off for 5 seconds")
		time.Sleep(5 * time.Second)
		return ExtractData(url)
	}
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
					for i := 0; i < 5; i++ {
						z.Next()
					}
					label = z.Token().Data
				}
			}
		}
	}
}

func GetSpotifyURL(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "MyCustomUserAgent/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
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
						for _, a := range t.Attr {
							if a.Key == "href" {
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
