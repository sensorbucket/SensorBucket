package tracing

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

var pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type Service struct {
	db *sqlx.DB

	onceInsertStep      sync.Once
	stmtInsertStep      *sqlx.Stmt
	onceInsertTrace     sync.Once
	stmtInsertTrace     *sqlx.Stmt
	onceInsertIngress   sync.Once
	stmtInsertIngress   *sqlx.Stmt
	onceSetTraceError   sync.Once
	stmtSetTraceError   *sqlx.Stmt
	oncePeriodicCleanup sync.Once
	stmtPeriodicCleanup *sqlx.Stmt
}

func Create(db *sqlx.DB) *Service {
	return &Service{
		db: db,
	}
}

func (svc *Service) insertTraceStatement() *sqlx.Stmt {
	svc.onceInsertTrace.Do(func() {
		var err error
		svc.stmtInsertTrace, err = svc.db.Preparex(`
INSERT INTO traces (
  id, tenant_id, pipeline_id, created_at
)
VALUES ($1, $2, $3, $4)
`)
		if err != nil {
			log.Fatalf("ERROR on InsertTrace statement preparation: %s\n", err.Error())
		}
	})
	return svc.stmtInsertTrace
}

func (svc *Service) insertTraceIngressStatement() *sqlx.Stmt {
	svc.onceInsertIngress.Do(func() {
		var err error
		svc.stmtInsertIngress, err = svc.db.Preparex(`
INSERT INTO trace_ingress (
  id, tenant_id, pipeline_id, archived_at, payload
)
VALUES ($1, $2, $3, $4, $5)
`)
		if err != nil {
			log.Fatalf("ERROR on InsertIngress statement preparation: %s\n", err.Error())
		}
	})
	return svc.stmtInsertIngress
}

func (svc *Service) StoreTrace(msg processing.IngressDTO, queueTime time.Time) error {
	_, err := svc.insertTraceStatement().Exec(msg.TracingID, msg.TenantID, msg.PipelineID, queueTime)
	if err != nil {
		return fmt.Errorf("while inserting trace: %w", err)
	}
	return nil
}

func (svc *Service) StoreIngress(body []byte, ingress processing.IngressDTO, queueTime time.Time) error {
	_, err := svc.insertTraceIngressStatement().Exec(
		ingress.TracingID, ingress.TenantID, ingress.PipelineID, queueTime, body,
	)
	if err != nil {
		return fmt.Errorf("while inserting trace: %w", err)
	}
	return nil
}

func (svc *Service) setTraceErrorStatement() *sqlx.Stmt {
	svc.onceSetTraceError.Do(func() {
		var err error
		svc.stmtSetTraceError, err = svc.db.Preparex(`
UPDATE traces SET error = $1, error_at = $2 WHERE id = $3
`)
		if err != nil {
			log.Fatalf("ERROR on SetTraceError statement preparation: %s\n", err.Error())
		}
	})
	return svc.stmtSetTraceError
}

func (svc *Service) StoreTraceError(tracingID string, queueTime time.Time, error string) error {
	_, err := svc.setTraceErrorStatement().Exec(error, queueTime, tracingID)
	if err != nil {
		return fmt.Errorf("while setting error on trace: %w", err)
	}
	return nil
}

func (svc *Service) insertStepStatement() *sqlx.Stmt {
	svc.onceInsertStep.Do(func() {
		var err error
		svc.stmtInsertStep, err = svc.db.Preparex(
			`INSERT INTO trace_steps (tracing_id, worker_id, queue_time, device_id)
        VALUES ($1, $2, $3, $4)`,
		)
		if err != nil {
			log.Fatalf("ERROR on InsertTraceStep statement preparation: %s\n", err.Error())
		}
	})
	return svc.stmtInsertStep
}

func (svc *Service) StoreTraceStep(msg pipeline.Message, queueTime time.Time) error {
	deviceID := int64(0)
	if msg.Device != nil {
		deviceID = msg.Device.ID
	}

	workerID, err := msg.CurrentStep()
	if err != nil {
		return err
	}

	_, err = svc.insertStepStatement().Exec(
		msg.TracingID, workerID, queueTime, deviceID,
	)
	if err != nil {
		return fmt.Errorf("while processing pipeline step: %w", err)
	}
	return nil
}

type TraceFilter struct {
	Pipeline []string
}

type Trace struct {
	ID          uuid.UUID   `json:"id"`
	PipelineID  uuid.UUID   `json:"pipeline_id"`
	DeviceID    int64       `json:"device_id"`
	StartTime   time.Time   `json:"start_time"`
	Workers     []string    `json:"workers"`
	WorkerTimes []time.Time `json:"worker_times"`
	Error       *string     `json:"error"`
	ErrorAt     *time.Time  `json:"error_at"`
}

type tracePagination struct {
	StartTime time.Time `pagination:"trace.created_at,desc"`
	TraceID   uuid.UUID `pagination:"trace.id,desc"`
}

