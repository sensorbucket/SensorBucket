// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package measurements_test

import (
	"context"
	"github.com/google/uuid"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
	"sync"
)

// Ensure, that StoreMock does implement measurements.Store.
// If this is not the case, regenerate this file with moq.
var _ measurements.Store = &StoreMock{}

// StoreMock is a mock implementation of measurements.Store.
//
//	func TestSomethingThatUsesStore(t *testing.T) {
//
//		// make and configure a mocked measurements.Store
//		mockedStore := &StoreMock{
//			FindOrCreateDatastreamFunc: func(ctx context.Context, tenantID int64, sensorID int64, observedProperty string, UnitOfMeasurement string) (*measurements.Datastream, error) {
//				panic("mock out the FindOrCreateDatastream method")
//			},
//			GetDatastreamFunc: func(ctx context.Context, id uuid.UUID, filter measurements.DatastreamFilter) (*measurements.Datastream, error) {
//				panic("mock out the GetDatastream method")
//			},
//			ListDatastreamsFunc: func(contextMoqParam context.Context, datastreamFilter measurements.DatastreamFilter, request pagination.Request) (*pagination.Page[measurements.Datastream], error) {
//				panic("mock out the ListDatastreams method")
//			},
//			QueryFunc: func(contextMoqParam context.Context, filter measurements.Filter, request pagination.Request) (*pagination.Page[measurements.Measurement], error) {
//				panic("mock out the Query method")
//			},
//			StoreMeasurementFunc: func(contextMoqParam context.Context, measurement measurements.Measurement) error {
//				panic("mock out the StoreMeasurement method")
//			},
//			StoreMeasurementsFunc: func(contextMoqParam context.Context, measurementsMoqParam []measurements.Measurement) error {
//				panic("mock out the StoreMeasurements method")
//			},
//		}
//
//		// use mockedStore in code that requires measurements.Store
//		// and then make assertions.
//
//	}
type StoreMock struct {
	// FindOrCreateDatastreamFunc mocks the FindOrCreateDatastream method.
	FindOrCreateDatastreamFunc func(ctx context.Context, tenantID int64, sensorID int64, observedProperty string, UnitOfMeasurement string) (*measurements.Datastream, error)

	// GetDatastreamFunc mocks the GetDatastream method.
	GetDatastreamFunc func(ctx context.Context, id uuid.UUID, filter measurements.DatastreamFilter) (*measurements.Datastream, error)

	// ListDatastreamsFunc mocks the ListDatastreams method.
	ListDatastreamsFunc func(contextMoqParam context.Context, datastreamFilter measurements.DatastreamFilter, request pagination.Request) (*pagination.Page[measurements.Datastream], error)

	// QueryFunc mocks the Query method.
	QueryFunc func(contextMoqParam context.Context, filter measurements.Filter, request pagination.Request) (*pagination.Page[measurements.Measurement], error)

	// StoreMeasurementFunc mocks the StoreMeasurement method.
	StoreMeasurementFunc func(contextMoqParam context.Context, measurement measurements.Measurement) error

	// StoreMeasurementsFunc mocks the StoreMeasurements method.
	StoreMeasurementsFunc func(contextMoqParam context.Context, measurementsMoqParam []measurements.Measurement) error

	// calls tracks calls to the methods.
	calls struct {
		// FindOrCreateDatastream holds details about calls to the FindOrCreateDatastream method.
		FindOrCreateDatastream []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// TenantID is the tenantID argument value.
			TenantID int64
			// SensorID is the sensorID argument value.
			SensorID int64
			// ObservedProperty is the observedProperty argument value.
			ObservedProperty string
			// UnitOfMeasurement is the UnitOfMeasurement argument value.
			UnitOfMeasurement string
		}
		// GetDatastream holds details about calls to the GetDatastream method.
		GetDatastream []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ID is the id argument value.
			ID uuid.UUID
			// Filter is the filter argument value.
			Filter measurements.DatastreamFilter
		}
		// ListDatastreams holds details about calls to the ListDatastreams method.
		ListDatastreams []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// DatastreamFilter is the datastreamFilter argument value.
			DatastreamFilter measurements.DatastreamFilter
			// Request is the request argument value.
			Request pagination.Request
		}
		// Query holds details about calls to the Query method.
		Query []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// Filter is the filter argument value.
			Filter measurements.Filter
			// Request is the request argument value.
			Request pagination.Request
		}
		// StoreMeasurement holds details about calls to the StoreMeasurement method.
		StoreMeasurement []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// Measurement is the measurement argument value.
			Measurement measurements.Measurement
		}
		// StoreMeasurements holds details about calls to the StoreMeasurements method.
		StoreMeasurements []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// MeasurementsMoqParam is the measurementsMoqParam argument value.
			MeasurementsMoqParam []measurements.Measurement
		}
	}
	lockFindOrCreateDatastream sync.RWMutex
	lockGetDatastream          sync.RWMutex
	lockListDatastreams        sync.RWMutex
	lockQuery                  sync.RWMutex
	lockStoreMeasurement       sync.RWMutex
	lockStoreMeasurements      sync.RWMutex
}

