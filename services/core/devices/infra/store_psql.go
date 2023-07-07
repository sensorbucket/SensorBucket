package deviceinfra

import (
	"database/sql"
	"errors"
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/core/devices"
)

var (
	pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	_ devices.Store = (*PSQLStore)(nil)
)

type DB interface {
	Select(dest any, query string, args ...any) error
	Get(dest any, query string, args ...any) error
	sqlx.Execer
	sqlx.Queryer
}

type PSQLStore struct {
	db *sqlx.DB
}

func NewPSQLStore(db *sqlx.DB) *PSQLStore {
	return &PSQLStore{
		db: db,
	}
}

type DeviceModel struct {
	devices.Device
}

type SensorModel struct {
	*devices.Sensor
	DeviceID int64 `db:"device_id"`
}

func sensorModelsToSensors(models []SensorModel) []devices.Sensor {
	var sensors = make([]devices.Sensor, len(models))
	for ix := range models {
		sensors[ix] = *models[ix].Sensor
	}
	return sensors
}

type DevicePaginationQuery struct {
	CreatedAt time.Time `pagination:"created_at,ASC"`
	ID        int64     `pagination:"id,ASC"`
}

func (s *PSQLStore) ListInBoundingBox(filter devices.DeviceFilter, p pagination.Request) (*pagination.Page[devices.Device], error) {
	return newDeviceQueryBuilder().WithPagination(p).WithFilters(filter).WithinBoundingBox(filter.BoundingBoxFilter).Query(s.db)
}

func (s *PSQLStore) ListInRange(filter devices.DeviceFilter, p pagination.Request) (*pagination.Page[devices.Device], error) {
	return newDeviceQueryBuilder().WithPagination(p).WithFilters(filter).WithinRange(filter.RangeFilter).Query(s.db)
}

func (s *PSQLStore) List(filter devices.DeviceFilter, p pagination.Request) (*pagination.Page[devices.Device], error) {
	return newDeviceQueryBuilder().WithPagination(p).WithFilters(filter).Query(s.db)
}

func (s *PSQLStore) Find(id int64) (*devices.Device, error) {
	return find(s.db, id)
}

