package models

import (
	"database/sql"
	"fmt"
	"time"
)

type EntryModelInterface interface {
	Insert() (int, error)
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

func Insert(item Entry) (int, error) {
	fmt.Println(item)

	return 0, nil
}