// FindOrCreateDatastream calls FindOrCreateDatastreamFunc.
func (mock *StoreMock) FindOrCreateDatastream(ctx context.Context, tenantID int64, sensorID int64, observedProperty string, UnitOfMeasurement string) (*measurements.Datastream, error) {
	if mock.FindOrCreateDatastreamFunc == nil {
		panic("StoreMock.FindOrCreateDatastreamFunc: method is nil but Store.FindOrCreateDatastream was just called")
	}
	callInfo := struct {
		Ctx               context.Context
		TenantID          int64
		SensorID          int64
		ObservedProperty  string
		UnitOfMeasurement string
	}{
		Ctx:               ctx,
		TenantID:          tenantID,
		SensorID:          sensorID,
		ObservedProperty:  observedProperty,
		UnitOfMeasurement: UnitOfMeasurement,
	}
	mock.lockFindOrCreateDatastream.Lock()
	mock.calls.FindOrCreateDatastream = append(mock.calls.FindOrCreateDatastream, callInfo)
	mock.lockFindOrCreateDatastream.Unlock()
	return mock.FindOrCreateDatastreamFunc(ctx, tenantID, sensorID, observedProperty, UnitOfMeasurement)
}

// FindOrCreateDatastreamCalls gets all the calls that were made to FindOrCreateDatastream.
// Check the length with:
//
//	len(mockedStore.FindOrCreateDatastreamCalls())
func (mock *StoreMock) FindOrCreateDatastreamCalls() []struct {
	Ctx               context.Context
	TenantID          int64
	SensorID          int64
	ObservedProperty  string
	UnitOfMeasurement string
} {
	var calls []struct {
		Ctx               context.Context
		TenantID          int64
		SensorID          int64
		ObservedProperty  string
		UnitOfMeasurement string
	}
	mock.lockFindOrCreateDatastream.RLock()
	calls = mock.calls.FindOrCreateDatastream
	mock.lockFindOrCreateDatastream.RUnlock()
	return calls
}

