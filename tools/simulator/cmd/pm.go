package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/aquilax/go-perlin"
	"github.com/spf13/cobra"

	"sensorbucket.nl/sensorbucket/pkg/api"
)

var (
	flagSince      int
	flagDevices    []string
	flagPipelineID string
	flagInterval   int
)

var (
	SEED                 = 0xdeadbeef
	speed        float64 = 0.1 / 1000
	centerLat    float64 = 51.505
	centerLng    float64 = 3.593
	startTime, _         = time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
)

func CalculateXYOffsets(lat1, lon1, lat2, lon2 float64) (float64, float64) {
	// Conversion factor: One degree of latitude in meters
	const oneDegree = 111320.0

	// Calculate the difference in latitude and longitude
	deltaLat := lat2 - lat1
	deltaLon := lon2 - lon1

	// Convert latitude difference to Y offset in meters
	deltaY := deltaLat * oneDegree

	// Convert longitude difference to X offset in meters
	// Note: cosine of latitude should be calculated using radians
	deltaX := deltaLon * oneDegree * math.Cos(lat1*math.Pi/180.0)

	return deltaX, deltaY
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

		// Initialize Perlin noise generator
		alpha := 2.   // Recommended values are in range [1., 2.]
		beta := 2.    // Recommended values are in range [1., 2.]
		n := int32(3) // Number of iterations
		p := perlin.NewPerlin(alpha, beta, n, int64(SEED))

		//	return p.Noise2D(dx+dt*speed, dy+dt*speed)

		simOne := func(ctx context.Context, t time.Time, d api.Device) {
			dx, dy := CalculateXYOffsets(centerLat, centerLng, d.GetLatitude(), d.GetLongitude())
			dt := t.Sub(startTime).Seconds()
			v := (p.Noise2D(dx/100+speed*dt, dy/100+speed*dt) + 1) * 80
			data := map[string]any{
				"timestamp": t.Format(time.RFC3339),
				"value":     v,
				"device_id": d.Properties["device_id"],
			}
			_, err := client.UplinkApi.ProcessUplinkData(ctx, flagPipelineID).Body(data).Execute()
			if err != nil {
				log.Printf("Failed to simulate for devID: %d\n", d.Id)
			}
		}
		simAll := func(t time.Time) {
			wg := sync.WaitGroup{}
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()
			for _, d := range devices {
				wg.Add(1)
				go func() {
					simOne(ctx, t, d)
					wg.Done()
				}()
				wg.Wait()
			}
		}

		// Since in days, interval in seconds
		if flagSince > 0 {
			start := time.Now().Add(-time.Duration(flagSince) * 24 * time.Hour)
			interval := time.Second * time.Duration(flagInterval)
			occurances := time.Duration(flagSince) * 24 * time.Hour
			occurances /= time.Duration(flagInterval) * time.Second
			log.Printf("Simulating %d occurances before now...\n", int(occurances))
			c := make(chan uint8, 10)
			for i := 0; i < int(occurances); i++ {
				c <- 1
				go func() {
					t := start.Add(interval * time.Duration(i))
					simAll(t)
					<-c
				}()
				if i%20 == 0 {
					log.Printf("Simulating %d/%d (%f.3)\n", i, int(occurances), float32(i)/float32(occurances)*100)
				}
			}
		}

		tick := time.NewTicker(time.Duration(flagInterval) * time.Second)
		for {
			select {
			case <-tick.C:
				log.Printf("Simulating measurements....\n")
				simAll(time.Now())
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
	//nolint:errcheck
	pmCmd.MarkFlagRequired("devices")
	//nolint:errcheck
	pmCmd.MarkFlagRequired("pipeline")
}
