package models

import (
	"database/sql"
	"fmt"
	"time"
)

type EntryModelInterface interface {
	Insert(item Entry) (int, error)
}

type Entry struct {
	GUID       string
	Title      string
	Content    string
	Categories string
	Elevation  int32
	Latitude   float32
	Longitude  float32
	Magnitude  float32
	Updated    *time.Time
	Published  *time.Time
}

type EntryModel struct {
	DB *sql.DB
}

func (m *EntryModel) Insert(item Entry) (int, error) {
	stmt := `INSERT INTO entries (
		guid, 
		title, 
		content, 
		categories, 
		elevation, 
		latitude, 
		longitude, 
		magnitude,
		updated, 
		published
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT (guid) DO NOTHING`

	result, err := m.DB.Exec(
		stmt,
		item.GUID,
		item.Title,
		item.Content,
		item.Categories,
		item.Elevation,
		item.Latitude,
		item.Longitude,
		item.Magnitude,
		item.Updated,
		item.Published,
	)
	if err != nil {
		return 0, err
	}

	// Use the LastInsertId() method on the result to get the id of our
	// newly inserted record in the snippets table.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, nil
	}

	// The ID returned has the type int64, so we convert it to an int type
	// before returning.
	return int(id), nil
}

func (m *EntryModel) QueryWithBounds(lat1, lat2, lng1, lng2 float64) (results []Entry) {
	stmt := `SELECT 
		guid, 
		title, 
		content, 
		categories, 
		elevation, 
		latitude, 
		longitude, 
		magnitude
		 FROM entries WHERE latitude >= ? AND latitude <= ? AND longitude >= ? AND longitude <= ?`
	rows, _ := m.DB.Query(stmt, lat1, lat2, lng1, lng2)
	defer rows.Close()

	for rows.Next() {
		var e Entry

		err := rows.Scan(&e.GUID, &e.Title, &e.Content, &e.Categories, &e.Elevation, &e.Latitude, &e.Longitude, &e.Magnitude)
		if err != nil {
			fmt.Println(err)
		}

		results = append(results, e)
	}

	return results
}
