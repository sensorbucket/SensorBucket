package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

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

	var (
		// SMTP Config
		SMTP_USERNAME      = env.Must("SMTP_USERNAME")
		SMTP_PASSWORD      = env.Must("SMTP_PASSWORD")
		SMTP_HOST          = env.Must("SMTP_HOST")
		TO_EMAIL           = env.Must("TO_EMAIL")
		FROM_EMAIL         = env.Must("FROM_EMAIL")
		DAYS_TILL_DELETION = env.Must("DAYS_TILL_DELETION")
	)

	// Use a mock while sending mails are not required yet
	mailSender := emailSender{
		username: SMTP_USERNAME,
		password: SMTP_PASSWORD,
		host:     SMTP_HOST,
	}
	s := service{
		mailer: &mailSender,
	}
	val, err := strconv.ParseInt(DAYS_TILL_DELETION, 10, 64)
	if err != nil {
		panic(err)
	}
	return s.warn(FROM_EMAIL, TO_EMAIL, int(val))
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
