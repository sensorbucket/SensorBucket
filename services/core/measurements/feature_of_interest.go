package measurements

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
	"sensorbucket.nl/sensorbucket/internal/web"
)

var ErrFeatureOfInterestNotFound = web.NewError(http.StatusNotFound, "The requested feature of interest was not found", "FEATURE_OF_INTEREST_NOT_FOUND")

type FeatureOfInterest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TenantID    int64  `json:"tenant_id"`
}

type FeatureOfInterestBinding struct {
	DatastreamID        uuid.UUID    `json:"datastream_id"`
	FeatureOfInterestID int64        `json:"feature_of_interest_id"`
	BoundAt             time.Time    `json:"bound_at"`
	UnboundAt           sql.NullTime `json:"unbound_at"`
}
