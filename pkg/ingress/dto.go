package ingress

import "time"

type DTO struct {
	PipelineID     string
	Authentication string
	Timestamp      time.Time
	Payload        []byte
}

func (dto DTO) Validate() error {
	return nil
}
