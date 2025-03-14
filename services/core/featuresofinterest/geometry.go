package featuresofinterest

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
	"github.com/twpayne/go-geom/encoding/geojson"
)

const mimeGeoJSON = "application/geo+json"

type Geometry struct {
	geom.T
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

func (g *Geometry) Value() (driver.Value, error) {
	sb := &strings.Builder{}
	if err := ewkb.Write(sb, ewkb.NDR, g); err != nil {
		return nil, err
	}
	return []byte(sb.String()), nil
}

func (g *Geometry) MarshalJSON() ([]byte, error) {
	return geojson.Marshal(g.T)
}

func (g *Geometry) UnmarshalJSON(src []byte) error {
	return geojson.Unmarshal(src, &g.T)
}