func (svc *Service) Query(ctx context.Context, filters TraceFilter, r pagination.Request) (*pagination.Page[Trace], error) {
	if err := auth.MustHavePermissions(ctx, auth.Permissions{auth.READ_DEVICES}); err != nil {
		return nil, err
	}

	cursor, err := pagination.GetCursor[tracePagination](r)
	if err != nil {
		return nil, fmt.Errorf("in query traces, getting pagination cursor: %w", err)
	}

	tracesQ := pq.Select(
		"trace.id", "trace.pipeline_id", "trace.created_at", "trace.error", "trace.error_at",
		"steps.device_id", "steps.workers", "steps.worker_times",
	).From("traces trace").OrderBy("trace.created_at DESC")

	if filters.Pipeline != nil {
		tracesQ = tracesQ.Where(sq.Eq{"trace.pipeline_id": filters.Pipeline})
		fmt.Printf("filters.PipelineID: %v\n", filters.Pipeline)
	}

	stepAggregationQ := pq.Select(
		"step.tracing_id",
		"array_agg(step.worker_id ORDER BY step.queue_time ASC) as workers",
		"array_agg(step.queue_time ORDER BY step.queue_time ASC) as worker_times",
		"max(step.device_id) as device_id",
	).From("trace_steps step").Where("trace.id = step.tracing_id").GroupBy("step.tracing_id")

	tracesQ = tracesQ.JoinClause(
		stepAggregationQ.Prefix("INNER JOIN LATERAL (").Suffix(") AS steps ON trace.id = steps.tracing_id"))

	tracesQ, err = pagination.Apply(tracesQ, cursor)
	if err != nil {
		return nil, err
	}
	tracesQ = auth.ProtectedQuery(ctx, tracesQ)

	query, params, err := tracesQ.ToSql()
	if err != nil {
		return nil, err
	}

	traces := make([]Trace, 0, r.Limit)

	c, err := svc.db.Conn(context.Background())
	if err != nil {
		return nil, fmt.Errorf("in GetTenantMember, could not get raw db conn: %w", err)
	}
	defer c.Close()
	err = c.Raw(func(driverConn any) error {
		stdlibConn, ok := driverConn.(*stdlib.Conn)
		if !ok {
			return errors.New("in GetHashedAPIKeyByNameAndTenantID, expected driverConnection to be of type stdlib.Conn")
		}
		conn := stdlibConn.Conn()
		rows, err := conn.Query(ctx, query, params...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var trace Trace
			if err := rows.Scan(
				&trace.ID, &trace.PipelineID, &trace.StartTime, &trace.Error, &trace.ErrorAt,
				&trace.DeviceID, &trace.Workers, &trace.WorkerTimes,
				&cursor.Columns.StartTime, &cursor.Columns.TraceID,
			); err != nil {
				return fmt.Errorf("in query traces while scanning row: %w", err)
			}
			traces = append(traces, trace)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("in query traces after rowscan: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	page := pagination.CreatePageT(traces, cursor)

	return &page, nil
}

func (svc *Service) PeriodicCleanup() error {
	tx, err := svc.db.Beginx()
	if err != nil {
		return fmt.Errorf("")
	}
	n, err := tx.Exec(`DELETE FROM traces WHERE created_at < (NOW() - $1::INTERVAL)`, "1 day")
	if err != nil {
		var rollbackErr error
		if rbErr := tx.Rollback(); rbErr != nil {
			rollbackErr = fmt.Errorf("while rolling back transaction: %w", rbErr)
		}
		err = errors.Join(err, rollbackErr)
		return fmt.Errorf("failed to delete traces: %w", err)
	} else {
		n, _ := n.RowsAffected()
		log.Printf("Cleaned %d traces\n", n)
	}

	n, err = tx.Exec(`DELETE FROM trace_steps WHERE queue_time < (NOW() - $1::INTERVAL)`, "1 day")
	if err != nil {
		var rollbackErr error
		if rbErr := tx.Rollback(); rbErr != nil {
			rollbackErr = fmt.Errorf("while rolling back transaction: %w", rbErr)
		}
		err = errors.Join(err, rollbackErr)
		return fmt.Errorf("failed to delete trace steps: %w", err)
	} else {
		n, _ := n.RowsAffected()
		log.Printf("Cleaned %d trace steps\n", n)
	}

	n, err = tx.Exec(`DELETE FROM trace_ingress WHERE archived_at < (NOW() - $1::INTERVAL)`, "1 day")
	if err != nil {
		var rollbackErr error
		if rbErr := tx.Rollback(); rbErr != nil {
			rollbackErr = fmt.Errorf("while rolling back transaction: %w", rbErr)
		}
		err = errors.Join(err, rollbackErr)
		return fmt.Errorf("failed to delete trace ingress: %w", err)
	} else {
		n, _ := n.RowsAffected()
		log.Printf("Cleaned %d trace ingresses\n", n)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("in periodic cleanu while committing transaction: %w", err)
	}

	_, err = svc.db.Exec("VACUUM traces, trace_steps, trace_ingress;")
	if err != nil {
		return fmt.Errorf("failed to vacuum tables: %w", err)
	} else {
		log.Println("Vacuumed tables")
	}

	return nil
}
