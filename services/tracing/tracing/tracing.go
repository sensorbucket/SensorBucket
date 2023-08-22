package tracing

type Step struct {
	TracingID      string `pg:"tracing_id"`
	StepIndex      int64  `pg:"step_index"`
	StepsRemaining int64  `pg:"steps_remaining"`
	StartTime      int64  `pg:"start_time"`
	Error          string `pg:"error"`
}

type Status int

const (
	Unknown    Status = 0
	Canceled   Status = 1
	Pending    Status = 2
	Success    Status = 3
	InProgress Status = 4
	Failed     Status = 5
)

func (s *Status) String() string {
	switch *s {
	case Canceled:
		return "canceled"
	case Pending:
		return "pending"
	case Success:
		return "success"
	case InProgress:
		return "in progress"
	case Failed:
		return "failed"
	}
	return "Unknown"
}
