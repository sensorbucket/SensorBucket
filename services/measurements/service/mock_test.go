// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package service_test

import (
	"sensorbucket.nl/sensorbucket/services/measurements/service"
	"sync"
)

// Ensure, that StoreMock does implement service.Store.
// If this is not the case, regenerate this file with moq.
var _ service.Store = &StoreMock{}

// StoreMock is a mock implementation of service.Store.
//
//	func TestSomethingThatUsesStore(t *testing.T) {
//
//		// make and configure a mocked service.Store
//		mockedStore := &StoreMock{
//			CreateDatastreamFunc: func(datastream *service.Datastream) error {
//				panic("mock out the CreateDatastream method")
//			},
//			FindDatastreamFunc: func(sensorID int64, observedProperty string) (*service.Datastream, error) {
//				panic("mock out the FindDatastream method")
//			},
//			InsertFunc: func(measurement service.Measurement) error {
//				panic("mock out the Insert method")
//			},
//			ListDatastreamsFunc: func() ([]service.Datastream, error) {
//				panic("mock out the ListDatastreams method")
//			},
//			QueryFunc: func(query service.Query, pagination service.Pagination) ([]service.Measurement, *service.Pagination, error) {
//				panic("mock out the Query method")
//			},
//		}
//
//		// use mockedStore in code that requires service.Store
//		// and then make assertions.
//
//	}
type StoreMock struct {
	// CreateDatastreamFunc mocks the CreateDatastream method.
	CreateDatastreamFunc func(datastream *service.Datastream) error

	// FindDatastreamFunc mocks the FindDatastream method.
	FindDatastreamFunc func(sensorID int64, observedProperty string) (*service.Datastream, error)

	// InsertFunc mocks the Insert method.
	InsertFunc func(measurement service.Measurement) error

	// ListDatastreamsFunc mocks the ListDatastreams method.
	ListDatastreamsFunc func() ([]service.Datastream, error)

	// QueryFunc mocks the Query method.
	QueryFunc func(query service.Filter, pagination service.Pagination) ([]service.Measurement, *service.Pagination, error)

	// calls tracks calls to the methods.
	calls struct {
		// CreateDatastream holds details about calls to the CreateDatastream method.
		CreateDatastream []struct {
			// Datastream is the datastream argument value.
			Datastream *service.Datastream
		}
		// FindDatastream holds details about calls to the FindDatastream method.
		FindDatastream []struct {
			// SensorID is the sensorID argument value.
			SensorID int64
			// ObservedProperty is the observedProperty argument value.
			ObservedProperty string
		}
		// Insert holds details about calls to the Insert method.
		Insert []struct {
			// Measurement is the measurement argument value.
			Measurement service.Measurement
		}
		// ListDatastreams holds details about calls to the ListDatastreams method.
		ListDatastreams []struct {
		}
		// Query holds details about calls to the Query method.
		Query []struct {
			// Query is the query argument value.
			Query service.Filter
			// Pagination is the pagination argument value.
			Pagination service.Pagination
		}
	}
	lockCreateDatastream sync.RWMutex
	lockFindDatastream   sync.RWMutex
	lockInsert           sync.RWMutex
	lockListDatastreams  sync.RWMutex
	lockQuery            sync.RWMutex
}

// CreateDatastream calls CreateDatastreamFunc.
func (mock *StoreMock) CreateDatastream(datastream *service.Datastream) error {
	if mock.CreateDatastreamFunc == nil {
		panic("StoreMock.CreateDatastreamFunc: method is nil but Store.CreateDatastream was just called")
	}
	callInfo := struct {
		Datastream *service.Datastream
	}{
		Datastream: datastream,
	}
	mock.lockCreateDatastream.Lock()
	mock.calls.CreateDatastream = append(mock.calls.CreateDatastream, callInfo)
	mock.lockCreateDatastream.Unlock()
	return mock.CreateDatastreamFunc(datastream)
}

// CreateDatastreamCalls gets all the calls that were made to CreateDatastream.
// Check the length with:
//
//	len(mockedStore.CreateDatastreamCalls())
func (mock *StoreMock) CreateDatastreamCalls() []struct {
	Datastream *service.Datastream
} {
	var calls []struct {
		Datastream *service.Datastream
	}
	mock.lockCreateDatastream.RLock()
	calls = mock.calls.CreateDatastream
	mock.lockCreateDatastream.RUnlock()
	return calls
}

