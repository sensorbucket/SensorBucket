package deviceinfra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/core/devices"
)

var _ devices.SensorGroupStore = (*PSQLSensorGroupStore)(nil)

type PSQLSensorGroupStore struct {
	db *sqlx.DB
}

func NewPSQLSensorGroupStore(db *sqlx.DB) *PSQLSensorGroupStore {
	return &PSQLSensorGroupStore{db}
}

func createSensorGroup(db DB, group *devices.SensorGroup) error {
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
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("error fetching existing sensor from sensor group: %w", err)
	}

	added, removed := lo.Difference(group.Sensors, dbSensorGroupSensors)

	if len(removed) > 0 {
		_, err = pq.Delete("sensor_groups_sensors").Where(sq.Eq{"sensor_group_id": group.ID, "sensor_id": removed}).RunWith(db).Exec()
		if err != nil {
			return fmt.Errorf("error deleting sensors from sensor group: %w", err)
		}
	}
	if len(added) > 0 {
		q := pq.Insert("sensor_groups_sensors").Columns("sensor_group_id", "sensor_id")
		for _, newID := range added {
			q = q.Values(group.ID, newID)
		}
		_, err = q.RunWith(db).Exec()
		if err != nil {
			return fmt.Errorf("error adding sensors to group: %w", err)
		}
	}

	return nil
}

func updateSensorGroup(db DB, group *devices.SensorGroup) error {
	if _, err := db.Exec(`
	UPDATE 
		sensor_groups
	SET
        description = $2, name = $3
	WHERE
		id=$1`,
		group.ID, group.Description, group.Name,
	); err != nil {
		return err
	}
	return nil
}

func (s *PSQLSensorGroupStore) Save(group *devices.SensorGroup) error {
	tx, err := s.db.BeginTxx(context.TODO(), nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	if group.ID == 0 {
		err = createSensorGroup(tx, group)
	} else {
		err = updateSensorGroup(tx, group)
	}
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback failed with %w while handling error: %w", rb, err)
		}
		return fmt.Errorf("error updating/creating group: %w", err)
	}

	if err := saveGroupSensors(tx, group); err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback failed with %w while handling error: %w", rb, err)
		}
		return fmt.Errorf("saving group sensors: %w", err)
	}

	return tx.Commit()
}

func (s *PSQLSensorGroupStore) List(p pagination.Request) (*pagination.Page[devices.SensorGroup], error) {
	type SensorGroupPaginationQuery struct {
		CreatedAt time.Time `pagination:"sg.created_at,ASC"`
		ID        int64     `pagination:"sg.id,ASC"`
	}
	var err error
	q := pq.Select("sg.id", "sg.name", "sg.description", "sgs.sensor_id").From("sensor_groups sg").
		LeftJoin("sensor_groups_sensors sgs on sgs.sensor_group_id = sg.id")

	cursor, err := pagination.GetCursor[SensorGroupPaginationQuery](p)
	if err != nil {
		return nil, fmt.Errorf("list sensorsGroups, error getting pagination cursor: %w", err)
	}
	q, err = pagination.Apply(q, cursor)
	if err != nil {
		return nil, err
	}

	rows, err := q.RunWith(s.db).Query()
	if err != nil {
		return nil, err
	}

	groupMap := make(map[int64]devices.SensorGroup)
	for rows.Next() {
		var (
			groupID          int64
			sensorID         *int64
			groupName        string
			groupDescription string
		)
		if err := rows.Scan(
			&groupID, &groupName, &groupDescription, &sensorID,
			&cursor.Columns.CreatedAt, &cursor.Columns.ID,
		); err != nil {
			return nil, fmt.Errorf("scanning sensor group: %w", err)
		}

		group, ok := groupMap[groupID]
		// If ok is false, then this group mustbe instantiate
		if !ok {
			group = devices.SensorGroup{
				ID:          groupID,
				Name:        groupName,
				Description: groupDescription,
				Sensors:     make([]int64, 0),
			}
		}
		// If sensorID is nil, then this is a group without sensors
		if sensorID != nil {
			group.Sensors = append(group.Sensors, *sensorID)
		}
		groupMap[groupID] = group
	}

	groups := lo.Values(groupMap)
	page := pagination.CreatePageT(groups, cursor)
	return &page, nil
}

func (s *PSQLSensorGroupStore) Get(id int64) (*devices.SensorGroup, error) {
	q := pq.Select("sg.id", "sg.name", "sg.description", "sgs.sensor_id").From("sensor_groups sg").
		LeftJoin("sensor_groups_sensors sgs on sgs.sensor_group_id = sg.id").Where(sq.Eq{"sg.id": id})

	rows, err := q.RunWith(s.db).Query()
	// As this is an exec query with a DB Cursor, sql.ErrNoRows will not be thrown
	// So at the end of this function we check if `for rows.Next()` was actually
	// ran, if not then throw 404
	if err != nil {
		return nil, err
	}

	var group *devices.SensorGroup
	for rows.Next() {
		var (
			groupID          int64
			sensorID         *int64
			groupName        string
			groupDescription string
		)
		if err := rows.Scan(
			&groupID, &groupName, &groupDescription, &sensorID,
		); err != nil {
			return nil, fmt.Errorf("scanning sensor group: %w", err)
		}

		// If group is nil, then first create it
		if group == nil {
			group = &devices.SensorGroup{
				ID:          groupID,
				Name:        groupName,
				Description: groupDescription,
				Sensors:     make([]int64, 0),
			}
		}
		// If sensorID is nil, then this is a group without sensors
		if sensorID != nil {
			group.Sensors = append(group.Sensors, *sensorID)
		}
	}
	if group == nil {
		return nil, devices.ErrSensorGroupNotFound
	}

	return group, nil
}

func (s *PSQLSensorGroupStore) Delete(id int64) error {
	tx, err := s.db.BeginTxx(context.TODO(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM sensor_groups_sensors WHERE sensor_group_id = $1", id)
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback failed with %w while handling error: %w", rb, err)
		}
		return fmt.Errorf("could not delete sensor_group_sensors: %w", err)
	}
	_, err = tx.Exec("DELETE FROM sensor_groups WHERE id = $1", id)
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback failed with %w while handling error: %w", rb, err)
		}
		return fmt.Errorf("could not delete sensor_group: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing: %w", err)
	}

	return nil
}
