package featuresofinterest

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/twpayne/go-geom"
)

type FeatureOfInterest struct {
	ID           int64           `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	EncodingType string          `json:"encoding_type"`
	Feature      *Geometry       `json:"feature"`
	Properties   json.RawMessage `json:"properties"`
	TenantID     int64           `json:"tenant_id"`
}

type CreateFeatureOfInterestOpts struct {
	Name         string
	Description  *string
	EncodingType *string
	Feature      *Geometry
	Properties   json.RawMessage
	TenantID     int64
}

func NewFeatureOfInterest(opts CreateFeatureOfInterestOpts) (*FeatureOfInterest, error) {
	if opts.Name == "" || opts.TenantID == 0 {
		return nil, fmt.Errorf("In NewFeatureOfInterest: missing required Name or TenantID")
	}

	var foi FeatureOfInterest
	foi.Name = opts.Name
	foi.TenantID = opts.TenantID
	if opts.Description != nil {
		foi.Description = *opts.Description
	}
	if opts.Properties != nil {
		foi.Properties = opts.Properties
	} else {
		opts.Properties = json.RawMessage("{}")
	}
	if opts.Feature != nil && opts.EncodingType != nil {
		if err := foi.SetFeature(*opts.EncodingType, opts.Feature); err != nil {
			return nil, err
		}
	}

	return &foi, nil
}

func (foi *FeatureOfInterest) SetFeature(encoding string, feature any) error {
	if g, ok := feature.(geom.T); ok {
		foi.Feature = &Geometry{g}
		foi.EncodingType = mimeGeoJSON
		return nil
	}
	if g, ok := feature.(*Geometry); ok {
		foi.Feature = g
		foi.EncodingType = mimeGeoJSON
		return nil
	}

	switch {
	case strings.EqualFold(mimeGeoJSON, encoding):
		d, ok := feature.([]byte)
		if !ok {
			return fmt.Errorf("SetFeature expected []byte but got %T", feature)
		}
		if err := foi.Feature.UnmarshalJSON(d); err != nil {
			return err
		}
		foi.EncodingType = mimeGeoJSON
		return nil
	default:
		return fmt.Errorf("unknown encoding type for feature geometry: %s", encoding)
	}
}

func (foi *FeatureOfInterest) ClearFeature() {
	foi.EncodingType = ""
	foi.Feature = nil
}
