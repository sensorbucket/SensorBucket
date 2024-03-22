/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// ttnCmd represents the ttn command
var ttnCmd = &cobra.Command{
	Use:   "ttn <pipeline_id> <dataset_file>",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("host")
		pipelineId := args[0]
		dataSet := args[1]
		interval, _ := cmd.Flags().GetInt64("interval")

		jsonFile, err := os.Open(dataSet)
		if err != nil {
			return err
		}
		log.Println("caching data")
		var data []json.RawMessage
		if err := json.NewDecoder(jsonFile).Decode(&data); err != nil {
			return errors.New("failed to read input file")
		}
		jsonFile.Close()
		log.Println("done caching")

		for {
			randomIndex := rand.Intn(len(data))
			el := data[randomIndex]
			fmt.Println("sending", string(el))
			req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/uplinks/%s", host, pipelineId), bytes.NewBuffer(el))
			if err != nil {
				fmt.Println("Error creating request:", err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Error sending request:", err)
				continue
			}
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error sending request:", err)
				continue
			}
			fmt.Println("Response Status:", resp.Status)
			fmt.Println("Response Body:", string(respBody))
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	},
}

func init() {
	rootCmd.AddCommand(ttnCmd)
	ttnCmd.Flags().Int64P("interval", "i", 3000, "The interval between sent uplinks")
}
