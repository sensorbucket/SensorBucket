package store

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/services/device/service"
)

var (
// Already defined in store_psql.go
// pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
)

type deviceQueryBuilder struct {
	query sq.SelectBuilder
	err   error
}

func newDeviceQueryBuilder() deviceQueryBuilder {
	q := pq.Select(
		"id",
		"code",
		"description",
		"organisation",
		"configuration",
		"location_description",
		"ST_X(location::geometry) AS longitude",
		"ST_Y(location::geometry) AS latitude",
	).From("devices")

	return deviceQueryBuilder{query: q}
}

func (b deviceQueryBuilder) WithFilters(f service.DeviceFilter) deviceQueryBuilder {
	if f.Configuration != nil {
		b.query = b.query.Where("configuration::jsonb @> ?::jsonb", f.Configuration)
	}
	return b
}

func (b deviceQueryBuilder) WithinBoundingBox(bb service.BoundingBox) deviceQueryBuilder {
	// TODO: check if coordinates are valid?
	b.query = b.query.Where(
		`location::geometry @ ST_SetSRID(ST_MakeBox2D(ST_Point(?, ?), ST_Point(?, ?)),4326)`,
		bb.West, bb.South, bb.East, bb.North,
	)
	return b
}

func (b deviceQueryBuilder) WithinRange(r service.LocationRange) deviceQueryBuilder {
	b.query = b.query.Where(
		`ST_DWithin(location, ST_MakePoint(?, ?)::geography, ?)`,
		r.Longitude, r.Latitude, r.Distance,
	)
	return b
}

func (b deviceQueryBuilder) Query(db *sqlx.DB) ([]service.Device, error) {
	deviceModels := []DeviceModel{}

	// Fetch devices
	fmt.Println("Estimate query: ", sq.DebugSqlizer(b.query))
	rows, err := b.query.RunWith(db).Query()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var model DeviceModel
		err := rows.Scan(
			&model.ID,
			&model.Code,
			&model.Description,
			&model.Organisation,
			&model.Configuration,
			&model.LocationDescription,
			&model.Longitude,
			&model.Latitude,
		)
		if err != nil {
			return nil, err
		}
		deviceModels = append(deviceModels, model)
	}

	ids := make([]int64, len(deviceModels))
	devMap := map[int64]*DeviceModel{}
	for ix := range deviceModels {
		dev := &deviceModels[ix]
		// Create an array of all device ids, used to filter upcoming queries
		ids[ix] = dev.ID
		// Initialize default fields
		dev.Sensors = []service.Sensor{}
		// Create a map for id => device ptr, used to add sensors and location
		devMap[dev.ID] = dev
	}

	// Fetch sensors for devices
	sensorModels, err := listSensors(db, func(q sq.SelectBuilder) sq.SelectBuilder {
		return q.Where(sq.Eq{"device_id": ids})
	})
	if err != nil {
		return nil, err
	}
	for ix := range sensorModels {
		model := sensorModels[ix]
		dev := devMap[model.DeviceID]
		dev.Sensors = append(dev.Sensors, *model.Sensor)
	}

	devices := make([]service.Device, len(deviceModels))
	for ix, model := range deviceModels {
		devices[ix] = model.Device
	}

	return devices, nil
}
