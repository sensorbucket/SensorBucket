package main

import (
	"flag"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/internal/env"
)

func main() {
	cleanPtr := flag.Bool("clean", false, "indicates whether to delete all expired data from the tracing and sensorbucket database")
	warnPtr := flag.Bool("warn", false, "indicates whether to alert users by email that expired data will soon be deleted")
	flag.Parse()

	if *cleanPtr {
		err := runCleanup()
		if err != nil {
			panic(err)
		}
	}
	if *warnPtr {
		err := runWarn()
		if err != nil {
			panic(err)
		}
	}

	if !*cleanPtr && !*warnPtr {
		log.Println("[Warning] no clean or warn parameter given. not running.")
	}
}

func runCleanup() error {
	var (
		// Database
		DB_DSN_SENSORBUCKET = env.Must("DB_DSN_SENSORBUCKET")
		DB_DSN_TRACING      = env.Must("DB_DSN_TRACING")
	)
	// Setup all database connections
	sensorbucketDb, err := createDB(DB_DSN_SENSORBUCKET)
	if err != nil {
		panic(fmt.Errorf("create conn to sensorbucket db: %w", err))
	}
	tracingDb, err := createDB(DB_DSN_TRACING)
	if err != nil {
		panic(fmt.Errorf("create conn to tracing db: %w", err))
	}
	dbStore := dbStore{
		sensorbucketDb: sensorbucketDb,
		tracingDb:      tracingDb,
	}
	s := service{
		store: &dbStore,
	}
	return s.clean()
}

func runWarn() error {
	// Use a mock while sending mails are not required yet
	mailSender := mailMock{}
	s := service{
		mailer: &mailSender,
	}
	return s.warn("mock@mock.nl", "mock2@mock.nl", 15)
}

func createDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(10)
	return db, nil
}
