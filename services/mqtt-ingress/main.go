package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/services/mqtt-ingress/service"
)

var APIKEY_TRADE_URL = env.Must("APIKEY_TRADE_URL")

func main() {
	if err := Run(); err != nil {
		log.Fatalf("Error: %s\n", err.Error())
	}
}

func Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	authRules := &auth.Ledger{
		ACL: auth.ACLRules{
			{Filters: auth.Filters{
				"#": auth.WriteOnly,
			}},
		},
	}

	server := mqtt.New(nil)
	if err := server.AddHook(new(service.Auther), &service.AuthHookOptions{
		Context: ctx,
		APIKeyTrader: func(apiKey string) (string, error) {
			req, _ := http.NewRequest("GET", APIKEY_TRADE_URL, nil)
			req.Header.Set("Authorization", "Bearer "+apiKey)
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return "", err
			}
			body, err := io.ReadAll(res.Body)
			if err != nil {
				return "", err
			}
			res.Body.Close()
			return string(body), nil
		},
	}); err != nil {
		return err
	}
	err := server.AddHook(new(auth.Hook), &auth.Options{
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
