package main

import "flag"

type Config struct {
	Host  string
	Port  string
	Debug bool
}

func NewConfiguration() *Config {
	debug := flag.Bool("debug", false, "Enable debug mode")
	host := flag.String("host", "127.0.0.1", "Listen on address")
	port := flag.String("port", "4000", "Listen on port")
	flag.Parse()

	config := Config{
		Debug: *debug,
		Host:  *host,
		Port:  *port,
	}

	return &config
}
