package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/services/fission-user-workers/migrations"
	userworkers "sensorbucket.nl/sensorbucket/services/fission-user-workers/service"
)

var (
	HTTP_ADDR        = env.Could("HTTP_ADDR", ":3000")
	WORKER_NAMESPACE = env.Could("WORKER_NAMESPACE", "default")
	CTRL_TYPE        = env.Could("CTRL_TYPE", "k8s")
	DB_DSN           = env.Must("DB_DSN")
	// The exchange to which workers will bind to
	AMQP_XCHG = env.Could("AMQP_XCHG", "pipeline.messages")
)

func main() {
	if err := Run(); err != nil {
		panic(err)
	}
}

type Controller interface {
	Reconcile(context.Context) error
}
type StubController struct{}

func (c *StubController) Reconcile(context.Context) error {
	log.Println("WARNING, reconciling with stub controller, nothing will happen")
	return nil
}

func Run() error {
	var err error

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db := sqlx.MustOpen("pgx", DB_DSN)
	store := userworkers.NewPSQLStore(db)
	if err := migrations.MigratePostgres(db.DB); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	app := userworkers.NewApplication(store)
	srv := userworkers.NewHTTPTransport(app, HTTP_ADDR)
	go srv.Start()

	var ctrl Controller

	switch CTRL_TYPE {
	case "k8s":
		ctrl, err = userworkers.CreateKubernetesController(store, WORKER_NAMESPACE, AMQP_XCHG)
		if err != nil {
			return err
		}
	case "docker":
		return errors.New("docker controller not yet implemented")
	default:
		log.Println("WARNING, no controller selected, defaulting to none meaning only the API will be accessible and no workers will be create")
		ctrl = &StubController{}
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
