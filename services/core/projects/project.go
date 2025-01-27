package projects

type Project struct {
	ID                 int64
	Name               string
	Description        string
	FeaturesOfInterest ProjectFeatureOfInterest
	TenantID           int64
}

type ProjectFeatureOfInterest struct {
	FeatureOfInterest          FeatureOfInterest
	InterestedObservationTypes []string
}

type FeatureOfInterest struct {
	ID          int64
	Name        string
	Description string
	TenantID    int64
}
