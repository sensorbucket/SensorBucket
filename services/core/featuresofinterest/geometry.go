package featuresofinterest

import (
	"database/sql/driver"
	"fmt"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
	"github.com/twpayne/go-geom/encoding/geojson"
)

const mimeGeoJSON = "application/geo+json"

type Geometry struct {
	T geom.T
}

func (g *Geometry) Scan(src any) error {
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("tried scanning non byte value to Geometry")
	}
	if len(b) == 0 {
		return nil
	}
	var err error
	g.T, err = ewkb.Unmarshal(b)
	return err
}

func (g Geometry) Value() (driver.Value, error) {
	data, err := ewkb.Marshal(g.T, ewkb.NDR)
	if err != nil {
		return nil, fmt.Errorf("could not marshal geometry to T: %w", err)
	}
	return data, nil
}

func (g *Geometry) MarshalJSON() ([]byte, error) {
	return geojson.Marshal(g.T)
}

func (g *Geometry) UnmarshalJSON(src []byte) error {
	return geojson.Unmarshal(src, &g.T)
}