// GetDatastream calls GetDatastreamFunc.
func (mock *StoreMock) GetDatastream(ctx context.Context, id uuid.UUID, filter measurements.DatastreamFilter) (*measurements.Datastream, error) {
	if mock.GetDatastreamFunc == nil {
		panic("StoreMock.GetDatastreamFunc: method is nil but Store.GetDatastream was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		ID     uuid.UUID
		Filter measurements.DatastreamFilter
	}{
		Ctx:    ctx,
		ID:     id,
		Filter: filter,
	}
	mock.lockGetDatastream.Lock()
	mock.calls.GetDatastream = append(mock.calls.GetDatastream, callInfo)
	mock.lockGetDatastream.Unlock()
	return mock.GetDatastreamFunc(ctx, id, filter)
}

// GetDatastreamCalls gets all the calls that were made to GetDatastream.
// Check the length with:
//
//	len(mockedStore.GetDatastreamCalls())
func (mock *StoreMock) GetDatastreamCalls() []struct {
	Ctx    context.Context
	ID     uuid.UUID
	Filter measurements.DatastreamFilter
} {
	var calls []struct {
		Ctx    context.Context
		ID     uuid.UUID
		Filter measurements.DatastreamFilter
	}
	mock.lockGetDatastream.RLock()
	calls = mock.calls.GetDatastream
	mock.lockGetDatastream.RUnlock()
	return calls
}

// ListDatastreams calls ListDatastreamsFunc.
func (mock *StoreMock) ListDatastreams(contextMoqParam context.Context, datastreamFilter measurements.DatastreamFilter, request pagination.Request) (*pagination.Page[measurements.Datastream], error) {
	if mock.ListDatastreamsFunc == nil {
		panic("StoreMock.ListDatastreamsFunc: method is nil but Store.ListDatastreams was just called")
	}
	callInfo := struct {
		ContextMoqParam  context.Context
		DatastreamFilter measurements.DatastreamFilter
		Request          pagination.Request
	}{
		ContextMoqParam:  contextMoqParam,
		DatastreamFilter: datastreamFilter,
		Request:          request,
	}
	mock.lockListDatastreams.Lock()
	mock.calls.ListDatastreams = append(mock.calls.ListDatastreams, callInfo)
	mock.lockListDatastreams.Unlock()
	return mock.ListDatastreamsFunc(contextMoqParam, datastreamFilter, request)
}

// ListDatastreamsCalls gets all the calls that were made to ListDatastreams.
// Check the length with:
//
//	len(mockedStore.ListDatastreamsCalls())
func (mock *StoreMock) ListDatastreamsCalls() []struct {
	ContextMoqParam  context.Context
	DatastreamFilter measurements.DatastreamFilter
	Request          pagination.Request
} {
	var calls []struct {
		ContextMoqParam  context.Context
		DatastreamFilter measurements.DatastreamFilter
		Request          pagination.Request
	}
	mock.lockListDatastreams.RLock()
	calls = mock.calls.ListDatastreams
	mock.lockListDatastreams.RUnlock()
	return calls
}

// Query calls QueryFunc.
func (mock *StoreMock) Query(contextMoqParam context.Context, filter measurements.Filter, request pagination.Request) (*pagination.Page[measurements.Measurement], error) {
	if mock.QueryFunc == nil {
		panic("StoreMock.QueryFunc: method is nil but Store.Query was just called")
	}
	callInfo := struct {
		ContextMoqParam context.Context
		Filter          measurements.Filter
		Request         pagination.Request
	}{
		ContextMoqParam: contextMoqParam,
		Filter:          filter,
		Request:         request,
	}
	mock.lockQuery.Lock()
	mock.calls.Query = append(mock.calls.Query, callInfo)
	mock.lockQuery.Unlock()
	return mock.QueryFunc(contextMoqParam, filter, request)
}

// QueryCalls gets all the calls that were made to Query.
// Check the length with:
//
//	len(mockedStore.QueryCalls())
func (mock *StoreMock) QueryCalls() []struct {
	ContextMoqParam context.Context
	Filter          measurements.Filter
	Request         pagination.Request
} {
	var calls []struct {
		ContextMoqParam context.Context
		Filter          measurements.Filter
		Request         pagination.Request
	}
	mock.lockQuery.RLock()
	calls = mock.calls.Query
	mock.lockQuery.RUnlock()
	return calls
}

// StoreMeasurement calls StoreMeasurementFunc.
func (mock *StoreMock) StoreMeasurement(contextMoqParam context.Context, measurement measurements.Measurement) error {
	if mock.StoreMeasurementFunc == nil {
		panic("StoreMock.StoreMeasurementFunc: method is nil but Store.StoreMeasurement was just called")
	}
	callInfo := struct {
		ContextMoqParam context.Context
		Measurement     measurements.Measurement
	}{
		ContextMoqParam: contextMoqParam,
		Measurement:     measurement,
	}
	mock.lockStoreMeasurement.Lock()
	mock.calls.StoreMeasurement = append(mock.calls.StoreMeasurement, callInfo)
	mock.lockStoreMeasurement.Unlock()
	return mock.StoreMeasurementFunc(contextMoqParam, measurement)
}

// StoreMeasurementCalls gets all the calls that were made to StoreMeasurement.
// Check the length with:
//
//	len(mockedStore.StoreMeasurementCalls())
func (mock *StoreMock) StoreMeasurementCalls() []struct {
	ContextMoqParam context.Context
	Measurement     measurements.Measurement
} {
	var calls []struct {
		ContextMoqParam context.Context
		Measurement     measurements.Measurement
	}
	mock.lockStoreMeasurement.RLock()
	calls = mock.calls.StoreMeasurement
	mock.lockStoreMeasurement.RUnlock()
	return calls
}

// StoreMeasurements calls StoreMeasurementsFunc.
func (mock *StoreMock) StoreMeasurements(contextMoqParam context.Context, measurementsMoqParam []measurements.Measurement) error {
	if mock.StoreMeasurementsFunc == nil {
		panic("StoreMock.StoreMeasurementsFunc: method is nil but Store.StoreMeasurements was just called")
	}
	callInfo := struct {
		ContextMoqParam      context.Context
		MeasurementsMoqParam []measurements.Measurement
	}{
		ContextMoqParam:      contextMoqParam,
		MeasurementsMoqParam: measurementsMoqParam,
	}
	mock.lockStoreMeasurements.Lock()
	mock.calls.StoreMeasurements = append(mock.calls.StoreMeasurements, callInfo)
	mock.lockStoreMeasurements.Unlock()
	return mock.StoreMeasurementsFunc(contextMoqParam, measurementsMoqParam)
}

// StoreMeasurementsCalls gets all the calls that were made to StoreMeasurements.
// Check the length with:
//
//	len(mockedStore.StoreMeasurementsCalls())
func (mock *StoreMock) StoreMeasurementsCalls() []struct {
	ContextMoqParam      context.Context
	MeasurementsMoqParam []measurements.Measurement
} {
	var calls []struct {
		ContextMoqParam      context.Context
		MeasurementsMoqParam []measurements.Measurement
	}
	mock.lockStoreMeasurements.RLock()
	calls = mock.calls.StoreMeasurements
	mock.lockStoreMeasurements.RUnlock()
	return calls
}
