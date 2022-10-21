package store

import (
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/services/device/service"
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
	LocationID int `db:"location_id"`
}

type SensorModel struct {
	service.Sensor
	ID       int
	DeviceID int `db:"device_id"`
}

func createLocationToDeviceMap(devs []DeviceModel) ([]int, map[int][]*DeviceModel) {
	ids := []int{}
	m := map[int][]*DeviceModel{}
	for ix := range devs {
		dev := &devs[ix]
		if dev.LocationID == 0 {
			continue
		}
		set, ok := m[dev.LocationID]
		if !ok {
			set = []*DeviceModel{}
			ids = append(ids, dev.LocationID)
		}
		set = append(set, dev)
		m[dev.LocationID] = set
	}

	return ids, m
}

func (s *PSQLStore) List(filter service.DeviceFilter) ([]service.Device, error) {
	deviceModels := []DeviceModel{}

	q := pq.Select("*").From("devices")
	if filter.Configuration != nil {
		q = q.Where("configuration::jsonb @> ?::jsonb", filter.Configuration)
	}
	query, params, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	// Fetch devices
	if err := s.db.Select(&deviceModels, query, params...); err != nil {
		return nil, err
	}

	ids := make([]int, len(deviceModels))
	devMap := map[int]*DeviceModel{}
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
	q = pq.Select("*").From("sensors").Where(sq.Eq{"device_id": ids})
	query, params, err = q.ToSql()
	if err != nil {
		return nil, err
	}
	var sensorModels []SensorModel
	if err := s.db.Select(&sensorModels, query, params...); err != nil {
		return nil, err
	}
	for _, model := range sensorModels {
		dev := devMap[model.DeviceID]
		dev.Sensors = append(dev.Sensors, model.Sensor)
	}

	// Fetch location for each device
	locIDs, devLoc := createLocationToDeviceMap(deviceModels)
	locations := []service.Location{}
	query, params, err = pq.Select("id", "name", "ST_X(location::geometry) AS latitude", "ST_Y(location::geometry) AS longitude").From("locations").Where(sq.Eq{"id": locIDs}).ToSql()
	if err != nil {
		return nil, err
	}
	if err := s.db.Select(&locations, query, params...); err != nil {
		return nil, err
	}
	for ix := range locations {
		loc := &locations[ix]
		devicesAtLocation := devLoc[loc.ID]
		for ix := range devicesAtLocation {
			devicesAtLocation[ix].Location = loc
		}
	}

	devices := make([]service.Device, len(deviceModels))
	for ix, model := range deviceModels {
		devices[ix] = model.Device
	}

	return devices, nil
}

func (s *PSQLStore) ListLocations() ([]service.Location, error) {
	var locs []service.Location
	if err := s.db.Select(&locs, `SELECT "id", "name", ST_X(location::geometry) AS latitude, ST_Y(location::geometry) as longitude FROM locations`); err != nil {
		return nil, err
	}
	return locs, nil
}

func (s *PSQLStore) createDevice(dev *service.Device) error {
	if err := s.db.Get(&dev.ID, "INSERT INTO devices (code, description, organisation, configuration) VALUES ($1, $2, $3, $4) RETURNING id", dev.Code, dev.Description, dev.Organisation, dev.Configuration); err != nil {
		return err
	}
	return nil
}

func (s *PSQLStore) Find(id int) (*service.Device, error) {
	var dev DeviceModel

	if err := s.db.Get(&dev, "SELECT * FROM devices WHERE id=$1", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	sensors := []service.Sensor{}
	if err := s.db.Select(&sensors, "SELECT code, description, external_id, configuration FROM sensors WHERE device_id=$1", id); err != nil {
		return nil, err
	}

	dev.Sensors = sensors
	return &dev.Device, nil
}

func (s *PSQLStore) updateDevice(dev *service.Device) error {
	if _, err := s.db.Exec(
		"UPDATE devices SET description=$2, configuration=$3 WHERE id=$1",
		dev.ID, dev.Description, dev.Configuration,
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
