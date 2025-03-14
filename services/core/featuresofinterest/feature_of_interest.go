package featuresofinterest

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/twpayne/go-geom"
)

type FeatureOfInterest struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	EncodingType    string `json:"encoding_type"`
	GeometryFeature *Geometry
	Feature         []byte
	Properties      json.RawMessage `json:"properties"`
	TenantID        int64           `json:"tenant_id"`
}

type CreateFeatureOfInterestOpts struct {
	Name         string
	Description  *string
	EncodingType *string
	Feature      json.RawMessage
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
	}
	if opts.Feature != nil {
		if err := foi.SetFeature(foi.EncodingType, foi.Feature); err != nil {
			return nil, err
		}
	}

	return &foi, nil
}

func (foi *FeatureOfInterest) SetFeature(encoding string, feature any) error {
	if g, ok := feature.(geom.T); ok {
		foi.GeometryFeature.T = g
		foi.EncodingType = mimeGeoJSON
	}

	switch {
	case strings.EqualFold(mimeGeoJSON, encoding):
		d, ok := feature.([]byte)
		if !ok {
			return fmt.Errorf("SetFeature expected []byte but got %T", feature)
		}
		if err := foi.GeometryFeature.UnmarshalJSON(d); err != nil {
			return err
		}
		return nil
	default:
		d, ok := feature.([]byte)
		if !ok {
			return fmt.Errorf("SetFeature expected []byte but got %T", feature)
		}
		foi.EncodingType = encoding
		foi.Feature = d
	}
	return nil
}

type foiJSONModel struct {
	ID           int64           `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	EncodingType string          `json:"encoding_type"`
	Feature      any             `json:"feature"`
	Properties   json.RawMessage `json:"properties"`
	TenantID     int64           `json:"tenant_id"`
}

func (foi *FeatureOfInterest) MarshalJSON() ([]byte, error) {
	model := foiJSONModel{
		ID:           foi.ID,
		Name:         foi.Name,
		Description:  foi.Description,
		EncodingType: foi.EncodingType,
		// Feature: ,
		Properties: foi.Properties,
		TenantID:   foi.TenantID,
	}

	if foi.GeometryFeature != nil {
		model.Feature = foi.GeometryFeature
	} else {
		model.Feature = foi.Feature
	}

	return json.Marshal(model)
}
