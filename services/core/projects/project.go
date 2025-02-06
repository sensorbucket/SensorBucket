package projects

// Agg. Root

type Project struct {
	ID                 int64                      `json:"id"`
	Name               string                     `json:"name"`
	Description        string                     `json:"description"`
	FeaturesOfInterest []ProjectFeatureOfInterest `json:"features_of_interest"`
}

type ProjectFeatureOfInterest struct {
	FeatureOfInterest          FeatureOfInterest `json:"feature_of_interest"`
	InterestedObservationTypes []string          `json:"interested_observation_types"`
}

// Agg. Root

type FeatureOfInterest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
