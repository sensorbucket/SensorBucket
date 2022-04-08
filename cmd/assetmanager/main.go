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
		AssetStore:     store,
		AssetTypeStore: store,
		PipelineID:     os.Getenv("AM_PIPELINE_ID"),
	})

	// Register asset types
	if err := registerAssetTypes(svc); err != nil {
		return err
	}

	// Start http server
	fmt.Println("Starting http server on :3000")
	return http.ListenAndServe(":3000", cors.AllowAll().Handler(svc))
}

// assetTypeFile ...
type assetTypeFile struct {
	Version int             `json:"version,omitempty"`
	Labels  []string        `json:"labels,omitempty"`
	Schema  json.RawMessage `json:"schema,omitempty"`
}

func registerAssetTypes(svc *assetmanager.Service) error {
	wd, _ := os.Getwd()
	file, err := os.Open(path.Join(wd, "tools/seed/assetmanager/assettypes.json"))
	if err != nil {
		return err
	}

	var ats map[string]assetTypeFile
	if err := json.NewDecoder(file).Decode(&ats); err != nil {
		return err
	}

	for name, at := range ats {
		err := svc.RegisterAssetType(assetmanager.RegisterAssetTypeOpts{
			Name:       name,
			Labels:     at.Labels,
			Version:    at.Version,
			Schema:     at.Schema,
			PipelineID: os.Getenv("AM_PIPELINE_ID"),
		})
		if err != nil {
			return fmt.Errorf("could not register asset type %s: %w", name, err)
		}
	}

	return nil
}
