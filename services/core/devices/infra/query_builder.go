package deviceinfra

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
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
		"device.id",
		"device.code",
		"device.description",
		"device.tenant_id",
		"device.properties",
		"device.location_description",
		"ST_X(device.location::geometry) AS longitude",
		"ST_Y(device.location::geometry) AS latitude",
		"device.altitude",
		"device.state",
		"device.created_at",
	).From("devices device")

	return deviceQueryBuilder{query: q}
}

func (b deviceQueryBuilder) WithPagination(p pagination.Request) deviceQueryBuilder {
	if b.err != nil {
		return b
	}
	b.cursor, b.err = pagination.GetCursor[DevicePaginationQuery](p)
	if b.err != nil {
		b.err = fmt.Errorf("list devices, error getting pagination cursor: %w", b.err)
	}
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

func (b deviceQueryBuilder) Query(ctx context.Context, db *sqlx.DB) (*pagination.Page[devices.Device], error) {
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
		q = q.Where("device.properties::jsonb @> ?::jsonb", b.filters.Properties)
	}
	if len(b.filters.Sensor) > 0 {
		// Update query here
		subQ, subArgs, err := sq.Select("DISTINCT sensor.device_id").From("sensors sensor").Where(sq.Eq{"sensor.id": b.filters.Sensor}).ToSql()
		if err != nil {
			return nil, err
		}
		q = q.Where(fmt.Sprintf("device.id in (%s)", subQ), subArgs...)
	}
	if len(b.filters.ID) > 0 {
		q = q.Where(sq.Eq{"device.id": b.filters.ID})
	}
	if len(b.filters.Code) > 0 {
		q = q.Where(sq.Eq{"device.code": b.filters.Code})
	}

	// Authorize
	q = auth.ProtectedQuery(ctx, "device.tenant_id", q)

	// Fetch devices
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deviceModels := []DeviceModel{}
	for rows.Next() {
		var model DeviceModel
		err := rows.Scan(
			&model.ID,
			&model.Code,
			&model.Description,
			&model.TenantID,
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
	sensors, err := listSensors(ctx, db, ListSensorsFilter{DeviceID: ids})
	if err != nil {
		return nil, err
	}
	for ix := range sensors {
		sensor := sensors[ix]
		dev := devMap[sensor.DeviceID]
		dev.Sensors = append(dev.Sensors, sensor)
	}

	devices := make([]devices.Device, len(deviceModels))
	for ix, model := range deviceModels {
		devices[ix] = model.Device
	}

	// Create pagination Page
	page := pagination.CreatePageT(devices, b.cursor)
	return &page, nil
}
