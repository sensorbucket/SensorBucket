package cmd

import (
	"net/url"

	"github.com/spf13/cobra"

	"sensorbucket.nl/sensorbucket/pkg/api"
)

func CreateAPIClient(cmd *cobra.Command) *api.APIClient {
	host, _ := cmd.Flags().GetString("host")
	u, err := url.Parse(host)
	if err != nil {
		panic(err)
	}
	cfg := api.NewConfiguration()
	cfg.Scheme = u.Scheme
	cfg.Host = u.Host
	client := api.NewAPIClient(cfg)
	return client
}
