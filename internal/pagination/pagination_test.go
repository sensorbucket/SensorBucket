package pagination

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
)

var pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func Test(t *testing.T) {
	cols := []string{"a", "b", "c"}
	vals := []any{1, 2, 3}
	q := pq.Select("*").From("employees").Where(multiColumnCompare(cols, vals))
	query, _, _ := q.ToSql()

	t.Log(query)
}
