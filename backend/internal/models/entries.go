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
	Time       *time.Time
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
		published,
		time
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) 
	ON CONFLICT (guid) 
	DO UPDATE SET 
		latitude=?, 
		longitude=?, 
		elevation=?, 
		updated=?, 
		magnitude=?, 
		content=?, 
		time=?;
	`

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
		item.Time,
		// -- on updat values
		item.Latitude,
		item.Longitude,
		item.Elevation,
		item.Updated,
		item.Magnitude,
		item.Content,
		item.Time,
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
	stmt := `
		SELECT 
			guid, 
			title, 
			content, 
			categories, 
			time,
			elevation, 
			latitude, 
			longitude, 
			magnitude
		 FROM entries WHERE latitude >= ? AND latitude <= ? AND longitude >= ? AND longitude <= ?
		 ORDER BY time DESC
	`
	rows, _ := m.DB.Query(stmt, lat1, lat2, lng1, lng2)
	defer rows.Close()

	for rows.Next() {
		var e Entry

		err := rows.Scan(&e.GUID, &e.Title, &e.Content, &e.Categories, &e.Time, &e.Elevation, &e.Latitude, &e.Longitude, &e.Magnitude)
		if err != nil {
			fmt.Println(err)
		}

		results = append(results, e)
	}

	return results
}
