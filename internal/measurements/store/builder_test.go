package store

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestBuilderShouldNotIncludeLocationCoordinatesIfNil(t *testing.T) {
	for _, tc := range []struct {
		id        *int64
		name      *string
		longitude *float64
		latitude  *float64
		expect    string
	}{
		{
			id:        nil,
			name:      nil,
			longitude: nil,
			latitude:  nil,
			expect:    "INSERT INTO measurements VALUES ()",
		},
		{
			id:        lo.ToPtr(int64(1)),
			name:      lo.ToPtr("hello"),
			latitude:  lo.ToPtr(float64(2323)),
			longitude: nil,
			expect:    "INSERT INTO measurements VALUES ()",
		},
		{
			id:        lo.ToPtr(int64(1)),
			name:      lo.ToPtr("hello"),
			latitude:  lo.ToPtr(float64(2323)),
			longitude: lo.ToPtr(float64(2323)),
			expect:    "INSERT INTO measurements (location_coordinates,location_id,location_name) VALUES (ST_SETSRID(ST_POINT($1,$2),4326),$3,$4)",
		},
	} {
		query, _, _ := newInsertBuilder().
			TrySetLocation(tc.id, tc.name, tc.longitude, tc.latitude).Build()
		assert.Equal(t, tc.expect, query)
	}
}

func TestBuilderShouldNotIncludeCoordinatesIfNil(t *testing.T) {
	for _, tc := range []struct {
		longitude *float64
		latitude  *float64
		expect    string
	}{
		{
			longitude: nil,
			latitude:  nil,
			expect:    "INSERT INTO measurements VALUES ()",
		},
		{
			latitude:  lo.ToPtr(float64(2323)),
			longitude: nil,
			expect:    "INSERT INTO measurements VALUES ()",
		},
		{
			latitude:  lo.ToPtr(float64(2323)),
			longitude: lo.ToPtr(float64(2323)),
			expect:    "INSERT INTO measurements (coordinates) VALUES (ST_SETSRID(ST_POINT($1,$2),4326))",
		},
	} {
		query, _, _ := newInsertBuilder().
			TrySetCoordinates(tc.longitude, tc.latitude).Build()
		assert.Equal(t, tc.expect, query)
	}
}
