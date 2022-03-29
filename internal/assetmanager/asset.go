package assetmanager

import "encoding/json"

// Asset ...
type Asset struct {
	URN     string          `json:"id"`
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}

// AssetSchema ...
type AssetSchema struct {
	Type   string
	Schema json.RawMessage
}

func newAssetURN(assetType string, assetID string) AssetURN {
	return AssetURN{
		AssetType: assetType,
		AssetID:   assetID,
	}
}
