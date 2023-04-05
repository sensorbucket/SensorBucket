package pagination

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/fxamacker/cbor/v2"
)

type Order string

const (
	ASC  Order = "ASC"
	DESC Order = "DESC"
)

type Field struct {
	Column string
	Order  Order
}

// Fields determines what fields the query cursor is based on.
// an example: { "created_at": pagination.ASC, "id": pagination.ASC }
// TODO: The order of keys is important, this should probably be an array of some sorts
type Fields []Field

func (f Fields) toOrder() string {
	orders := make([]string, len(f))
	for ix, v := range f {
		orders[ix] = v.Column + " " + string(v.Order)
	}
	return strings.Join(orders, ", ")
}

type Defaults struct {
	defaultLimit uint64
	maxLimit     uint64
	fields       Fields
}

func NewDefaults(defaultLimit, maxLimit uint64, fields Fields) *Defaults {
	return &Defaults{
		defaultLimit: defaultLimit,
		maxLimit:     maxLimit,
		fields:       fields,
	}
}

type Links struct {
	Previous string `json:"previous"`
	Next     string `json:"next"`
}

type APIResponse[T any] struct {
	Links      Links `json:"links"`
	PageSize   int   `json:"page_size"`
	TotalCount int   `json:"total_count"`
	Data       T     `json:"data"`
}

type Page[T any] struct {
	Cursor string
	Data   T
}

type Request struct {
	Cursor string `json:"cursor"`
	Limit  uint64 `json:"limit"`
}

func GetCursor[T any](r Request) Cursor[T] {
	// Normalize request limit size
	if r.Limit == 0 {
		r.Limit = 100
	} else if r.Limit > 250 {
		r.Limit = 250
	}

	// Case 1: New request, cursor empty
	if r.Cursor == "" {
		return Cursor[T]{Columns: *new(T), Limit: r.Limit}
	}

	// Case 2: Cursor given,
	c := DecodeCursor[T](r.Cursor)
	// Normalize cursor limit size
	if r.Limit == 0 {
		r.Limit = 100
	} else if r.Limit > 250 {
		r.Limit = 250
	}

	return c
}

type Cursor[T any] struct {
	Limit   uint64
	Columns T
}

func CreatePageT[T1 any, T2 any](data []T1, cursor Cursor[T2]) Page[[]T1] {
	var cursorString string
	if len(data) == int(cursor.Limit) {
		cursorString = EncodeCursor(cursor)
	}
	page := Page[[]T1]{
		Cursor: cursorString,
		Data:   data,
	}
	return page
}

func EncodeCursor[T any](f Cursor[T]) string {
	data, err := cbor.Marshal(&f)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

func DecodeCursor[T any](cursor string) Cursor[T] {
	data, _ := base64.RawURLEncoding.DecodeString(cursor)
	var t Cursor[T]
	if cursor == "" {
		return t
	}
	if err := cbor.Unmarshal(data, &t); err != nil {
		panic(err)
	}
	return t
}

type whereCol struct {
	column string
	value  any
	order  string
}

func Apply[T any](q squirrel.SelectBuilder, c Cursor[T]) (squirrel.SelectBuilder, error) {
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
			return q, fmt.Errorf("invalid pagination tag on struct %s, for field %s\n", rt.Name(), rf.Name)
		}

		column, order := tagParts[0], strings.ToUpper(tagParts[1])
		if order != "ASC" && order != "DESC" {
			return q, fmt.Errorf("invalid order in pagination tag on struct %s, for field %s\n", rt.Name(), rf.Name)
		}
		q = q.OrderBy(column + " " + order).Column(column)

		rvf := rv.Field(ix)
		if rvf.IsZero() {
			continue
		}

		columns = append(columns, whereCol{
			column: column,
			order:  order,
			value:  rvf.Interface(),
		})
	}

	for ix, v := range columns {
		if ix == len(columns)-1 {
			if v.order == "ASC" {
				q = q.Where(squirrel.Gt{v.column: v.value})
			} else {
				q = q.Where(squirrel.Lt{v.column: v.value})
			}
			continue
		}
		if v.order == "ASC" {
			q = q.Where(squirrel.GtOrEq{v.column: v.value})
		} else {
			q = q.Where(squirrel.LtOrEq{v.column: v.value})
		}
	}

	return q, nil
}
