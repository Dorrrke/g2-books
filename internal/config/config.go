package config

import (
	"flag"
)

type Config struct {
	Host  string
	Debug bool
}

func ReadConfig() Config {
	var host string
	flag.StringVar(&host, "host", ":8080", "server host")
	debug := flag.Bool("debug", false, "enable debug logging level")
	flag.Parse()

	return Config{
		Host:  host,
		Debug: *debug,
	}
}