// FindDatastream calls FindDatastreamFunc.
func (mock *StoreMock) FindDatastream(sensorID int64, observedProperty string) (*service.Datastream, error) {
	if mock.FindDatastreamFunc == nil {
		panic("StoreMock.FindDatastreamFunc: method is nil but Store.FindDatastream was just called")
	}
	callInfo := struct {
		SensorID         int64
		ObservedProperty string
	}{
		SensorID:         sensorID,
		ObservedProperty: observedProperty,
	}
	mock.lockFindDatastream.Lock()
	mock.calls.FindDatastream = append(mock.calls.FindDatastream, callInfo)
	mock.lockFindDatastream.Unlock()
	return mock.FindDatastreamFunc(sensorID, observedProperty)
}

// FindDatastreamCalls gets all the calls that were made to FindDatastream.
// Check the length with:
//
//	len(mockedStore.FindDatastreamCalls())
func (mock *StoreMock) FindDatastreamCalls() []struct {
	SensorID         int64
	ObservedProperty string
} {
	var calls []struct {
		SensorID         int64
		ObservedProperty string
	}
	mock.lockFindDatastream.RLock()
	calls = mock.calls.FindDatastream
	mock.lockFindDatastream.RUnlock()
	return calls
}

// Insert calls InsertFunc.
func (mock *StoreMock) Insert(measurement service.Measurement) error {
	if mock.InsertFunc == nil {
		panic("StoreMock.InsertFunc: method is nil but Store.Insert was just called")
	}
	callInfo := struct {
		Measurement service.Measurement
	}{
		Measurement: measurement,
	}
	mock.lockInsert.Lock()
	mock.calls.Insert = append(mock.calls.Insert, callInfo)
	mock.lockInsert.Unlock()
	return mock.InsertFunc(measurement)
}

// InsertCalls gets all the calls that were made to Insert.
// Check the length with:
//
//	len(mockedStore.InsertCalls())
func (mock *StoreMock) InsertCalls() []struct {
	Measurement service.Measurement
} {
	var calls []struct {
		Measurement service.Measurement
	}
	mock.lockInsert.RLock()
	calls = mock.calls.Insert
	mock.lockInsert.RUnlock()
	return calls
}

// ListDatastreams calls ListDatastreamsFunc.
func (mock *StoreMock) ListDatastreams() ([]service.Datastream, error) {
	if mock.ListDatastreamsFunc == nil {
		panic("StoreMock.ListDatastreamsFunc: method is nil but Store.ListDatastreams was just called")
	}
	callInfo := struct {
	}{}
	mock.lockListDatastreams.Lock()
	mock.calls.ListDatastreams = append(mock.calls.ListDatastreams, callInfo)
	mock.lockListDatastreams.Unlock()
	return mock.ListDatastreamsFunc()
}

// ListDatastreamsCalls gets all the calls that were made to ListDatastreams.
// Check the length with:
//
//	len(mockedStore.ListDatastreamsCalls())
func (mock *StoreMock) ListDatastreamsCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockListDatastreams.RLock()
	calls = mock.calls.ListDatastreams
	mock.lockListDatastreams.RUnlock()
	return calls
}

// Query calls QueryFunc.
func (mock *StoreMock) Query(query service.Filter, pagination service.Pagination) ([]service.Measurement, *service.Pagination, error) {
	if mock.QueryFunc == nil {
		panic("StoreMock.QueryFunc: method is nil but Store.Query was just called")
	}
	callInfo := struct {
		Query      service.Filter
		Pagination service.Pagination
	}{
		Query:      query,
		Pagination: pagination,
	}
	mock.lockQuery.Lock()
	mock.calls.Query = append(mock.calls.Query, callInfo)
	mock.lockQuery.Unlock()
	return mock.QueryFunc(query, pagination)
}

// QueryCalls gets all the calls that were made to Query.
// Check the length with:
//
//	len(mockedStore.QueryCalls())
func (mock *StoreMock) QueryCalls() []struct {
	Query      service.Filter
	Pagination service.Pagination
} {
	var calls []struct {
		Query      service.Filter
		Pagination service.Pagination
	}
	mock.lockQuery.RLock()
	calls = mock.calls.Query
	mock.lockQuery.RUnlock()
	return calls
}

// Ensure, that DatastreamFinderCreaterMock does implement service.DatastreamFinderCreater.
// If this is not the case, regenerate this file with moq.
var _ service.DatastreamFinderCreater = &DatastreamFinderCreaterMock{}

