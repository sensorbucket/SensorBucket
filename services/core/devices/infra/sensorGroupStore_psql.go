package deviceinfra

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/services/core/devices"
)

var _ devices.SensorGroupStore = (*PSQLSensorGroupStore)(nil)

type PSQLSensorGroupStore struct {
	db *sqlx.DB
}

func NewPSQLSensorGroupStore(db *sqlx.DB) *PSQLSensorGroupStore {
	return &PSQLSensorGroupStore{db}
}

func create(db DB, group *devices.SensorGroup) error {
	err := db.Get(&group.ID, `
        INSERT INTO sensor_groups (name, description)
        VALUES ($1, $2) RETURNING id;
        `, group.Name, group.Description)
	if err != nil {
		return err
	}
	return nil
}

func saveGroupSensors(db DB, group *devices.SensorGroup) error {
	var dbSensorGroupSensors []int64
	err := db.Select(&dbSensorGroupSensors, `
        SELECT sensor_id FROM sensor_groups_sensors WHERE sensor_group_id = $1
        `, group.ID)
	if err != nil {
		return err
	}

	added, removed := lo.Difference(group.Sensors, dbSensorGroupSensors)

	if len(removed) > 0 {
		_, err = pq.Delete("sensor_groups_sensors").Where(sq.Eq{"sensor_group_id": group.ID, "sensor_id": removed}).RunWith(db).Exec()
		if err != nil {
			return err
		}
	}
	if len(added) > 0 {
		q := pq.Insert("sensor_groups_sensors").Columns("sensor_group_id", "sensor_id")
		for _, newID := range added {
			q = q.Values(group.ID, newID)
		}
		_, err = q.RunWith(db).Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PSQLSensorGroupStore) Save(group *devices.SensorGroup) error {
	tx, err := s.db.BeginTxx(context.TODO(), nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	if group.ID == 0 {
		err = create(tx, group)
	} else {
		// err = update(tx, group)
		panic("Updating sensor group currently not implemented")
	}
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error updating/creating group: %w", err)
	}

	if err := saveGroupSensors(tx, group); err != nil {
		tx.Rollback()
		return fmt.Errorf("saving group sensors: %w", err)
	}

	return tx.Commit()
}
