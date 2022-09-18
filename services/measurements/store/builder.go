package store

import (
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
)

type insertBuilder struct {
	values map[string]any
}

func newInsertBuilder() *insertBuilder {
	return &insertBuilder{
		values: map[string]any{},
	}
}

func (ib *insertBuilder) SetThingURN(thingURN string) *insertBuilder {
	ib.values["thing_urn"] = thingURN
	return ib
}
func (ib *insertBuilder) SetTimestamp(timestamp time.Time) *insertBuilder {
	ib.values["timestamp"] = timestamp
	return ib
}
func (ib *insertBuilder) SetValue(value float64) *insertBuilder {
	ib.values["value"] = value
	return ib
}
func (ib *insertBuilder) SetMeasurementType(typ, unit string) *insertBuilder {
	ib.values["measurement_type"] = typ
	ib.values["measurement_type_unit"] = unit
	return ib
}
func (ib *insertBuilder) SetMetadata(metadata json.RawMessage) *insertBuilder {
	ib.values["metadata"] = metadata
	return ib
}
func (ib *insertBuilder) TrySetLocation(id *int64, name *string, longitude, latitude *float64) *insertBuilder {
	if id == nil || name == nil || longitude == nil || latitude == nil {
		return ib
	}
	ib.values["location_id"] = *id
	ib.values["location_name"] = *name
	ib.values["location_coordinates"] = squirrel.Expr("ST_SETSRID(ST_POINT(?,?),4326)", *longitude, *latitude)
	return ib
}
func (ib *insertBuilder) TrySetCoordinates(longitude, latitude *float64) *insertBuilder {
	if longitude == nil || latitude == nil {
		return ib
	}
	ib.values["coordinates"] = squirrel.Expr("ST_SETSRID(ST_POINT(?,?),4326)", longitude, latitude)
	return ib
}
func (ib *insertBuilder) Build() (string, []any, error) {
	return pq.Insert("measurements").SetMap(ib.values).ToSql()
}
