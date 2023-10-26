/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"sensorbucket.nl/sensorbucket/pkg/api"
)

// sensorCmd represents the sensor command
var sensorCmd = &cobra.Command{
	Use:   "sensor",
	Short: "Create a new sensor for a device",
	Long:  ``,
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := CreateAPIClient(cmd)
		deviceIDStr := args[0]
		deviceID, err := strconv.ParseInt(deviceIDStr, 10, 64)
		if err != nil {
			return errors.New("device ID is not an integer")
		}
		var req api.CreateSensorRequest
		sCode, _ := cmd.Flags().GetString("code")
		sDesc, _ := cmd.Flags().GetString("description")
		sEID, _ := cmd.Flags().GetString("externalid")
		sBrand, _ := cmd.Flags().GetString("brand")

		req.SetCode(sCode)
		req.SetExternalId(sEID)
		req.SetDescription(sDesc)
		req.SetBrand(sBrand)

		_, _, err = client.DevicesApi.CreateDeviceSensor(cmd.Context(), int32(deviceID)).CreateSensorRequest(req).Execute()
		if err != nil {
			return err
		}

		fmt.Printf("Created sensor for deviced: %d\n", deviceID)

		return nil
	},
}

func init() {
	createCmd.AddCommand(sensorCmd)
	sensorCmd.Flags().StringP("code", "c", "", "Sensor code")
	sensorCmd.MarkFlagRequired("code")
	sensorCmd.Flags().StringP("description", "d", "", "Sensor description")
	sensorCmd.Flags().StringP("externalid", "e", "", "External sensor ID")
	sensorCmd.Flags().StringP("brand", "b", "", "Sensor brand")
}
