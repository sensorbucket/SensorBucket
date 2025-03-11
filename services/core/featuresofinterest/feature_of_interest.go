package featuresofinterest

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

type FeatureOfInterest struct {
	ID           int64           `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	EncodingType string          `json:"encoding_type"`
	Feature      *geom.Point     `json:"feature"`
	Properties   json.RawMessage `json:"properties"`
	TenantID     int64           `json:"tenant_id"`
}

type CreateFeatureOfInterestOpts struct {
	Name         string
	Description  *string
	EncodingType *string
	Feature      *json.RawMessage
	Properties   *json.RawMessage
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
		foi.Properties = *opts.Properties
	}

	if opts.EncodingType != nil && opts.Feature != nil {
		if err := foi.SetFeature(*opts.EncodingType, *opts.Feature); err != nil {
			return nil, err
		}
	} else if (opts.EncodingType != nil && opts.Feature == nil) || (opts.EncodingType == nil && opts.EncodingType != nil) {
		return nil, fmt.Errorf("in NewFeatureOfInterest: both encoding type and feature must be given, not or")
	}

	return &foi, nil
}

func (foi *FeatureOfInterest) SetFeature(encoding string, feature json.RawMessage) error {
	var g geom.T

	if strings.EqualFold(encoding, "application/geo+json") {
		// data, ok := feature.(json.RawMessage)
		// if !ok {
		// 	return fmt.Errorf("expected feature with encoding 'application/geo+json' to be json.RawMessage but got: %T", feature)
		// }
		if err := geojson.Unmarshal(feature, &g); err != nil {
			return fmt.Errorf("could not decode feature as geojson: %w", err)
		}
	} else {
		return fmt.Errorf("unsupported feature encoding type: %s", encoding)
	}

	point, ok := g.(*geom.Point)
	if !ok {
		return fmt.Errorf("expected feature to be a single point, but got: %T", g)
	}

	foi.Feature = point

	return nil
}
