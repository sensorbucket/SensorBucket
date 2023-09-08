package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/fission-user-workers/migrations"
)

var (
	HTTP_ADDR        = env.Could("HTTP_ADDR", ":3000")
	WORKER_NAMESPACE = env.Could("WORKER_NAMESPACE", "default")
	DB_DSN           = env.Must("DB_DSN")
)

func main() {
	if err := Run(); err != nil {
		panic(err)
	}
}

type Store interface {
	WorkersExists([]uuid.UUID) ([]uuid.UUID, error)
	ListUserWorkers(req pagination.Request) (*pagination.Page[UserWorker], error)
}

func Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db := sqlx.MustOpen("pgx", DB_DSN)
	store := newPSQLStore(db)
	if err := migrations.MigratePostgres(db.DB); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	srv := newHTTPTransport(HTTP_ADDR)
	go srv.Start()

	ctrl, err := createKubernetesController(store, WORKER_NAMESPACE)
	if err != nil {
		return err
	}

	// Start reconcile loop
	go func() {
		log.Println("Reconcile loop started")
		defer log.Println("Reconcile loop exited")
		ticker := time.After(1 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker:
				{
					if err := ctrl.Reconcile(ctx); err != nil {
						log.Printf("Error reconciliating: %s\n", err.Error())
					}
					ticker = time.After(10 * time.Second)
				}
			}
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down gracefully...")
	ctxTO, cancelTO := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTO()
	srv.Stop(ctxTO)

	return nil
}
