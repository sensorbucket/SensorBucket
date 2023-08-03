package deviceinfra

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/core/devices"
)

var (
// Already defined in store_psql.go
// pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
)

type deviceQueryBuilder struct {
	query   sq.SelectBuilder
	filters devices.DeviceFilter
	cursor  pagination.Cursor[DevicePaginationQuery]
	err     error
}

func newDeviceQueryBuilder() deviceQueryBuilder {
	q := sq.Select(
		"devices.id",
		"devices.code",
		"devices.description",
		"devices.organisation",
		"devices.properties",
		"devices.location_description",
		"ST_X(devices.location::geometry) AS longitude",
		"ST_Y(devices.location::geometry) AS latitude",
		"devices.altitude",
		"devices.state",
		"devices.created_at",
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

func (b deviceQueryBuilder) WithFilters(f devices.DeviceFilter) deviceQueryBuilder {
	b.filters = f
	return b
}

func (b deviceQueryBuilder) WithinBoundingBox(bb devices.BoundingBoxFilter) deviceQueryBuilder {
	// TODO: check if coordinates are valid?
	b.query = b.query.Where(
		`location::geometry @ ST_SetSRID(ST_MakeBox2D(ST_Point(?, ?), ST_Point(?, ?)),4326)`,
		bb.West, bb.South, bb.East, bb.North,
	)
	return b
}

func (b deviceQueryBuilder) WithinRange(r devices.RangeFilter) deviceQueryBuilder {
	b.query = b.query.Where(
		`ST_DWithin(location, ST_MakePoint(?, ?)::geography, ?)`,
		r.Longitude, r.Latitude, r.Distance,
	)
	return b
}

func (b deviceQueryBuilder) Query(db *sqlx.DB) (*pagination.Page[devices.Device], error) {
	if b.err != nil {
		return nil, b.err
	}

	// Apply pagination
	q, err := pagination.Apply(b.query, b.cursor)
	if err != nil {
		return nil, err
	}

	// Apply filters
	if b.filters.Properties != nil {
		q = q.Where("properties::jsonb @> ?::jsonb", b.filters.Properties)
	}
	if len(b.filters.Sensor) > 0 {
		// Update query here
		subQ, subArgs, err := sq.Select("DISTINCT sensors.device_id").From("sensors").Where(sq.Eq{"sensors.id": b.filters.Sensor}).ToSql()
		if err != nil {
			return nil, err
		}
		q = q.Where(fmt.Sprintf("devices.id in (%s)", subQ), subArgs...)
	}

	// Fetch devices
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(db).Query()
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
		dev.Sensors = []devices.Sensor{}
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

	devices := make([]devices.Device, len(deviceModels))
	for ix, model := range deviceModels {
		devices[ix] = model.Device
	}

	// Create pagination Page
	page := pagination.CreatePageT(devices, b.cursor)
	return &page, nil
}
