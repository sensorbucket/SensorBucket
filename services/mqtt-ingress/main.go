package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	mqtt "github.com/mochi-mqtt/server/v2"
	mqqtauth "github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/services/mqtt-ingress/service"
)

var (
	APIKEY_TRADE_URL = env.Must("APIKEY_TRADE_URL")
	AMQP_HOST        = env.Could("AMQP_HOST", "amqp://guest:guest@localhost/")
	AMQP_XCHG        = env.Could("AMQP_XCHG", "ingress")
	AMQP_XCHG_TOPIC  = env.Could("AMQP_XCHG_TOPIC", "ingress.httpimporter")
)

func main() {
	if err := Run(); err != nil {
		log.Fatalf("Error: %s\n", err.Error())
	}
}

func Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Create AMQP Message Queue
	mqConn := mq.NewConnection(AMQP_HOST)
	go mqConn.Start()
	defer mqConn.Shutdown()
	publisher := service.StartIngressDTOPublisher(mqConn, AMQP_XCHG, AMQP_XCHG_TOPIC)
	log.Printf("AMQP Publisher started...\n")

	authRules := &mqqtauth.Ledger{
		ACL: mqqtauth.ACLRules{
			{Filters: mqqtauth.Filters{
				"#": mqqtauth.WriteOnly,
			}},
		},
	}

	server := mqtt.New(nil)
	if err := server.AddHook(new(service.Auther), &service.AuthHookOptions{
		Context:   ctx,
		Publisher: publisher,
		APIKeyTrader: func(apiKey string) (string, error) {
			req, _ := http.NewRequest("GET", APIKEY_TRADE_URL, nil)
			req.Header.Set("Authorization", "Bearer "+apiKey)
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return "", err
			}
			if res.StatusCode != http.StatusOK {
				return "", fmt.Errorf("expected status 200 got: %d", res.StatusCode)
			}
			newAuth, _ := auth.StripBearer(res.Header.Get("Authorization"))
			return newAuth, nil
		},
	}); err != nil {
		return err
	}
	err := server.AddHook(new(mqqtauth.Hook), &mqqtauth.Options{
		Ledger: authRules,
	})
	if err != nil {
		return err
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:      "t1",
		Address: ":1883",
	})
	if err := server.AddListener(tcp); err != nil {
		return err
	}

	errC := make(chan error, 1)
	go func() {
		if err := server.Serve(); err != nil {
			errC <- err
		}
	}()

	log.Println("Server running")
	select {
	case err = <-errC:
		log.Println("Closing due to error")
		break
	case <-ctx.Done():
	}

	log.Println("Shutting down...")
	server.Close()

	return err
}