// DatastreamFinderCreaterMock is a mock implementation of service.DatastreamFinderCreater.
//
//	func TestSomethingThatUsesDatastreamFinderCreater(t *testing.T) {
//
//		// make and configure a mocked service.DatastreamFinderCreater
//		mockedDatastreamFinderCreater := &DatastreamFinderCreaterMock{
//			CreateDatastreamFunc: func(datastream *service.Datastream) error {
//				panic("mock out the CreateDatastream method")
//			},
//			FindDatastreamFunc: func(sensorID int64, observedProperty string) (*service.Datastream, error) {
//				panic("mock out the FindDatastream method")
//			},
//		}
//
//		// use mockedDatastreamFinderCreater in code that requires service.DatastreamFinderCreater
//		// and then make assertions.
//
//	}
type DatastreamFinderCreaterMock struct {
	// CreateDatastreamFunc mocks the CreateDatastream method.
	CreateDatastreamFunc func(datastream *service.Datastream) error

	// FindDatastreamFunc mocks the FindDatastream method.
	FindDatastreamFunc func(sensorID int64, observedProperty string) (*service.Datastream, error)

	// calls tracks calls to the methods.
	calls struct {
		// CreateDatastream holds details about calls to the CreateDatastream method.
		CreateDatastream []struct {
			// Datastream is the datastream argument value.
			Datastream *service.Datastream
		}
		// FindDatastream holds details about calls to the FindDatastream method.
		FindDatastream []struct {
			// SensorID is the sensorID argument value.
			SensorID int64
			// ObservedProperty is the observedProperty argument value.
			ObservedProperty string
		}
	}
	lockCreateDatastream sync.RWMutex
	lockFindDatastream   sync.RWMutex
}

// CreateDatastream calls CreateDatastreamFunc.
func (mock *DatastreamFinderCreaterMock) CreateDatastream(datastream *service.Datastream) error {
	if mock.CreateDatastreamFunc == nil {
		panic("DatastreamFinderCreaterMock.CreateDatastreamFunc: method is nil but DatastreamFinderCreater.CreateDatastream was just called")
	}
	callInfo := struct {
		Datastream *service.Datastream
	}{
		Datastream: datastream,
	}
	mock.lockCreateDatastream.Lock()
	mock.calls.CreateDatastream = append(mock.calls.CreateDatastream, callInfo)
	mock.lockCreateDatastream.Unlock()
	return mock.CreateDatastreamFunc(datastream)
}

// CreateDatastreamCalls gets all the calls that were made to CreateDatastream.
// Check the length with:
//
//	len(mockedDatastreamFinderCreater.CreateDatastreamCalls())
func (mock *DatastreamFinderCreaterMock) CreateDatastreamCalls() []struct {
	Datastream *service.Datastream
} {
	var calls []struct {
		Datastream *service.Datastream
	}
	mock.lockCreateDatastream.RLock()
	calls = mock.calls.CreateDatastream
	mock.lockCreateDatastream.RUnlock()
	return calls
}

// FindDatastream calls FindDatastreamFunc.
func (mock *DatastreamFinderCreaterMock) FindDatastream(sensorID int64, observedProperty string) (*service.Datastream, error) {
	if mock.FindDatastreamFunc == nil {
		panic("DatastreamFinderCreaterMock.FindDatastreamFunc: method is nil but DatastreamFinderCreater.FindDatastream was just called")
	}
	callInfo := struct {
		SensorID         int64
		ObservedProperty string
	}{
		SensorID:         sensorID,
		ObservedProperty: observedProperty,
	}
	mock.lockFindDatastream.Lock()
	mock.calls.FindDatastream = append(mock.calls.FindDatastream, callInfo)
	mock.lockFindDatastream.Unlock()
	return mock.FindDatastreamFunc(sensorID, observedProperty)
}

// FindDatastreamCalls gets all the calls that were made to FindDatastream.
// Check the length with:
//
//	len(mockedDatastreamFinderCreater.FindDatastreamCalls())
func (mock *DatastreamFinderCreaterMock) FindDatastreamCalls() []struct {
	SensorID         int64
	ObservedProperty string
} {
	var calls []struct {
		SensorID         int64
		ObservedProperty string
	}
	mock.lockFindDatastream.RLock()
	calls = mock.calls.FindDatastream
	mock.lockFindDatastream.RUnlock()
	return calls
}