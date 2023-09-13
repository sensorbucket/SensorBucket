package cmd

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aquilax/go-perlin"
	"github.com/spf13/cobra"
)

var (
	flagSince      int
	flagDevices    []string
	flagPipelineID string
	flagInterval   int
)

type Location struct {
	x, y float64
}

func perl() {
	// Initialize Perlin noise generator
	alpha := 2.   // Recommended values are in range [1., 2.]
	beta := 2.    // Recommended values are in range [1., 2.]
	n := int32(3) // Number of iterations
	p := perlin.NewPerlin(alpha, beta, n, int64(time.Now().Unix()))

	// Define locations with x, y offsets
	locations := []Location{
		{0, 0},
		{10, 0},
		{0, 10},
		{10, 10},
	}

	// Time step (can be real-world time)
	timeStep := 0.1

	// Simulation loop
	for t := 0.; t <= 10; t += timeStep {
		fmt.Printf("Time: %f\n", t)
		for _, loc := range locations {
			// Generate Perlin noise based on location and time
			noiseValue := p.Noise2D(loc.x+t, loc.y+t)
			fmt.Printf("Location (%f, %f): %f\n", loc.x, loc.y, noiseValue)
		}
		fmt.Println("------------")
		time.Sleep(1 * time.Second)
	}
}

// pmCmd represents the pm command
var pmCmd = &cobra.Command{
	Use:   "pm",
	Short: "Simulates Particulate Matter measurements",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(flagDevices) == 0 {
			return errors.New("must define atleast one device")
		}
		ids := make([]int64, 0, len(flagDevices))
		for _, idString := range flagDevices {
			devID, err := strconv.ParseInt(idString, 10, 64)
			if err != nil {
				fmt.Printf("Could not use device ID: %s as it is not a number\n", idString)
				continue
			}
			ids = append(ids, devID)
		}

		client := CreateAPIClient(cmd)
		res, _, err := client.DevicesApi.ListDevices(cmd.Context()).Id(ids).Execute()
		if err != nil {
			return fmt.Errorf("could not get devices from API: %w", err)
		}
		devices := res.Data
		if len(devices) == 0 {
			return errors.New("no devices found for given IDs")
		}
		log.Printf("starting simulation with %d devices\n", len(devices))

		tick := time.NewTicker(time.Duration(flagInterval) * time.Second)
		for {
			select {
			case <-tick.C:
				log.Printf("Simulating measurements....\n")
				continue
			case <-cmd.Context().Done():
				return nil
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(pmCmd)
	pmCmd.Flags().IntVar(&flagSince, "since", 0, "Simulates N previous days as instantly. To simulate the previous 7 days use --since 7")
	pmCmd.Flags().StringSliceVarP(&flagDevices, "devices", "d", []string{}, "The device ID's to simulate")
	pmCmd.Flags().StringVarP(&flagPipelineID, "pipeline", "p", "", "The pipeline UUID to post to")
	pmCmd.Flags().IntVarP(&flagInterval, "interval", "i", 60, "Interval between simulated measurements in seconds")
}
