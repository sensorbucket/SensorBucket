package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/samber/lo"
	"sensorbucket.nl/sensorbucket/pkg/api"
)

const (
	// Email templates
	BaseEmailTemplate = "email-templates/base_inlined.html"

	DataExpiredEmailTemplate = "email-templates/expired.html"
	DataStuckEmailTemplate   = "email-templates/stuck.html"
	DataErredEmailTemplate   = "email-templates/errors.html"

	// Status codes for data points
	Pending    = 2
	InProgress = 4
	Failed     = 5

	MaxRecords = 1000
)

type mailer interface {
	SendMail(subject string, from string, to string, templateHtml []string, content interface{}) error
}

type store interface {
	DeleteExpiredData() error
}

type apiClient interface {
	ListTraces(ctx context.Context) api.ApiListTracesRequest
	ListIngresses(ctx context.Context) api.ApiListIngressesRequest
}

type service struct {
	mailer    mailer
	store     store
	apiClient apiClient
}

func newService() *service {
	return &service{}
}

// Will cleanup any data that have an expired expiration date
func (s *service) clean() error {
	log.Println("running cleanup...")
	return s.store.DeleteExpiredData()
}

// Will send a warning email to configured recipients that their data is scheduled for deletion
func (s *service) warnExpired(from string, recipient string, daysTillDeletion int) error {
	log.Println("warning recipients of scheduled data cleanup...")
	return s.mailer.SendMail(
		"SensorBucket - Geplande gegevensopruiming",
		from,
		recipient,
		[]string{
			BaseEmailTemplate,
			DataExpiredEmailTemplate,
		},
		mailContent{
			Title: "Geplande gegevensopruiming",
			Body: struct {
				DaysTillDeletion int
				DeletionDate     string
			}{
				DeletionDate:     time.Now().Add(time.Hour * 24 * time.Duration(daysTillDeletion)).Format("2 January 2006"),
				DaysTillDeletion: daysTillDeletion,
			},
		})
}

// Will send a warning mail to configured recipients if a recurring error exists for their datapoints
func (s *service) warnRecurringErrors(from string, recipient string, threshold int, checkLastHours int) error {
	traces, resp, err := s.apiClient.ListTraces(context.Background()).
		Status(Failed).
		StartTime(time.Now().Add(-time.Duration(checkLastHours * int(time.Hour)))).
		Limit(MaxRecords).
		Execute()
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	if len(traces.Data) == 0 {
		log.Printf("no errors detected, exiting\n")
		return nil
	}

	groupedByErrors := lo.GroupBy(traces.Data, func(item api.Trace) string {
		return errorFromSteps(item.Steps)
	})

	type recurringError struct {
		Error  string
		Amount int
		Since  string
	}

	res := []recurringError{}

	for key, value := range groupedByErrors {
		if key == "" || len(value) < threshold {
			continue
		}

		firstDetected := lo.MinBy(value, func(a, b api.Trace) bool {
			return b.StartTime.Before(a.StartTime)
		})

		res = append(res, recurringError{
			Error:  key,
			Amount: len(value),
			Since:  firstDetected.StartTime.Format(time.RFC3339),
		})
	}

	if len(res) == 0 {
		log.Printf("not enough recurring errors detected to pass the threshold ('%d'), not warning any users\n", threshold)
		return nil
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Amount > res[j].Amount
	})

	resultFiltered := lo.Slice(res, 0, 10)

	return s.mailer.SendMail(
		"SensorBucket - Fouten gedetecteerd",
		from,
		recipient,
		[]string{
			BaseEmailTemplate,
			DataErredEmailTemplate,
		},
		mailContent{
			Title: "Fouten gedetecteerd",
			Body: struct {
				Rows []recurringError
			}{
				resultFiltered,
			},
		},
	)
}

// Will send a warning mail to configured recipients if one or more datapoints is stuck in the pipeline for longer than the configured amount
func (s *service) warnStuck(from string, recipient string, timeout int32, checkLastHours int) error {
	log.Println("warning recipients of stuck data")

	// Find all datapoints that are pending or in progress for more than the configured amount
	traces, resp, err := s.apiClient.ListTraces(context.Background()).
		Status(Pending).
		Status(InProgress).
		StartTime(time.Now().Add(-time.Duration(checkLastHours * int(time.Hour)))).
		Limit(MaxRecords).
		Execute()
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if len(traces.Data) == 0 {
		// There is no data stuck in the pipeline according to the given parameters
		log.Printf("no stuck data detected, exiting\n")
		return nil
	}

	tracingIds := lo.SliceToMap(traces.Data, func(item api.Trace) (string, struct{}) {
		return item.TracingId, struct{}{}
	})

	// Now that the traces have been retrieved we need information on the pipelines for each trace
	ingresses, resp, err := s.apiClient.
		ListIngresses(context.Background()).
		Limit(int32(len(traces.Data))).
		TracingId(lo.Keys(tracingIds)).
		Execute()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if len(ingresses.Data) == 0 {
		return fmt.Errorf("couldn't find pipelines belonging to stuck datapoints")
	}

	type tableRow struct {
		Pipeline string
		Amount   int
		Since    string
	}

	pipelinesByTraceIds := lo.SliceToMap(ingresses.Data, func(item api.ArchivedIngress) (string, string) {
		return item.GetTracingId(), item.IngressDto.GetPipelineId()
	})

	statTable := []tableRow{}

	for tracingId, pipelineId := range pipelinesByTraceIds {
		filtered := lo.Filter(traces.Data, func(item api.Trace, index int) bool {
			return item.TracingId == tracingId
		})

		stuckLongest := lo.MinBy(filtered, func(a, b api.Trace) bool {
			return b.StartTime.Before(a.StartTime)
		})

		statTable = append(statTable, tableRow{
			Amount:   len(filtered),
			Pipeline: pipelineId,
			Since:    stuckLongest.StartTime.Format(time.RFC3339),
		})
	}

	sort.Slice(statTable, func(i, j int) bool {
		return statTable[i].Amount > statTable[j].Amount
	})

	resultFiltered := lo.Slice(statTable, 0, 10)

	return s.mailer.SendMail(
		"SensorBucket - Vastgelopen gegevens",
		from,
		recipient,
		[]string{
			BaseEmailTemplate,
			DataStuckEmailTemplate,
		},
		mailContent{
			Title: "Vastgelopen gegevens",
			Body: struct {
				Rows []tableRow
			}{
				resultFiltered,
			},
		},
	)
}

func errorFromSteps(steps []api.TraceStep) string {
	erred, ok := lo.Find(steps, func(item api.TraceStep) bool {
		return item.Error != ""
	})
	if ok {
		return erred.Error
	}
	return ""
}

type mailContent struct {
	Title string
	Body  any
}
