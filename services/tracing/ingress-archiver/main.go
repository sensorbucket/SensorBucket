package main

import (
	"fmt"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	ingressarchiver "sensorbucket.nl/sensorbucket/services/tracing/ingress-archiver/service"
	"sensorbucket.nl/sensorbucket/services/tracing/migrations"
)

var (
	DB_DSN                  = env.Must("DB_DSN")
	HTTP_ADDR               = env.Could("HTTP_ADDR", ":3000")
	AMQP_HOST               = env.Must("AMQP_HOST")
	AMQP_QUEUE_INGRESS      = env.Could("AMQP_QUEUE_INGRESS", "archive-ingress")
	AMQP_XCHG_INGRESS       = env.Could("AMQP_XCHG_INGRESS", "ingress")
	AMQP_XCHG_INGRESS_TOPIC = env.Could("AMQP_XCHG_INGRESS_TOPIC", "ingress.*")
)

func main() {
	if err := Run(); err != nil {
		panic(fmt.Sprintf("Error: %v\n", err))
	}
}

func Run() error {
	db, err := createDB()
	if err != nil {
		return err
	}
	amqpConn := mq.NewConnection(AMQP_HOST)
	go amqpConn.Start()

	store := ingressarchiver.NewStorePSQL(db)
	svc := ingressarchiver.New(store)
	go ingressarchiver.StartIngressDTOConsumer(
		amqpConn, svc,
		AMQP_QUEUE_INGRESS, AMQP_XCHG_INGRESS, AMQP_XCHG_INGRESS_TOPIC,
	)
	httpTransport := ingressarchiver.CreateHTTPTransport(svc)

	srv := &http.Server{
		Addr:         HTTP_ADDR,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      httpTransport,
	}

	return srv.ListenAndServe()
}

func createDB() (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", DB_DSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(10)
	if err := migrations.MigratePostgres(db.DB); err != nil {
		return nil, fmt.Errorf("failed to migrate db: %w", err)
	}
	return db, nil
}
