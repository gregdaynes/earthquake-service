package models

import (
	"database/sql"
	"time"
)

type EntryModelInterface interface {
	Insert(item Entry) (int, error)
}

type Entry struct {
	GUID       string
	Title      string
	Content    string
	Updated    *time.Time
	Published  *time.Time
	Categories string
	Elevation  int32
	Latitude   float32
	Longitude  float32
	Magnitude  float32
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
