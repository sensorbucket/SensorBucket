package tracinginfra

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/tracing/tracing"
)

func NewStorePSQL(db *sqlx.DB) *stepStore {
	return &stepStore{
		db: db,
	}
}

type TraceQueryPage struct {
	TraceId     string `pagination:"trace_id,DESC"`
	SensorIndex int64  `pagination:"sensor_index,DESC"`
}

func (s *stepStore) Insert(step tracing.Step) error {
	if _, err := s.db.Exec(
		`INSERT INTO "steps" ("tracing_id", "step_index", "steps_remaining", "start_time", "error") VALUES ($1, $2, $3, $4, $5)`,
		step.TracingID,
		step.StepIndex,
		step.StepsRemaining,
		step.StartTime,
		step.Error); err != nil {
		return err
	}

	return nil
}

func (s *stepStore) Query(query tracing.Filter, r pagination.Request) (*pagination.Page[tracing.TraceDTO], error) {
	args := lo.Map(query.TraceIds, func(id uuid.UUID, _ int) any {
		return id.String()
	})
	rows, err := s.db.Query(traceQuery(query.TraceIds), args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	list := []traceStep{}
	for rows.Next() {
		var t traceStep
		err = rows.Scan(
			&t.StepIndex,
			&t.StepsRemaining,
			&t.TracingId,
			&t.Status,
			&t.TotalStatus,
			&t.Duration,
			&t.Error,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, t)
	}

	m := map[string][]traceStep{}

	for _, res := range list {
		if _, ok := m[res.TracingId]; !ok {
			m[res.TracingId] = []traceStep{
				res,
			}
			continue
		}
		m[res.TracingId] = append(m[res.TracingId], res)
	}

	total := []tracing.TraceDTO{}
	for key, val := range m {
		steps := []tracing.StepDTO{}
		for _, v := range val {
			var e string
			if v.Error != nil {
				e = *v.Error
			}
			steps = append(steps, tracing.StepDTO{
				Status:   tracing.Status(v.Status),
				Duration: v.Duration,
				Error:    e,
			})
		}

		if len(val) != (val[0].StepIndex + 1 + val[0].StepsRemaining) {
			for i := 0; i < val[0].StepIndex+1+val[0].StepsRemaining-len(val); i++ {
				last := val[len(val)-1]
				if last.Status == 4 {
					steps = append(steps, tracing.StepDTO{
						Status: tracing.Pending,
					})
				} else if last.Status == 5 {
					steps = append(steps, tracing.StepDTO{
						Status: tracing.Canceled,
					})
				} else {
					steps = append(steps, tracing.StepDTO{
						Status: tracing.Unknown,
					})
				}
			}
		}

		total = append(total, tracing.TraceDTO{
			TracingId: key,
			Status:    tracing.Status(val[0].Status),
			Steps:     steps,
		})
	}

	return &pagination.Page[tracing.TraceDTO]{
		Cursor: "TODO",
		Data:   total,
	}, nil
}

type traceStep struct {
	StepIndex      int
	StepsRemaining int
	TracingId      string
	Status         int
	TotalStatus    int
	Duration       time.Duration
	Error          *string
}

func traceQuery(traceIds []uuid.UUID) string {
	return `
	SELECT
	s1.step_index,
	s1.steps_remaining,
    s1.tracing_id,
    CASE
        WHEN s1.error <> '' THEN 5
        WHEN s2.start_time = 0 THEN 4
        ELSE 3
    END AS status,
    (
        SELECT MAX(CASE
                    WHEN s.error <> '' THEN 5
                    WHEN s.start_time = 0 THEN 4
                    ELSE 3
                END)
        FROM steps s
        WHERE s.tracing_id = s1.tracing_id
    ) AS total_status,
	COALESCE(s2.start_time - s1.start_time, 0) AS duration,
	s1.error
FROM
    steps s1
LEFT JOIN
    steps s2 ON s1.tracing_id = s2.tracing_id AND s1.step_index + 1 = s2.step_index
WHERE
    s1.tracing_id in ($1` + strings.Join(lo.RepeatBy(len(traceIds)-1, func(index int) string {
		return fmt.Sprintf("$%d", index+2)
	}), ",") + `)
ORDER BY
    s1.step_index;`
}

type stepStore struct {
	db *sqlx.DB
}
