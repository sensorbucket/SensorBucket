package tracinginfra

import (
	"time"

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

func (s *stepStore) UpsertStep(step tracing.Step, withError bool) error {
	q := sq.Insert("steps").
		Columns("tracing_id", "step_index", "steps_remaining", "start_time", "error", "device_id").
		Values(step.TracingID, step.StepIndex, step.StepsRemaining, step.StartTime, step.Error, step.DeviceID)
	if withError {
		q = q.Suffix("ON CONFLICT ON CONSTRAINT steps_pkey DO UPDATE SET error = ?", step.Error)
	} else {
		q = q.Suffix("ON CONFLICT ON CONSTRAINT steps_pkey DO NOTHING")
	}

	_, err := q.PlaceholderFormat(sq.Dollar).RunWith(s.db).Exec()
	if err != nil {
		return err
	}

	return nil
}

type TraceQueryPage struct {
	StartTime time.Time `pagination:"archive.archived_at,DESC"`
	TracingID uuid.UUID `pagination:"archive.tracing_id,DESC"`
}

func (s *stepStore) QueryTraces(filter tracing.Filter, r pagination.Request) (*pagination.Page[string], error) {
	var err error

	// Pagination
	cursor, err := pagination.GetCursor[TraceQueryPage](r)
	if err != nil {
		return nil, err
	}

	q := sq.Select().Distinct().From("archived_ingress_dtos archive").RightJoin("enriched_steps_view steps on archive.tracing_id = steps.tracing_id")
	if len(filter.DeviceIds) > 0 {
		q = q.Where(sq.Eq{"steps.device_id": filter.DeviceIds})
	}

	if len(filter.TracingIds) > 0 {
		q = q.Where(sq.Eq{"steps.tracing_id": filter.TracingIds})
	}

	if len(filter.Status) > 0 {
		q = q.Where(sq.Eq{"steps.trace_status": tracing.StatusStringsToStatusCodes(filter.Status)})
	}

	if filter.DurationGreaterThan != nil {
		q = q.Where(sq.Gt{"steps.duration": *filter.DurationGreaterThan})
	}

	if filter.DurationLowerThan != nil {
		q = q.Where(sq.Lt{"steps.duration": *filter.DurationLowerThan})
	}
	q, err = pagination.Apply(q, cursor)
	if err != nil {
		return nil, err
	}
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(s.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]string, 0, cursor.Limit)
	for rows.Next() {
		var tracingId string
		err = rows.Scan(
			&cursor.Columns.StartTime,
			&tracingId,
		)
		if err != nil {
			return nil, err
		}
		cursor.Columns.TracingID, err = uuid.Parse(tracingId)
		if err != nil {
			return nil, err
		}

		list = append(list, tracingId)
	}
	page := pagination.CreatePageT(list, cursor)
	return &page, nil
}

func (s *stepStore) GetStepsByTracingIDs(tracingIds []string) ([]tracing.EnrichedStep, error) {
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
		Where(sq.Eq{"tracing_id": tracingIds}).
		OrderBy("tracing_id, start_time, step_index ASC")

	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(s.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []tracing.EnrichedStep{}
	for rows.Next() {
		var t tracing.EnrichedStep
		err = rows.Scan(
			&t.Step.TracingID,
			&t.Step.DeviceID,
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
		list = append(list, t)
	}
	return list, nil
}

type stepStore struct {
	db *sqlx.DB
}
