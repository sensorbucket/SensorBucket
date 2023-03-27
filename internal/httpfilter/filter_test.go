package httpfilter_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sensorbucket.nl/sensorbucket/internal/httpfilter"
)

func TestCreateAndFilterPrimitives(t *testing.T) {
	type TestStruct struct {
		A string   `url:"a"`
		B int      `url:"b"`
		C uint     `url:"c"`
		D float64  `url:"d"`
		E bool     `url:"e"`
		F []string `url:"f"`
	}

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
	type TestStruct struct {
		A string   `url:"a"`
		B int      `url:"b"`
		C uint     `url:"c"`
		D float64  `url:"d"`
		E bool     `url:"e"`
		F []string `url:"f"`
		R string   `url:"r,required"`
	}

	filterCreator, err := httpfilter.Create[TestStruct]()
	require.NoError(t, err)

	testCases := []struct {
		name     string
		query    string
		expected error
	}{
		{
			name:     "Invalid int value",
			query:    "r=true&b=text",
			expected: httpfilter.ErrConvertingString,
		},
		{
			name:     "Invalid uint value",
			query:    "r=true&c=text",
			expected: httpfilter.ErrConvertingString,
		},
		{
			name:     "Invalid float64 value",
			query:    "r=true&d=text",
			expected: httpfilter.ErrConvertingString,
		},
		{
			name:     "Invalid bool value",
			query:    "r=true&e=text",
			expected: httpfilter.ErrConvertingString,
		},
		{
			name:     "Missing required value",
			query:    "",
			expected: httpfilter.ErrMissingParameter,
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
}

func TestFilterParseTime(t *testing.T) {
	type filter struct {
		Start time.Time
		End   time.Time
	}
	var f filter
	start := "2022-01-01T00:00:01Z"
	end := "2022-01-02T00:00:01Z"
	q := url.Values{
		"start": []string{start},
		"end":   []string{url.QueryEscape(end)},
	}
	expectedStart, _ := time.Parse(time.RFC3339, start)
	expectedEnd, _ := time.Parse(time.RFC3339, end)
	parseFilter, err := httpfilter.Create[filter]()
	require.NoError(t, err)

	err = parseFilter(q, &f)
	require.NoError(t, err)

	assert.Equal(t, expectedStart, f.Start)
	assert.Equal(t, expectedEnd, f.End)
}
