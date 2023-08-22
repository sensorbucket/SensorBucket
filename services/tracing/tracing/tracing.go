package tracing

type Step struct {
	TracingID      string `db:"tracing_id"`
	StepIndex      int64  `db:"step_index"`
	StepsRemaining int64  `db:"steps_remaining"`
	StartTime      int64  `db:"start_time"`
	Error          string `db:"error"`
}
