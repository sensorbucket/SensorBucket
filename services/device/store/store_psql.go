package store

import (
	"database/sql"
	"errors"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/services/device/service"
)

var (
	pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	_ service.Store = (*PSQLStore)(nil)
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
	service.Device
}

type SensorModel struct {
	*service.Sensor
	DeviceID int64 `db:"device_id"`
}

func sensorModelsToSensors(models []SensorModel) []service.Sensor {
	var sensors = make([]service.Sensor, len(models))
	for ix := range models {
		sensors[ix] = *models[ix].Sensor
	}
	return sensors
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

func (s *PSQLStore) ListSensorGoals() ([]service.SensorGoal, error) {
	return listSensorGoals(s.db)
}

func (s *PSQLStore) ListSensorTypes() ([]service.SensorType, error) {
	return listSensorTypes(s.db)
}

func (s *PSQLStore) Find(id int64) (*service.Device, error) {
	return find(s.db, id)
}

func (s *PSQLStore) FindSensorGoal(id int64) (*service.SensorGoal, error) {
	return findSensorGoal(s.db, id)
}

func (s *PSQLStore) FindSensorType(id int64) (*service.SensorType, error) {
	return findSensorType(s.db, id)
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

func (s *PSQLStore) createDevice(dev *service.Device) error {
	if err := s.db.Get(&dev.ID,
		`
			INSERT INTO devices (code, description, organisation, configuration, location, location_description)
			VALUES ($1, $2, $3, $4, ST_POINT($5, $6), $7)
			RETURNING id
		`,
		dev.Code, dev.Description, dev.Organisation, dev.Configuration,
		dev.Latitude, dev.Longitude, dev.LocationDescription,
	); err != nil {
		return err
	}
	return nil
}

func (s *PSQLStore) updateDevice(dev *service.Device) error {
	if _, err := s.db.Exec(
		"UPDATE devices SET description=$2, configuration=$3, location=ST_POINT($4, $5), location_description=$6 WHERE id=$1",
		dev.ID, dev.Description, dev.Configuration, dev.Latitude, dev.Longitude, dev.LocationDescription,
	); err != nil {
		return err
	}
	return nil
}

func (s *PSQLStore) updateSensors(devID int64, sensors []service.Sensor) error {
	// Replace sensors
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	// Get delta with db
	dev, err := find(tx, devID)
	if err != nil {
		tx.Rollback()
		return err
	}
	var createdSensors []SensorModel
	var updatedSensors []SensorModel
	var deletedSensors []int64
	for ix := range sensors {
		s := &sensors[ix]
		// If the sensor has ID 0 then it is new and must be created
		if s.ID == 0 {
			createdSensors = append(createdSensors, SensorModel{Sensor: s, DeviceID: devID})
			continue
		}
		// If the sensor is present in db and this list then assume it is updated
		for _, dbs := range dev.Sensors {
			if dbs.ID == s.ID {
				updatedSensors = append(updatedSensors, SensorModel{Sensor: s, DeviceID: devID})
				continue
			}
		}
		// If we reach this, it means the sensor has an ID but is not present
		// in the database. Should not be possible!
		log.Printf("[WARNING] device/store_psql#updateSensors was called with an unknown sensor (id: %d) that is neither new or already exists\n", s.ID)
	}
	// If a sensor is present in the database but not in the new list then it was removed
	for _, dbs := range dev.Sensors {
		for _, s := range sensors {
			if s.ID == dbs.ID {
				continue
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

func listSensorGoals(db DB) ([]service.SensorGoal, error) {
	goals := []service.SensorGoal{}
	if err := db.Select(&goals, "SELECT id, name, description FROM sensor_goals"); err != nil {
		return nil, err
	}
	return goals, nil
}

func listSensorTypes(db DB) ([]service.SensorType, error) {
	types := []service.SensorType{}
	if err := db.Select(&types, "SELECT id, description FROM sensor_types"); err != nil {
		return nil, err
	}
	return types, nil

}

func find(db DB, id int64) (*service.Device, error) {
	var dev DeviceModel

	// Get device model
	if err := db.Get(&dev, `SELECT "id", "code", "description", "organisation", "configuration", "location_description", ST_X("location"::geometry) AS latitude, ST_Y("location"::geometry) AS longitude FROM devices WHERE id=$1`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrDeviceNotFound
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
		"s.id", "s.brand", "s.code", "s.description", "s.external_id", "s.configuration", "s.archive_time",
		"s.device_id",
		"sg.id", "sg.name", "sg.description",
		"st.id", "st.description",
	).
		From("sensors s").
		LeftJoin("sensor_goals sg ON s.goal_id = sg.id").
		LeftJoin("sensor_types st ON s.type_id = st.id")

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
			Sensor: &service.Sensor{
				Type: &service.SensorType{},
				Goal: &service.SensorGoal{},
			},
		}
		if err := rows.Scan(
			&s.ID, &s.Brand, &s.Code, &s.Description, &s.ExternalID, &s.Configuration, &s.ArchiveTime,
			&s.DeviceID,
			&s.Goal.ID, &s.Goal.Name, &s.Goal.Description,
			&s.Type.ID, &s.Type.Description,
		); err != nil {
			return nil, err
		}
		sensors = append(sensors, s)
	}
	return sensors, nil
}

func findSensorGoal(db DB, id int64) (*service.SensorGoal, error) {
	var goal service.SensorGoal
	if err := db.Get(&goal, "SELECT id, name, description FROM sensor_goals WHERE id = $1", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrSensorGoalNotFound
		}
		return nil, err
	}
	return &goal, nil
}

func findSensorType(db DB, id int64) (*service.SensorType, error) {
	var typ service.SensorType
	if err := db.Get(&typ, "SELECT id, description FROM sensor_types WHERE id = $1", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrSensorTypeNotFound
		}
		return nil, err
	}
	return &typ, nil
}

func createSensors(tx DB, sensors []SensorModel) error {
	if len(sensors) == 0 {
		return nil
	}
	q := pq.Insert("sensors").Columns(
		"code", "brand", "description", "goal_id", "type_id", "archive_time", "configuration", "external_id", "device_id",
	).Suffix("RETURNING id")
	for _, s := range sensors {
		q = q.Values(
			s.Code, s.Brand, s.Description, s.Goal.ID, s.Type.ID, s.ArchiveTime, s.Configuration, s.ExternalID, s.DeviceID,
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
		_, err := pq.Update("sensors").SetMap(map[string]any{
			"code":          s.Code,
			"brand":         s.Brand,
			"description":   s.Description,
			"goal_id":       s.Goal.ID,
			"type_id":       s.Type.ID,
			"archive_time":  s.ArchiveTime,
			"configuration": s.Configuration,
			"external_id":   s.ExternalID,
			"device_id":     s.DeviceID,
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
