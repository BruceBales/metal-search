package models

type Band struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Country     string  `json:"country"`
	Location    string  `json:"location"`
	FormedIn    int     `json:"formed_in"`
	Status      string  `json:"status"`
	YearsActive string  `json:"years_active"`
	Genre       string  `json:"genre"`
	Themes      string  `json:"themes"`
	Label       string  `json:"label"`
	BandCover   string  `json:"band_cover"`
	Albums      []Album `json:"albums"`
	Links       string  `json:"links"`
}

type Links struct {
	Spotify string `json:"spotify"`
}

type Album struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Year int    `json:"year"`
	Link string `json:"link"`
}
