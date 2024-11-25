package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/cleanupper"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/healthchecker"
	"sensorbucket.nl/sensorbucket/services/fission-user-workers/migrations"
	userworkers "sensorbucket.nl/sensorbucket/services/fission-user-workers/service"
)

var (
	HTTP_ADDR = env.Could("HTTP_ADDR", ":3000")
	HTTP_BASE = env.Could("HTTP_BASE", "http://127.0.0.1:3000/api/workers")
	CTRL_TYPE = env.Could("CTRL_TYPE", "k8s")
	DB_DSN    = env.Must("DB_DSN")
	// The exchange to which workers will bind to
	AMQP_XCHG     = env.Could("AMQP_XCHG", "pipeline.messages")
	AUTH_JWKS_URL = env.Could("AUTH_JWKS_URL", "http://oathkeeper:4456/.well-known/jwks.json")
)

func main() {
	cleanup := cleanupper.Create()
	defer func() {
		if err := cleanup.Execute(5 * time.Second); err != nil {
			log.Printf("[Warn] Cleanup error(s) occured: %s\n", err)
		}
	}()
	if err := Run(cleanup); err != nil {
		panic(err)
	}
}

type (
	StubController struct{}
	Shutdowner     interface {
		Shutdown(context.Context) error
	}
)

type Controller interface {
	Reconcile(context.Context) error
}

func (c *StubController) Reconcile(context.Context) error {
	log.Println("WARNING, reconciling with stub controller, nothing will happen")
	return nil
}

func Run(cleanup cleanupper.Cleanupper) error {
	var err error

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	stopProfiler, err := web.RunProfiler()
	if err != nil {
		fmt.Printf("could not setup profiler server: %s\n", err)
	}
	cleanup.Add(stopProfiler)

	db := sqlx.MustOpen("pgx", DB_DSN)
	store := userworkers.NewPSQLStore(db)
	if err := migrations.MigratePostgres(db.DB); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	app := userworkers.NewApplication(store)
	jwks := auth.NewJWKSHttpClient(AUTH_JWKS_URL)
	srv := userworkers.NewHTTPTransport(app, HTTP_BASE, HTTP_ADDR, jwks)
	cleanup.Add(func(ctx context.Context) error {
		return srv.Stop(ctx)
	})
	go func() {
		if err := srv.Start(); !errors.Is(err, http.ErrServerClosed) && err != nil {
			log.Printf("Error starting HTTP Server: %v\n", err)
		}
	}()

	var ctrl Controller

	switch CTRL_TYPE {
	case "k8s":
		ctrl, err = userworkers.CreateKubernetesController(store, AMQP_XCHG)
		if err != nil {
			return err
		}
	case "docker":
		ctrl, err = userworkers.CreateDockerController(store)
		if err != nil {
			return err
		}
	default:
		log.Println("WARNING, no controller selected, defaulting to none meaning only the API will be accessible and no workers will be create")
		ctrl = &StubController{}
	}
	if shutdown, ok := ctrl.(Shutdowner); ok {
		cleanup.Add(shutdown.Shutdown)
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
					ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
					defer cancel()
					if err := ctrl.Reconcile(ctx); err != nil {
						log.Printf("Error reconciliating: %s\n", err.Error())
					}
					ticker = time.After(10 * time.Second)
				}
			}
		}
	}()

	shutdownHealth := healthchecker.Create().WithEnv().Start(ctx)
	cleanup.Add(shutdownHealth)

	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	return nil
}
