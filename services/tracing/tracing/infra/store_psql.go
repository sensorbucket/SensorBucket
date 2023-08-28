package tracinginfra

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/tracing/tracing"
)

func NewStorePSQL(db *sqlx.DB) *stepStore {
	return &stepStore{
		db: db,
	}
}

func (s *stepStore) Insert(step tracing.Step) error {
	q := sq.Insert("steps").
		Columns("tracing_id", "step_index", "steps_remaining", "start_time", "error", "device_id").
		Values(step.TracingID, step.StepIndex, step.StepsRemaining, step.StartTime, step.Error, step.DeviceId)
	_, err := q.PlaceholderFormat(sq.Dollar).RunWith(s.db).Exec()
	if err != nil {
		return err
	}

	return nil
}

type TraceQueryPage struct {
	StartTime int64     `pagination:"traces.trace_start,DESC"`
	TracingID uuid.UUID `pagination:"tracing_id,DESC"`
}

func (s *stepStore) QueryEnrichedSteps() (*pagination.Page[tracing.EnrichedStep], error) {

}

func (s *stepStore) Query(filter tracing.Filter, r pagination.Request) (*pagination.Page[tracing.EnrichedStep], error) {
	var err error

	// Pagination
	cursor := pagination.GetCursor[TraceQueryPage](r)

	// First retrieve all the traces with pagination applied
	paginationQ := sq.Select("MIN(start_time) AS trace_start", "tracing_id").From("enriched_steps_view").GroupBy("tracing_id")

	if len(filter.DeviceIds) > 0 {
		paginationQ = paginationQ.Where(sq.Eq{"device_id": filter.DeviceIds})
	}

	if len(filter.TracingIds) > 0 {
		paginationQ = paginationQ.Where(sq.Eq{"tracing_id": filter.TracingIds})
	}

	if len(filter.Status) > 0 {
		paginationQ = paginationQ.Where(sq.Eq{"trace_status": tracing.StatusStringsToStatusCodes(filter.Status)})
	}

	if filter.DurationGreaterThan != nil {
		paginationQ = paginationQ.Where(sq.Gt{"duration": *filter.DurationGreaterThan})
	}

	if filter.DurationLowerThan != nil {
		paginationQ = paginationQ.Where(sq.Lt{"duration": *filter.DurationLowerThan})
	}

	paginationQ = sq.Select("traces.tracing_id").FromSelect(paginationQ, "traces")
	paginationQ, err = pagination.Apply(paginationQ, cursor)
	if err != nil {
		return nil, err
	}

	// Now get the more detailed information of a trace from the enriched_steps_view
	q := sq.Select(
		"tracing_id",
		"device_id",
		"step_index",
		"steps_remaining",
		"start_time",
		"error",
		"status",
		"duration",
		"trace_status",
	).
		From("enriched_steps_view").
		//Where(sq.Expr("tracing_id IN (?)", paginationQ)).
		OrderBy("start_time, tracing_id ASC").
		OrderBy("step_index ASC")

	q = q.JoinClause(paginationQ.Prefix("JOIN (").Suffix(") enriched ON (traces.tracing_id = enriched.tracing_id)"))
	fmt.Println(sq.DebugSqlizer(q))
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(s.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]tracing.EnrichedStep, 0, cursor.Limit)
	for rows.Next() {
		var t tracing.EnrichedStep
		err = rows.Scan(
			&t.Step.TracingID,
			&t.Step.DeviceId,
			&t.Step.StepIndex,
			&t.Step.StepsRemaining,
			&t.Step.StartTime,
			&t.Step.Error,
			&t.Status,
			&t.Duration,
			&t.HighestCollectiveStatus,
		)
		if err != nil {
			return nil, err
		}

		cursor.Columns.StartTime = t.StartTime
		cursor.Columns.TracingID, err = uuid.Parse(t.TracingID)
		if err != nil {
			return nil, err
		}

		list = append(list, t)
	}
	page := pagination.CreatePageT(list, cursor)
	return &page, nil
}

type stepStore struct {
	db *sqlx.DB
}
