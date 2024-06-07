package cmd

import (
	"net/url"

	"github.com/spf13/cobra"

	"sensorbucket.nl/sensorbucket/pkg/api"
)

func CreateAPIClient(cmd *cobra.Command) *api.APIClient {
	apiKey, err := cmd.Flags().GetString("apikey")
	if err != nil {
		panic(err)
	}
	if apiKey == "" {
		panic("Must provide APIKey using -k or --apikey")
	}

	host, _ := cmd.Flags().GetString("host")
	u, err := url.Parse(host)
	if err != nil {
		panic(err)
	}
	cfg := api.NewConfiguration()
	cfg.Scheme = u.Scheme
	cfg.Host = u.Host
	cfg.AddDefaultHeader("Authorization", "Bearer "+apiKey)
	client := api.NewAPIClient(cfg)
	return client
}
