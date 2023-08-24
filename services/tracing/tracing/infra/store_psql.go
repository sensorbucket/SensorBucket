package tracinginfra

import (
	"fmt"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	sq "github.com/Masterminds/squirrel"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/tracing/tracing"
)

func NewStorePSQL(db *sqlx.DB) *stepStore {
	return &stepStore{
		db: db,
	}
}

type TraceQueryPage struct {
	StartTime int64 `pagination:"s3.start_time,DESC"`
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

func (s *stepStore) Query(filter tracing.Filter, r pagination.Request) (*pagination.Page[tracing.EnrichedStep], error) {
	var err error

	// Build the case statement we need to derive the state of a step
	stepStatusCase := sq.Case().
		When("s1.error <> ''", strconv.Itoa(int(tracing.Failed))).
		When("s2 IS NULL AND s1.steps_remaining <> 0", strconv.Itoa(int(tracing.InProgress))).
		Else(strconv.Itoa(int(tracing.Success)))

	// Create the expression used to derivce a step's duration
	durationExpression := sq.ConcatExpr("COALESCE(s2.start_time - s1.start_time, 0)")
	highestStatusExpression := sq.ConcatExpr("MAX(s3.status) OVER (PARTITION BY s3.tracing_id)")

	// Finally build the query
	subQ := sq.Select(
		"s1.tracing_id",
		"s1.device_id",
		"s1.step_index",
		"s1.steps_remaining",
		"s1.start_time",
		"s1.error",
	).
		Column(sq.Alias(stepStatusCase, "status")).
		Column(sq.Alias(durationExpression, "duration")).
		From("steps s1").

		// To derive the state of the step we need to retrieve the next step in the database if it exists
		LeftJoin("steps s2 ON s1.tracing_id = s2.tracing_id AND s1.step_index + 1 = s2.step_index")

	q := sq.Select("*").Column(sq.Alias(highestStatusExpression, "highest_status")).FromSelect(subQ, "s3")
	if len(filter.DeviceIds) > 0 {
		q = q.Where(sq.Eq{"s3.device_id": filter.DeviceIds})
	}

	if len(filter.TracingIds) > 0 {
		q = q.Where(sq.Eq{"s3.tracing_id": filter.TracingIds})
	}

	if len(filter.Status) > 0 {
		q = sq.Select("*").FromSelect(q, "s3")

		// highest status is not yet evaluated in the main query which is why we have to put it into another sub query in order to do a where clause
		q = q.Where(sq.Eq{"s3.highest_status": tracing.StatusStringsToStatusCodes(filter.Status)})
	}

	if filter.DurationGreaterThan != nil {
		q = q.Where(sq.GtOrEq{"s3.duration": *filter.DurationGreaterThan})
	}

	if filter.DurationLowerThan != nil {
		q = q.Where(sq.LtOrEq{"s3.duration": *filter.DurationLowerThan})
	}

	// Pagination
	cursor := pagination.GetCursor[TraceQueryPage](r)
	q, err = pagination.Apply(q, cursor)
	if err != nil {
		return nil, err
	}
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
			&cursor.Columns.StartTime,
		)
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
