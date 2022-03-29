package assetmanager

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidAssetURN = errors.New("Invalid URN")

	ASSET_URN_ID_LENGTH = 16
)

// AssetURN URN specifically for assets in an asset manager
type AssetURN struct {
	PipelineName string
	AssetType    string
	AssetID      string
}

func (a *AssetURN) String() string {
	return "srn:pipelineasset" + a.PipelineName + ":" + a.AssetType + ":" + a.AssetID
}

func ParseAssetURN(urn string) (AssetURN, error) {
	if !strings.HasPrefix(urn, "srn:pipelineasset") {
		return AssetURN{}, fmt.Errorf("%w: an asset URN must start with srn:pipelineasset", ErrInvalidAssetURN)
	}

	parts := strings.Split(urn, ":")
	if len(parts) != 4 {
		return AssetURN{}, fmt.Errorf("%w: an asset URN must have the following format srn:pipelineasset:<pipeline>:<asset>:<id>", ErrInvalidAssetURN)
	}

	return AssetURN{
		PipelineName: parts[2],
		AssetType:    parts[3],
		AssetID:      parts[4],
	}, nil
}

// urnGenerator ...
type urnGenerator struct{}

func newURNGenerator() *urnGenerator {
	return &urnGenerator{}
}

func (u *urnGenerator) Generate(asset *Asset) (AssetURN, error) {
	return newAssetURN(asset.Type, randomString(ASSET_URN_ID_LENGTH)), nil
}

var randomChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	for i, v := range b {
		b[i] = randomChars[v%byte(len(randomChars))]
	}
	return string(b)
}
