package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/cors"
	"sensorbucket.nl/internal/assetmanager"
)

var (
	AM_PIPELINE_ID     = mustEnv("AM_PIPELINE_ID")
	AM_DEFINITION_PATH = mustEnv("AM_DEFINITION_PATH")
	AM_HTTP_ADDR       = mustEnv("AM_HTTP_ADDR")
	AM_MONGO_DSN       = mustEnv("AM_MONGO_DSN")
)

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("%s must be set", key))
	}
	return val
}

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
}

func Run() error {
	store := assetmanager.NewMongoDBStore()
	if err := store.Connect(AM_MONGO_DSN, "assets"); err != nil {
		return err
	}

	svc := assetmanager.New(assetmanager.Opts{
		AssetStore:           store,
		AssetDefinitionStore: store,
		PipelineID:           AM_PIPELINE_ID,
	})

	// Register asset definitions
	if err := registerAssetDefinitions(svc); err != nil {
		return err
	}

	// Start http server
	fmt.Printf("Starting http server on %s\n", AM_HTTP_ADDR)
	return http.ListenAndServe(AM_HTTP_ADDR, cors.AllowAll().Handler(svc))
}

// assetDefinitionFile ...
type assetDefinitionFile struct {
	Version    int             `json:"version,omitempty"`
	Labels     []string        `json:"labels,omitempty"`
	Schema     json.RawMessage `json:"schema,omitempty"`
	PrimaryKey string          `json:"primary_key,omitempty"`
}

func registerAssetDefinitions(svc *assetmanager.Service) error {
	file, err := os.Open(AM_DEFINITION_PATH)
	if err != nil {
		return err
	}

	var ats map[string]assetDefinitionFile
	if err := json.NewDecoder(file).Decode(&ats); err != nil {
		return err
	}

	for name, at := range ats {
		err := svc.RegisterAssetDefinition(assetmanager.RegisterAssetDefinitionOpts{
			Name:       name,
			Labels:     at.Labels,
			Version:    at.Version,
			Schema:     at.Schema,
			PrimaryKey: at.PrimaryKey,
			PipelineID: AM_PIPELINE_ID,
		})
		if err != nil {
			return fmt.Errorf("could not register asset definition %s: %w", name, err)
		}
	}

	return nil
}
