package httpfilter_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sensorbucket.nl/sensorbucket/internal/httpfilter"
)

type TestStruct struct {
	A string   `url:"a"`
	B int      `url:"b"`
	C uint     `url:"c"`
	D float64  `url:"d"`
	E bool     `url:"e"`
	F []string `url:"f"`
}

func TestCreateAndFilter(t *testing.T) {
	filterCreator, err := httpfilter.Create[TestStruct]()
	if err != nil {
		t.Fatalf("Error creating filterCreator: %v", err)
	}

	testCases := []struct {
		name     string
		query    string
		expected TestStruct
	}{
		{
			name:     "Empty query",
			query:    "",
			expected: TestStruct{},
		},
		{
			name:  "All fields",
			query: "a=hello&b=-123&c=456&d=7.89&e=true&f=a&f=b&f=c",
			expected: TestStruct{
				A: "hello",
				B: -123,
				C: 456,
				D: 7.89,
				E: true,
				F: []string{"a", "b", "c"},
			},
		},
		{
			name:  "Partial fields",
			query: "a=test&b=100&d=0.1&f=1&f=2",
			expected: TestStruct{
				A: "test",
				B: 100,
				D: 0.1,
				F: []string{"1", "2"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, _ := url.ParseQuery(tc.query)
			var result TestStruct
			err := filterCreator(q, &result)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFilterInvalidValues(t *testing.T) {
	filterCreator, err := httpfilter.Create[TestStruct]()
	require.NoError(t, err)

	testCases := []struct {
		name     string
		query    string
		expected error
	}{
		{
			name:     "Invalid int value",
			query:    "b=text",
			expected: httpfilter.ErrConvertingString,
		},
		{
			name:     "Invalid uint value",
			query:    "c=text",
			expected: httpfilter.ErrConvertingString,
		},
		{
			name:     "Invalid float64 value",
			query:    "d=text",
			expected: httpfilter.ErrConvertingString,
		},
		{
			name:     "Invalid bool value",
			query:    "e=text",
			expected: httpfilter.ErrConvertingString,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, _ := url.ParseQuery(tc.query)
			var result TestStruct
			err := filterCreator(q, &result)
			assert.Error(t, err)
		})
	}
}

func TestCreateInvalidTypes(t *testing.T) {
	var err error
	_, err = httpfilter.Create[int]()
	assert.Error(t, err, "Should error if T in Create[T] is not a struct")
	_, err = httpfilter.Create[struct {
		A any
		B []any
		C struct{}
		D []struct{}
	}]()
	assert.Error(t, err, "Should not allow a 'any' for exported field")

	_, err = httpfilter.Create[struct {
		a any
		b struct{}
	}]()
	assert.NoError(t, err, "Should allow invalid values for non-exported fields")
}
