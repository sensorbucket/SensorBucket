package pagination

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/fxamacker/cbor/v2"
)

const (
	MAX_LIMIT     = 1000
	DEFAULT_LIMIT = 100
)

type Links struct {
	Previous string `json:"previous"`
	Next     string `json:"next"`
}

type APIResponse[T any] struct {
	Links      Links `json:"links"`
	PageSize   int   `json:"page_size"`
	TotalCount int   `json:"total_count"`
	Data       []T   `json:"data"`
}

type Page[T any] struct {
	Cursor string
	Data   []T
}

type Request struct {
	Cursor string `json:"cursor"`
	Limit  uint64 `json:"limit"`
}

func GetCursor[T any](r Request) (Cursor[T], error) {
	// Normalize request limit size
	if r.Limit == 0 {
		r.Limit = DEFAULT_LIMIT
	} else if r.Limit > MAX_LIMIT {
		r.Limit = MAX_LIMIT
	}

	// Case 1: New request, cursor empty
	if r.Cursor == "" {
		return Cursor[T]{Columns: *new(T), Limit: r.Limit}, nil
	}

	// Case 2: Cursor given,
	c, err := DecodeCursor[T](r.Cursor)
	if err != nil {
		return c, err
	}
	// Normalize cursor limit size
	if c.Limit == 0 {
		c.Limit = DEFAULT_LIMIT
	} else if c.Limit > MAX_LIMIT {
		c.Limit = MAX_LIMIT
	}

	return c, nil
}

type Cursor[T any] struct {
	Limit   uint64
	Columns T
}

func CreatePageT[T1 any, T2 any](data []T1, cursor Cursor[T2]) Page[T1] {
	var cursorString string

	// TODO: Small bug, at this point the data has already been Limited and the data will always be equal to the Limit or smaller
	// this means that a cursor will be made if the set is exactly equal to the limit but in reality, there is no more data
	// meaning a new cursor is sent that won't result in any data
	// https://github.com/sensorbucket/SensorBucket/issues/82
	if len(data) >= int(cursor.Limit) {
		cursorString = EncodeCursor(cursor)
	}
	page := Page[T1]{
		Cursor: cursorString,
		Data:   data,
	}
	return page
}

func EncodeCursor[T any](f Cursor[T]) string {
	opt := cbor.CanonicalEncOptions()
	opt.Time = cbor.TimeUnix
	enc, _ := opt.EncMode()

	data, err := enc.Marshal(&f)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

func DecodeCursor[T any](cursor string) (Cursor[T], error) {
	var t Cursor[T]

	data, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return t, err
	}
	if cursor == "" {
		return t, nil
	}
	if err := cbor.Unmarshal(data, &t); err != nil {
		return t, err
	}
	return t, nil
}

type whereCol struct {
	column string
	value  any
	order  string
}

func multiColumnCompare(columns []whereCol) sq.Sqlizer {
	if len(columns) == 0 {
		return nil
	}
	clause := sq.Or{}
	for i := 0; i < len(columns); i++ {
		and := sq.And{}
		for j := 0; j <= i; j++ {
			col := columns[j]
			if j == i {
				if col.order == "ASC" {
					and = append(and, sq.Gt{col.column: col.value})
				} else {
					and = append(and, sq.Lt{col.column: col.value})
				}
				continue
			}
			and = append(and, sq.Eq{col.column: col.value})
		}
		clause = append(clause, and)
	}
	return clause
}

func Apply[T any](q sq.SelectBuilder, c Cursor[T]) (sq.SelectBuilder, error) {
	q = q.Limit(c.Limit)
	rt := reflect.TypeOf(c.Columns)
	rv := reflect.ValueOf(c.Columns)
	columns := []whereCol{}
	for ix := 0; ix < rt.NumField(); ix++ {
		rf := rt.Field(ix)
		if !rf.IsExported() {
			continue
		}

		tag, ok := rf.Tag.Lookup("pagination")
		if !ok {
			continue
		}

		tagParts := strings.Split(tag, ",")
		if len(tagParts) != 2 {
			return q, fmt.Errorf("invalid pagination tag on struct %s, for field %s", rt.Name(), rf.Name)
		}

		column, order := tagParts[0], strings.ToUpper(tagParts[1])
		if order != "ASC" && order != "DESC" {
			return q, fmt.Errorf("invalid order in pagination tag on struct %s, for field %s", rt.Name(), rf.Name)
		}
		q = q.OrderBy(column + " " + order).Column(column)

		rvf := rv.Field(ix)
		if rvf.IsZero() {
			continue
		}

		columns = append(columns, whereCol{
			column: column,
			value:  rvf.Interface(),
			order:  order,
		})
	}

	q = q.Where(multiColumnCompare(columns))

	return q, nil
}

func CreateResponse[T any](r *http.Request, baseURL string, page Page[T]) APIResponse[T] {
	var next string
	if page.Cursor != "" {
		q := r.URL.Query()
		q.Set("cursor", page.Cursor)
		next = baseURL + r.URL.Path + "?" + q.Encode()
	}

	response := APIResponse[T]{
		Links: Links{
			Next: next,
		},
		PageSize: len(page.Data),
		Data:     page.Data,
	}
	return response
}

func CursorFromURL(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return u.Query().Get("cursor")
}
