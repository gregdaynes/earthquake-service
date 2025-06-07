package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
)

//go:embed "schema.sql"
var schemaSQL []byte

type Config struct {
	Host     string
	Port     string
	Debug    bool
	DSN      string
	Schema   string
	AtomFeed string
}

func NewConfiguration() *Config {
	debug := flag.Bool("debug", false, "Enable debug mode")
	host := flag.String("host", "127.0.0.1", "Listen on address")
	port := flag.String("port", "4000", "Listen on port")
	dsn := flag.String("dsn", "file:quakes.sqlite3", "Database connection string")
	schema := flag.String("schema", "./cmd/web/schema.sql", "Custom database schema")
	flag.Parse()

	// the schema.sql file is embedded during build
	// if the operator doesn't override the schema during run
	// the embedded schema will be used.
	var err error
	if *schema != "./cmd/web/schema.sql" {
		fmt.Println("using custom schema")

		schemaSQL, err = os.ReadFile(*schema)
		if err != nil {
			log.Fatal(err)
		}
	}

	config := Config{
		Debug:    *debug,
		Host:     *host,
		Port:     *port,
		DSN:      *dsn,
		Schema:   string(schemaSQL),
		AtomFeed: "https://www.earthquakescanada.nrcan.gc.ca/cache/earthquakes/canada-en.atom",
	}

	return &config
}
