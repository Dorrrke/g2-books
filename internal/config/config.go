package config

import (
	"cmp"
	"flag"
	"log"
	"os"
)

type Config struct {
	Host        string
	DBDsn       string
	MigratePath string
	AuthAddr    string
	Debug       bool
}

const (
	defaultDBDSN       = "postgres://postgres:6406655@localhost:5432/cource_g2?sslmode=disable"
	defaultMigratePath = "migrations"
	defaultHost        = ":8080"
	defaultAuthAddr    = "localhost:8081"
)

func ReadConfig() Config {
	var host string
	var dbDsn string
	var migratePath string
	flag.StringVar(&host, "host", defaultHost, "server host")
	flag.StringVar(&dbDsn, "db", defaultDBDSN, "data base addres")
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
	if dbDsnEnv != "" && dbDsn == defaultDBDSN {
		dbDsn = dbDsnEnv
	}
	if migratePathEnv != "" && migratePath == defaultMigratePath {
		migratePath = migratePathEnv
	}
	authAddr := cmp.Or(os.Getenv("AUTH_ADDR"), defaultAuthAddr)
	return Config{
		Host:        host,
		DBDsn:       dbDsn,
		MigratePath: migratePath,
		AuthAddr:    authAddr,
		Debug:       *debug,
	}
}
