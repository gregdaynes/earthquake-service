package main

import "flag"

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
	schema := flag.String("schema", "./schema.sql", "Database schema")
	flag.Parse()

	config := Config{
		Debug:    *debug,
		Host:     *host,
		Port:     *port,
		DSN:      *dsn,
		Schema:   *schema,
		AtomFeed: "https://www.earthquakescanada.nrcan.gc.ca/cache/earthquakes/canada-en.atom",
	}

	return &config
}
