package tracing

type Step struct {
	TracingID      string `pg:"tracing_id"`
	StepIndex      int64  `pg:"step_index"`
	StepsRemaining int64  `pg:"steps_remaining"`
	StartTime      int64  `pg:"start_time"`
	Error          string `pg:"error"`
}
