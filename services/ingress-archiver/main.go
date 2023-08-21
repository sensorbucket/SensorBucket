package main

import (
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/services/ingress-archiver/migrations"
	ingressarchiver "sensorbucket.nl/sensorbucket/services/ingress-archiver/service"
)

var (
	DB_DSN                  = env.Must("DB_DSN")
	AMQP_HOST               = env.Must("AMQP_HOST")
	AMQP_QUEUE_INGRESS      = env.Could("AMQP_QUEUE_INGRESS", "archive-ingress")
	AMQP_XCHG_INGRESS       = env.Could("AMQP_XCHG_INGRESS", "ingress")
	AMQP_XCHG_INGRESS_TOPIC = env.Could("AMQP_XCHG_INGRESS_TOPIC", "ingress.*")
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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
	ingressarchiver.StartIngressDTOConsumer(
		amqpConn, svc,
		AMQP_QUEUE_INGRESS, AMQP_XCHG_INGRESS, AMQP_XCHG_INGRESS_TOPIC,
	)

	return nil
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
