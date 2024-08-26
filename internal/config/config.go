package config

import (
	"flag"
	"log"
	"os"
)

type Config struct {
	Host        string
	DbDsn       string
	MigratePath string
	Debug       bool
}

const (
	defaultDbDSN       = "postgres://postgres:6406655@localhost:5432/cource_g2?sslmode=disable"
	defaultMigratePath = "migrations"
	defaultHost        = ":8080"
)

func ReadConfig() Config {
	var host string
	var dbDsn string
	var migratePath string
	flag.StringVar(&host, "host", defaultHost, "server host")
	flag.StringVar(&dbDsn, "db", defaultDbDSN, "data base addres")
	flag.StringVar(&migratePath, "m", defaultMigratePath, "path to migrations")
	debug := flag.Bool("debug", false, "enable debug logging level")
	flag.Parse()

	hostEnv := os.Getenv("SERVER_HOS")
	dbDsnEnv := os.Getenv("DB_DSN")
	migratePathEnv := os.Getenv("MIGRATE_PATH")
	log.Println(hostEnv)
	if hostEnv != "" && host == defaultHost {
		host = hostEnv
	}
	if dbDsnEnv != "" && dbDsn == defaultDbDSN {
		dbDsn = dbDsnEnv
	}
	if migratePathEnv != "" && migratePath == defaultMigratePath {
		migratePath = migratePathEnv
	}
	return Config{
		Host:        host,
		DbDsn:       dbDsn,
		MigratePath: migratePath,
		Debug:       *debug,
	}
}
