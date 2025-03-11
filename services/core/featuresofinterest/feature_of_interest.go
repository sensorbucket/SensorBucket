package featuresofinterest

import (
	"encoding/json"
	"errors"
	"fmt"
)

type FeatureOfInterest struct {
	ID           int64           `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	EncodingType string          `json:"encoding_type"`
	Feature      any             `json:"feature"`
	Properties   json.RawMessage `json:"properties"`
	TenantID     int64           `json:"tenant_id"`
}

type CreateFeatureOfInterestOpts struct {
	Name         string
	Description  *string
	EncodingType *string
	Feature      any
	Properties   *json.RawMessage
	TenantID     int64
}

func NewFeatureOfInterest(opts CreateFeatureOfInterestOpts) (*FeatureOfInterest, error) {
	if opts.Name == "" || opts.TenantID == 0 {
		return nil, fmt.Errorf("In NewFeatureOfInterest: Missing required Name or TenantID")
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
		if err := foi.SetFeature(*opts.EncodingType, opts.Feature); err != nil {
			return nil, err
		}
	} else if (opts.EncodingType != nil && opts.Feature == nil) || (opts.EncodingType == nil && opts.EncodingType != nil) {
		return nil, fmt.Errorf("in NewFeatureOfInterest: both encoding type and feature must be given, not or")
	}

	return &foi, nil
}

func (foi *FeatureOfInterest) SetFeature(encoding string, feature any) error {
	// TODO: Implements featur on features-of-interest
	return errors.New("Not implemented")
}
