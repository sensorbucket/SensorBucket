package assetmanager

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidAssetURN = errors.New("Invalid URN")

	URN_PREFIX = "srn:pipelineasset"
)

// AssetURN URN specifically for assets in an asset manager
type AssetURN struct {
	PipelineName string
	AssetType    string
	AssetID      string
}

func (a *AssetURN) String() string {
	return strings.Join([]string{URN_PREFIX, a.PipelineName, a.AssetType, a.AssetID}, ":")
}

func ParseAssetURN(urn string) (AssetURN, error) {
	if !strings.HasPrefix(urn, URN_PREFIX) {
		return AssetURN{}, fmt.Errorf("%w: an asset URN must start with srn:pipelineasset", ErrInvalidAssetURN)
	}

	parts := strings.Split(urn, ":")
	if len(parts) != 5 {
		return AssetURN{}, fmt.Errorf("%w: an asset URN must have the following format %s:<pipeline>:<asset>:<id>", ErrInvalidAssetURN, URN_PREFIX)
	}

	return AssetURN{
		PipelineName: parts[2],
		AssetType:    parts[3],
		AssetID:      parts[4],
	}, nil
}
