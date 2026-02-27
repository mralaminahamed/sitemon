package history

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type CheckRecord struct {
	ID           int       `json:"id"`
	URL          string    `json:"url"`
	Status       string    `json:"status"`
	StatusCode   int       `json:"status_code"`
	ResponseTime int64     `json:"response_time_ms"`
	Timestamp    time.Time `json:"timestamp"`
	Error        string    `json:"error,omitempty"`
}

type RequestRecord struct {
	ID           int       `json:"id"`
	URL          string    `json:"url"`
	Method       string    `json:"method"`
	Status       string    `json:"status"`
	StatusCode   int       `json:"status_code"`
	ResponseTime int64     `json:"response_time_ms"`
	Timestamp    time.Time `json:"timestamp"`
}

type Database struct {
	db *sql.DB
}

func NewDatabase(path string) (*Database, error) {
	if path == "" {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, ".portman", "history.db")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	d := &Database{db: db}

	if err := d.createTables(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Database) createTables() error {
	checks := `
		CREATE TABLE IF NOT EXISTS checks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT NOT NULL,
			status TEXT NOT NULL,
			status_code INTEGER,
			response_time_ms INTEGER,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			error TEXT
		);
	`

	requests := `
		CREATE TABLE IF NOT EXISTS requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT NOT NULL,
			method TEXT NOT NULL,
			status TEXT NOT NULL,
			status_code INTEGER,
			response_time_ms INTEGER,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := d.db.Exec(checks); err != nil {
		return fmt.Errorf("failed to create checks table: %w", err)
	}

	if _, err := d.db.Exec(requests); err != nil {
		return fmt.Errorf("failed to create requests table: %w", err)
	}

	return nil
}

func (d *Database) SaveCheck(record CheckRecord) error {
	query := `INSERT INTO checks (url, status, status_code, response_time_ms, timestamp, error) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := d.db.Exec(query, record.URL, record.Status, record.StatusCode, record.ResponseTime, record.Timestamp, record.Error)
	return err
}

func (d *Database) SaveRequest(record RequestRecord) error {
	query := `INSERT INTO requests (url, method, status, status_code, response_time_ms, timestamp) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := d.db.Exec(query, record.URL, record.Method, record.Status, record.StatusCode, record.ResponseTime, record.Timestamp)
	return err
}

func (d *Database) GetChecks(limit int, from, to *time.Time) ([]CheckRecord, error) {
	query := "SELECT id, url, status, status_code, response_time_ms, timestamp, error FROM checks"
	args := []interface{}{}

	if from != nil || to != nil {
		query += " WHERE"
		if from != nil {
			query += " timestamp >= ?"
			args = append(args, from)
			if to != nil {
				query += " AND"
			}
		}
		if to != nil {
			query += " timestamp <= ?"
			args = append(args, to)
		}
	}

	query += " ORDER BY timestamp DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []CheckRecord
	for rows.Next() {
		var r CheckRecord
		var timestamp []byte
		err := rows.Scan(&r.ID, &r.URL, &r.Status, &r.StatusCode, &r.ResponseTime, &timestamp, &r.Error)
		if err != nil {
			return nil, err
		}
		r.Timestamp, _ = time.Parse("2006-01-02 15:04:05", string(timestamp))
		records = append(records, r)
	}

	return records, nil
}

func (d *Database) GetRequests(limit int, from, to *time.Time) ([]RequestRecord, error) {
	query := "SELECT id, url, method, status, status_code, response_time_ms, timestamp FROM requests"
	args := []interface{}{}

	if from != nil || to != nil {
		query += " WHERE"
		if from != nil {
			query += " timestamp >= ?"
			args = append(args, from)
			if to != nil {
				query += " AND"
			}
		}
		if to != nil {
			query += " timestamp <= ?"
			args = append(args, to)
		}
	}

	query += " ORDER BY timestamp DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []RequestRecord
	for rows.Next() {
		var r RequestRecord
		var timestamp []byte
		err := rows.Scan(&r.ID, &r.URL, &r.Method, &r.Status, &r.StatusCode, &r.ResponseTime, &timestamp)
		if err != nil {
			return nil, err
		}
		r.Timestamp, _ = time.Parse("2006-01-02 15:04:05", string(timestamp))
		records = append(records, r)
	}

	return records, nil
}

func (d *Database) GetStats(url string) (map[string]interface{}, error) {
	var totalChecks, successful, failed int64
	var avgResponseTime float64

	checkQuery := `SELECT 
		COUNT(*) as total,
		COUNT(CASE WHEN status = 'UP' THEN 1 END) as successful,
		COUNT(CASE WHEN status != 'UP' THEN 1 END) as failed,
		AVG(response_time_ms) as avg_response
	FROM checks WHERE url = ?`

	err := d.db.QueryRow(checkQuery, url).Scan(&totalChecks, &successful, &failed, &avgResponseTime)
	if err != nil {
		return nil, err
	}

	uptime := 0.0
	if totalChecks > 0 {
		uptime = float64(successful) / float64(totalChecks) * 100
	}

	return map[string]interface{}{
		"url":               url,
		"total_checks":     totalChecks,
		"successful":       successful,
		"failed":           failed,
		"uptime_percentage": uptime,
		"avg_response_ms":  avgResponseTime,
	}, nil
}

func (d *Database) ExportJSON() ([]byte, error) {
	checks, err := d.GetChecks(10000, nil, nil)
	if err != nil {
		return nil, err
	}

	requests, err := d.GetRequests(10000, nil, nil)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"checks":     checks,
		"requests":   requests,
	}

	return json.MarshalIndent(data, "", "  ")
}

func (d *Database) Close() error {
	return d.db.Close()
}
