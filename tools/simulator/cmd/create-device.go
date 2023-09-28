/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"sensorbucket.nl/sensorbucket/pkg/api"
)

// deviceCmd represents the device command
var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Create a new device",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		var req api.CreateDeviceRequest
		client := CreateAPIClient(cmd)

		dCode, _ := cmd.Flags().GetString("code")
		dDesc, _ := cmd.Flags().GetString("description")
		dLat, _ := cmd.Flags().GetFloat64("latitude")
		dLng, _ := cmd.Flags().GetFloat64("longitude")
		dLDesc, _ := cmd.Flags().GetString("locationdescription")
		dProp, _ := cmd.Flags().GetString("properties")

		req.SetCode(dCode)
		req.SetDescription(dDesc)
		if dLat != 0 && dLng != 0 {
			req.SetLatitude(dLat)
			req.SetLongitude(dLng)
		}
		if dLDesc != "" {
			req.SetLocationDescription(dLDesc)
		}
		if dProp != "" {
			json.Unmarshal([]byte(dProp), &req.Properties)
		}

		res, _, err := client.DevicesApi.CreateDevice(cmd.Context()).CreateDeviceRequest(req).Execute()
		if err != nil {
			return err
		}

		fmt.Printf("Created device. Device's ID is %d\n", int(res.Data.Id))

		return nil
	},
}

func init() {
	createCmd.AddCommand(deviceCmd)
	deviceCmd.Flags().StringP("code", "c", "", "Device's code")
	deviceCmd.Flags().StringP("description", "d", "", "Device's description")
	deviceCmd.Flags().Float64P("latitude", "l", 0, "Devices' latitude")
	deviceCmd.Flags().Float64P("longitude", "L", 0, "Devices' longitude")
	deviceCmd.Flags().StringP("locationdescription", "D", "", "Description for the device's location")
	deviceCmd.Flags().StringP("properties", "p", "", "Device properties")
	deviceCmd.MarkFlagRequired("code")
}
