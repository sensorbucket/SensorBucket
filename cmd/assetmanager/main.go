package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/rs/cors"
	"sensorbucket.nl/internal/assetmanager"
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
}

func Run() error {
	store := assetmanager.NewMongoDBStore()
	if err := store.Connect("mongodb://root:root@assetdb:27017/", "assets"); err != nil {
		return err
	}

	svc := assetmanager.New(assetmanager.Opts{
		Store:      store,
		PipelineID: os.Getenv("AM_PIPELINE_ID"),
	})

	// Register schemas
	schemas, err := readSchemas()
	if err != nil {
		return err
	}
	for _, schema := range schemas {
		if err := svc.RegisterAssetType(schema); err != nil {
			return err
		}
	}

	// Start http server
	fmt.Println("Starting http server on :3000")
	return http.ListenAndServe(":3000", cors.AllowAll().Handler(svc))
}

func readSchemas() ([]assetmanager.AssetSchema, error) {
	wd, _ := os.Getwd()
	file, err := os.Open(path.Join(wd, "tools/seed/assetmanager/schemas.json"))
	if err != nil {
		return nil, err
	}

	var schemas map[string]json.RawMessage
	if err := json.NewDecoder(file).Decode(&schemas); err != nil {
		return nil, err
	}

	var result []assetmanager.AssetSchema
	for name, schema := range schemas {
		result = append(result, assetmanager.AssetSchema{
			Type:   name,
			Schema: schema,
		})
	}

	return result, nil
}
