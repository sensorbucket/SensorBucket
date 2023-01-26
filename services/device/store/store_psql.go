package store

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/services/device/service"
)

var (
	pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	_ service.Store = (*PSQLStore)(nil)
)

type PSQLStore struct {
	db *sqlx.DB
}

func NewPSQLStore(db *sqlx.DB) *PSQLStore {
	return &PSQLStore{
		db: db,
	}
}

type DeviceModel struct {
	service.Device
}

type SensorModel struct {
	service.Sensor
	ID       int
	DeviceID int `db:"device_id"`
}

func (s *PSQLStore) ListInBoundingBox(bb service.BoundingBox, filter service.DeviceFilter) ([]service.Device, error) {
	return newDeviceQueryBuilder().WithFilters(filter).WithinBoundingBox(bb).Query(s.db)
}

func (s *PSQLStore) ListInRange(r service.LocationRange, filter service.DeviceFilter) ([]service.Device, error) {
	return newDeviceQueryBuilder().WithFilters(filter).WithinRange(r).Query(s.db)
}

func (s *PSQLStore) List(filter service.DeviceFilter) ([]service.Device, error) {
	return newDeviceQueryBuilder().WithFilters(filter).Query(s.db)
}

func (s *PSQLStore) createDevice(dev *service.Device) error {
	if err := s.db.Get(&dev.ID,
		`
			INSERT INTO devices (code, description, organisation, configuration, location, location_description)
			VALUES ($1, $2, $3, $4, ST_POINT($5, $6), $7)
			RETURNING id
		`,
		dev.Code, dev.Description, dev.Organisation, dev.Configuration,
		dev.Longitude, dev.Latitude, dev.LocationDescription,
	); err != nil {
		return err
	}
	return nil
}

func (s *PSQLStore) Find(id int) (*service.Device, error) {
	var dev DeviceModel

	// Get device model
	if err := s.db.Get(&dev, `SELECT "id", "code", "description", "organisation", "configuration", "location_description", ST_X("location"::geometry) AS latitude, ST_Y("location"::geometry) AS longitude FROM devices WHERE id=$1`, id); err != nil {
		return nil, err
	}

	// Set sensors
	sensors := []service.Sensor{}
	if err := s.db.Select(&sensors, "SELECT code, description, external_id, configuration FROM sensors WHERE device_id=$1", id); err != nil {
		return nil, err
	}

	dev.Sensors = sensors
	return &dev.Device, nil
}

func (s *PSQLStore) updateDevice(dev *service.Device) error {
	if _, err := s.db.Exec(
		"UPDATE devices SET description=$2, configuration=$3, location=ST_POINT($4, $5), location_description=$6 WHERE id=$1",
		dev.ID, dev.Description, dev.Configuration, dev.Longitude, dev.Latitude, dev.LocationDescription,
	); err != nil {
		return err
	}
	return nil
}

func (s *PSQLStore) updateSensors(devID int, sensors []service.Sensor) error {
	// Replace sensors
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	if _, err := tx.Exec("DELETE FROM sensors WHERE device_id=$1", devID); err != nil {
		tx.Rollback()
		return err
	}

	// Only try to insert sensors if there are any left
	// otherwise SQL query will fail because there are no input values
	if len(sensors) > 0 {
		q := pq.Insert("sensors").Columns("code", "device_id", "description", "external_id", "configuration")
		for _, sensor := range sensors {
			q = q.Values(sensor.Code, devID, sensor.Description, sensor.ExternalID, sensor.Configuration)
		}
		query, params, err := q.ToSql()
		if err != nil {
			tx.Rollback()
			return err
		}

		if _, err := s.db.Exec(query, params...); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (s *PSQLStore) Save(dev *service.Device) error {
	var err error
	if dev.ID == 0 {
		err = s.createDevice(dev)
	} else {
		err = s.updateDevice(dev)
	}
	if err != nil {
		return err
	}

	if err := s.updateSensors(dev.ID, dev.Sensors); err != nil {
		return err
	}

	return nil
}

func (s *PSQLStore) Delete(dev *service.Device) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	if _, err := tx.Exec("DELETE FROM sensors WHERE device_id=$1", dev.ID); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec("DELETE FROM devices WHERE id=$1", dev.ID); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
