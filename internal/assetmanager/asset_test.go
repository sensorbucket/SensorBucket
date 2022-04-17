package assetmanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateIDFromPKShouldReplaceAllFields(t *testing.T) {
	content := []byte(`{"field_a": "value_a", "field_b": "value_b"}`)
	id, err := generateIDFromPK("random_${field_a}_something_${field_b}", content)
	assert.NoError(t, err)
	assert.Equal(t, "random_value_a_something_value_b", id)
}

func BenchmarkIDFromPK(b *testing.B) {
	content := []byte(`{"field_a": "value_a", "field_b": "value_b"}`)
	for i := 0; i < b.N; i++ {
		_, err := generateIDFromPK("random_${field_a}_something_${field_b}", content)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestNewAssetDefinitionShouldErrorIfPrimaryKeyInvalid(t *testing.T) {
	testCases := []struct {
		desc string
		pk   string
	}{
		{
			desc: "Error if PK contains space",
			pk:   "some_${eui} thing",
		},
		{
			desc: "Error if PK contains no template",
			pk:   "some_thing",
		},
		{
			desc: "Error if PK contains colon",
			pk:   "some:${eui}",
		},
		{
			desc: "Error if PK contains hashtag",
			pk:   "some_${eui}#thing",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			_, err := newAssetDefinition(newAssetDefinitionOpts{
				Name:       "test_definition",
				PipelineID: "test_pipeline",
				Labels:     []string{},
				Version:    1,
				PrimaryKey: tC.pk,
				Schema:     []byte(`{"type":"object"}`),
			})
			assert.Error(t, err)
		})
	}
}

func TestNewAssetShouldUsePrimaryKey(t *testing.T) {
	ad, err := newAssetDefinition(newAssetDefinitionOpts{
		Name:       "test_definition",
		PipelineID: "test_pipeline",
		Labels:     []string{},
		Version:    1,
		PrimaryKey: "test_key",
		Schema: []byte(`{
			"type": "object",
			"properties": {
				"test_key": {
					"type": "string"
				}
			}
		}`),
	})
	assert.NoError(t, err)

	asset, err := newAsset(ad, []byte(`{"test_key": "test_value"}`))
	assert.NoError(t, err)
	assert.Equal(t, "test_value", asset.ID)
}

func TestNewAssetShouldFallbackToRandomIDIfPrimaryKeySetButNotFound(t *testing.T) {
	ad, err := newAssetDefinition(newAssetDefinitionOpts{
		Name:       "test_definition",
		PipelineID: "test_pipeline",
		Labels:     []string{},
		Version:    1,
		PrimaryKey: "test_key_but_different",
		Schema: []byte(`{
			"type": "object",
			"properties": {
				"test_key": {
					"type": "string"
				}
			}
		}`),
	})
	assert.NoError(t, err)

	asset, err := newAsset(ad, []byte(`{"test_key": "test_value"}`))
	assert.NoError(t, err)
	assert.NotEqual(t, "test_value", asset.ID)
	assert.Len(t, asset.ID, ASSET_ID_LENGTH)
}

func TestNewAssetShouldFallbackToRandomIDIfNoPrimaryKey(t *testing.T) {
	ad, err := newAssetDefinition(newAssetDefinitionOpts{
		Name:       "test_definition",
		PipelineID: "test_pipeline",
		Labels:     []string{},
		Version:    1,
		// PrimaryKey: "",
		Schema: []byte(`{
			"type": "object",
			"properties": {
				"test_key": {
					"type": "string"
				}
			}
		}`),
	})
	assert.NoError(t, err)

	asset, err := newAsset(ad, []byte(`{"test_key": "test_value"}`))
	assert.NoError(t, err)
	assert.NotEqual(t, "test_value", asset.ID)
	assert.Len(t, asset.ID, ASSET_ID_LENGTH)
}
