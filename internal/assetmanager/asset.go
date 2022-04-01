package assetmanager

import "encoding/json"

// Asset ...
type Asset struct {
	ID       string          `json:"id"`
	Pipeline string          `json:"pipeline"`
	Type     string          `json:"type"`
	Content  json.RawMessage `json:"content"`
}

func (a *Asset) URN() string {
	urn := AssetURN{
		PipelineName: a.Pipeline,
		AssetType:    a.Type,
		AssetID:      a.ID,
	}
	return urn.String()
}

// AssetSchema ...
type AssetSchema struct {
	Type   string
	Schema json.RawMessage
}
