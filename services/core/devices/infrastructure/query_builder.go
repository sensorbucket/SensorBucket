package deviceinfra

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/device/service"
)

var (
// Already defined in store_psql.go
// pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
)

type deviceQueryBuilder struct {
	query  sq.SelectBuilder
	cursor pagination.Cursor[DevicePaginationQuery]
	err    error
}

func newDeviceQueryBuilder() deviceQueryBuilder {
	q := pq.Select(
		"id",
		"code",
		"description",
		"organisation",
		"properties",
		"location_description",
		"ST_X(location::geometry) AS longitude",
		"ST_Y(location::geometry) AS latitude",
		"altitude",
		"state",
		"created_at",
	).From("devices")

	return deviceQueryBuilder{query: q}
}

func (b deviceQueryBuilder) WithPagination(p pagination.Request) deviceQueryBuilder {
	if b.err != nil {
		return b
	}
	b.cursor = pagination.GetCursor[DevicePaginationQuery](p)
	return b
}

func (b deviceQueryBuilder) WithFilters(f service.DeviceFilter) deviceQueryBuilder {
	if f.Properties != nil {
		b.query = b.query.Where("properties::jsonb @> ?::jsonb", f.Properties)
	}
	return b
}

func (b deviceQueryBuilder) WithinBoundingBox(bb service.BoundingBoxFilter) deviceQueryBuilder {
	// TODO: check if coordinates are valid?
	b.query = b.query.Where(
		`location::geometry @ ST_SetSRID(ST_MakeBox2D(ST_Point(?, ?), ST_Point(?, ?)),4326)`,
		bb.West, bb.South, bb.East, bb.North,
	)
	return b
}

func (b deviceQueryBuilder) WithinRange(r service.RangeFilter) deviceQueryBuilder {
	b.query = b.query.Where(
		`ST_DWithin(location, ST_MakePoint(?, ?)::geography, ?)`,
		r.Longitude, r.Latitude, r.Distance,
	)
	return b
}

func (b deviceQueryBuilder) Query(db *sqlx.DB) (*pagination.Page[service.Device], error) {
	if b.err != nil {
		return nil, b.err
	}

	// Apply pagination
	q, err := pagination.Apply(b.query, b.cursor)
	if err != nil {
		return nil, err
	}

	// Fetch devices
	rows, err := q.RunWith(db).Query()
	if err != nil {
		return nil, err
	}

	deviceModels := []DeviceModel{}
	for rows.Next() {
		var model DeviceModel
		err := rows.Scan(
			&model.ID,
			&model.Code,
			&model.Description,
			&model.Organisation,
			&model.Properties,
			&model.LocationDescription,
			&model.Longitude,
			&model.Latitude,
			&model.Altitude,
			&model.State,
			&model.CreatedAt,
			&b.cursor.Columns.CreatedAt,
			&b.cursor.Columns.ID,
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

	// Create pagination Page
	page := pagination.CreatePageT(devices, b.cursor)
	return &page, nil
}