func (s *PSQLStore) Delete(dev *devices.Device) error {
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

type SensorPaginationQuery struct {
	CreatedAt time.Time `pagination:"created_at,ASC"`
	ID        int64     `pagination:"id,ASC"`
}

func (s *PSQLStore) ListSensors(p pagination.Request) (*pagination.Page[devices.Sensor], error) {
	var err error

	q := pq.Select(
		"id", "code", "description", "external_id", "properties", "archive_time",
		"brand", "created_at",
	).From("sensors")

	cursor := pagination.GetCursor[SensorPaginationQuery](p)
	q, err = pagination.Apply(q, cursor)
	if err != nil {
		return nil, err
	}

	rows, err := q.RunWith(s.db).Query()
	if err != nil {
		return nil, err
	}

	var sensors []devices.Sensor
	for rows.Next() {
		var sensor devices.Sensor
		err := rows.Scan(
			&sensor.ID,
			&sensor.Code,
			&sensor.Description,
			&sensor.ExternalID,
			&sensor.Properties,
			&sensor.ArchiveTime,
			&sensor.Brand,
			&sensor.CreatedAt,
			&cursor.Columns.CreatedAt,
			&cursor.Columns.ID,
		)
		if err != nil {
			return nil, err
		}
		sensors = append(sensors, sensor)
	}

	page := pagination.CreatePageT(sensors, cursor)
	return &page, nil
}

func (s *PSQLStore) createDevice(dev *devices.Device) error {
	if err := s.db.Get(&dev.ID,
		`
			INSERT INTO "devices" (
				"code", "description", "organisation", "properties", "location",
				"altitude", "location_description", "state", "created_at"
			)
			VALUES ($1, $2, $3, $4, ST_POINT($5, $6), $7, $8, $9, $10)
			RETURNING id
		`,
		dev.Code, dev.Description, dev.Organisation, dev.Properties,
		dev.Longitude, dev.Latitude, dev.Altitude, dev.LocationDescription,
		dev.State, dev.CreatedAt,
	); err != nil {
		return err
	}
	return nil
}

func (s *PSQLStore) updateDevice(dev *devices.Device) error {
	if _, err := s.db.Exec(`
	UPDATE 
		devices
	SET
		description=$2, properties=$3, location=ST_POINT($4, $5), altitude=$6,
		location_description=$7, state=$8
	WHERE
		id=$1`,
		dev.ID, dev.Description, dev.Properties, dev.Longitude, dev.Latitude,
		dev.Altitude, dev.LocationDescription, dev.State,
	); err != nil {
		return err
	}
	return nil
}

func (s *PSQLStore) updateSensors(devID int64, sensors []devices.Sensor) error {
	// Replace sensors
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	// Get delta with db
	dbSensors, err := listSensors(tx, func(q sq.SelectBuilder) sq.SelectBuilder {
		return q.Where(sq.Eq{"device_id": devID})
	})
	if err != nil {
		tx.Rollback()
		return err
	}
	var createdSensors []SensorModel
	var updatedSensors []SensorModel
	var deletedSensors []int64
create_delta_loop:
	for ix := range sensors {
		s := &sensors[ix]
		// If the sensor has ID 0 then it is new and must be created
		if s.ID == 0 {
			createdSensors = append(createdSensors, SensorModel{Sensor: s, DeviceID: devID})
			continue
		}
		// If the sensor is present in db and this list then assume it is updated
		for _, dbs := range dbSensors {
			if dbs.ID == s.ID {
				updatedSensors = append(updatedSensors, SensorModel{Sensor: s, DeviceID: devID})
				continue create_delta_loop
			}
		}
		// If we reach this, it means the sensor has an ID but is not present
		// in the database. Should not be possible!
		log.Printf("[WARNING] device/store_psql#updateSensors was called with an unknown sensor (id: %d) that is neither new or already exists\n", s.ID)
	}
	// If a sensor is present in the database but not in the new list then it was removed
deleted_delta_loop:
	for _, dbs := range dbSensors {
		for _, s := range sensors {
			if s.ID == dbs.ID {
				continue deleted_delta_loop
			}
		}
		deletedSensors = append(deletedSensors, dbs.ID)
	}

	if err := createSensors(tx, createdSensors); err != nil {
		tx.Rollback()
		return err
	}
	if err := updateSensors(tx, updatedSensors); err != nil {
		tx.Rollback()
		return err
	}
	if err := deleteSensors(tx, deletedSensors); err != nil {
		tx.Rollback()
		return err
	}

	// Commit changes
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func find(db DB, id int64) (*devices.Device, error) {
	var dev DeviceModel

	// Get device model
	if err := db.Get(&dev, `
		SELECT 
			"id", "code", "description", "organisation", "properties", "location_description",
			ST_X("location"::geometry) AS longitude, ST_Y("location"::geometry) AS latitude,
			"altitude", "state"
		FROM devices
		WHERE id=$1
	`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, devices.ErrDeviceNotFound
		}
		return nil, err
	}

	sensors, err := listSensors(db, func(q sq.SelectBuilder) sq.SelectBuilder {
		return q.Where(sq.Eq{"device_id": id})
	})
	if err != nil {
		return nil, err
	}
	dev.Sensors = sensorModelsToSensors(sensors)

	return &dev.Device, nil
}

type SelectQueryMod func(q sq.SelectBuilder) sq.SelectBuilder

func listSensors(db DB, mods ...SelectQueryMod) ([]SensorModel, error) {
	q := pq.Select(
		"s.id", "s.brand", "s.code", "s.description", "s.external_id", "s.properties", "s.archive_time",
		"s.device_id", "s.created_at",
	).From("sensors s")

	// Apply mods
	for _, mod := range mods {
		q = mod(q)
	}

	rows, err := q.RunWith(db).Query()
	if err != nil {
		return nil, err
	}

	sensors := []SensorModel{}
	for rows.Next() {
		var s = SensorModel{
			Sensor: &devices.Sensor{},
		}
		if err := rows.Scan(
			&s.ID, &s.Brand, &s.Code, &s.Description, &s.ExternalID, &s.Properties, &s.ArchiveTime,
			&s.DeviceID, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		sensors = append(sensors, s)
	}
	return sensors, nil
}

func createSensors(tx DB, sensors []SensorModel) error {
	if len(sensors) == 0 {
		return nil
	}
	q := pq.Insert("sensors").Columns(
		"code", "brand", "description", "archive_time", "properties", "external_id",
		"device_id", "created_at",
	).Suffix("RETURNING id")
	for _, s := range sensors {
		q = q.Values(
			s.Code, s.Brand, s.Description, s.ArchiveTime, s.Properties, s.ExternalID,
			s.DeviceID, s.CreatedAt,
		)
	}
	query, params, err := q.ToSql()
	if err != nil {
		return err
	}
	var ids []int64
	if err := tx.Select(&ids, query, params...); err != nil {
		return err
	}
	for ix := range ids {
		sensors[ix].ID = ids[ix]
	}
	return nil
}
func updateSensors(tx DB, sensors []SensorModel) error {
	if len(sensors) == 0 {
		return nil
	}
	for _, s := range sensors {
		_, err := pq.Update("sensors").Where(sq.Eq{"id": s.ID}).
			SetMap(map[string]any{
				"code":         s.Code,
				"brand":        s.Brand,
				"description":  s.Description,
				"archive_time": s.ArchiveTime,
				"properties":   s.Properties,
				"external_id":  s.ExternalID,
				"device_id":    s.DeviceID,
			}).RunWith(tx).Exec()
		if err != nil {
			return err
		}
	}
	return nil
}
func deleteSensors(tx DB, sensors []int64) error {
	if len(sensors) == 0 {
		return nil
	}
	_, err := pq.Delete("sensors").Where(sq.Eq{"id": sensors}).RunWith(tx).Exec()
	return err
}

func (s *PSQLStore) Save(dev *devices.Device) error {
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
