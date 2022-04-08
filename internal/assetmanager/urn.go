package assetmanager

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidAssetURN = errors.New("Invalid URN")

	URN_PREFIX = "srn:asset"
)

// AssetURN URN specifically for assets in an asset manager
type AssetURN struct {
	PipelineID string
	AssetType  string
	AssetID    string
}

func (u AssetURN) String() string {
	parts := []string{URN_PREFIX, u.PipelineID, u.AssetType}
	if u.AssetID != "" {
		parts = append(parts, u.AssetID)
	}

	return strings.Join(parts, ":")
}

func ParseAssetURN(urnString string) (AssetURN, error) {
	if !strings.HasPrefix(urnString, URN_PREFIX) {
		return AssetURN{}, fmt.Errorf("%w: an asset URN must start with %s", ErrInvalidAssetURN, URN_PREFIX)
	}

	parts := strings.Split(urnString, ":")
	if len(parts) < 4 {
		return AssetURN{}, fmt.Errorf("%w: an asset URN must have the following format %s:<pipeline>:<asset>:[id]", ErrInvalidAssetURN, URN_PREFIX)
	}

	urn := AssetURN{
		PipelineID: parts[2],
		AssetType:  parts[3],
	}

	// Append ID if it exists
	if len(parts) > 4 {
		urn.AssetID = parts[4]
	}

	return urn, nil
}
