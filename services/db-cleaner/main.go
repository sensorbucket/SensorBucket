package main

import (
	"flag"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {
	cleanPtr := flag.Bool("clean", false, "indicates whether to delete all expired data from the tracing and sensorbucket database")
	warnExpiredPtr := flag.Bool("warn-expired", false, "indicates whether to alert users by email that expired data will soon be deleted")
	warnStuckPtr := flag.Bool("warn-stuck", false, "indicates whether to alert users by email about any data that has been stuck in a pipeline")
	warnErrorsPtr := flag.Bool("warn-errors", false, "indicated whether to alert users by email about any data that has been generating recurring errors")
	flag.Parse()

	if *cleanPtr {
		err := runCleanup()
		if err != nil {
			panic(err)
		}
	}
	if *warnExpiredPtr {
		err := runWarnExpired()
		if err != nil {
			panic(err)
		}
	}

	if *warnStuckPtr {
		err := runWarnStuck()
		if err != nil {
			panic(err)
		}
	}

	if *warnErrorsPtr {
		err := runWarnErrors()
		if err != nil {
			panic(err)
		}
	}

	if !*cleanPtr && !*warnExpiredPtr && !*warnStuckPtr && !*warnErrorsPtr {
		log.Println("[Warning] no clean or warn parameters given. not running.")
	}
}

func runCleanup() error {
	c := baseConfig().withSensorbucketAndTracingDb()
	s := service{
		store: &c.store,
	}
	return s.clean()
}

func runWarnExpired() error {
	c := baseConfig().withMailer().withDaysTillDeletionConfig()
	s := service{
		mailer: c.mailSender,
	}
	return s.warnExpired(c.fromEmail, c.toEmail, int(c.daysTillDeletion))
}

func runWarnErrors() error {
	c := baseConfig().withMailer().withErrorThreshold().withApiClient().withTimeframe()
	s := service{
		apiClient: c.apiClient.TracingApi,
		mailer:    c.mailSender,
	}
	return s.warnRecurringErrors(c.fromEmail, c.toEmail, int(c.errorThreshold), int(c.checkLastHours))
}

func runWarnStuck() error {
	c := baseConfig().withMailer().withApiClient().withDataTimeout().withTimeframe()
	s := service{
		mailer:    c.mailSender,
		apiClient: c.apiClient.TracingApi,
	}
	return s.warnStuck(c.fromEmail, c.toEmail, int32(c.dataTimeout), int(c.checkLastHours))
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
