package main

import "flag"

type Config struct {
	Host     string
	Port     string
	Debug    bool
	AtomFeed string
}

func NewConfiguration() *Config {
	debug := flag.Bool("debug", false, "Enable debug mode")
	host := flag.String("host", "127.0.0.1", "Listen on address")
	port := flag.String("port", "4000", "Listen on port")
	flag.Parse()

	config := Config{
		Debug:    *debug,
		Host:     *host,
		Port:     *port,
		AtomFeed: "https://www.earthquakescanada.nrcan.gc.ca/cache/earthquakes/canada-en.atom",
	}

	return &config
}
