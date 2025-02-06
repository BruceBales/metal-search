package helpers

import (
	"database/sql"
	"fmt"
	"net/http"
)

func GetDistinctValues(db *sql.DB, column string) ([]string, error) {
	query := fmt.Sprintf("SELECT DISTINCT %s FROM bands WHERE %s IS NOT NULL AND %s != '' AND spotify_link != '' ORDER BY %s", column, column, column, column)
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

func ReportHit(db *sql.DB, r *http.Request) error {
	ipAddress := r.RemoteAddr
	path := r.URL.Path
	params := r.URL.RawQuery

	userAgent := r.UserAgent()

	query := `
        INSERT INTO hits (ip_address, user_agent, path, params)
        VALUES (?, ?, ?, ?)
    `

	_, err := db.Exec(query, ipAddress, userAgent, path, params)
	if err != nil {
		return fmt.Errorf("failed to record hit: %v", err)
	}

	return nil
}
