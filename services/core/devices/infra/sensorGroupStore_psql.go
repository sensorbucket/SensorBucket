package deviceinfra

import (
	"context"
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

func (s *PSQLSensorGroupStore) List(p pagination.Request) (*pagination.Page[devices.SensorGroup], error) {
	type SensorGroupPaginationQuery struct {
		CreatedAt time.Time `pagination:"sg.created_at,ASC"`
		ID        int64     `pagination:"sg.id,ASC"`
	}
	var err error
	q := pq.Select("sg.id", "sg.name", "sg.description", "sgs.sensor_id").From("sensor_groups sg").
		LeftJoin("sensor_groups_sensors sgs on sgs.sensor_group_id = sg.id")

	cursor := pagination.GetCursor[SensorGroupPaginationQuery](p)
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
