package dbutils

import (
	"database/sql"
	"errors"
)

var ErrTooManyRows = errors.New("too many rows returned")

// RowToFunc is a function that scans or otherwise converts row to a T.
type RowToFunc[T any] func(row *sql.Rows) (T, error)

// AppendRows iterates through rows, calling fn for each row, and appending the results into a slice of T.
//
// This function closes the rows automatically on return.
func AppendRows[T any, S ~[]T](slice S, rows *sql.Rows, fn RowToFunc[T]) (S, error) {
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		value, err := fn(rows)
		if err != nil {
			return nil, err
		}
		slice = append(slice, value)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return slice, nil
}

// CollectRows iterates through rows, calling fn for each row, and collecting the results into a slice of T.
//
// This function closes the rows automatically on return.
func CollectRows[T any](rows *sql.Rows, fn RowToFunc[T]) ([]T, error) {
	return AppendRows([]T{}, rows, fn)
}

// CollectOneRow calls fn for the first row in rows and returns the result. If no rows are found returns an error where errors.Is(ErrNoRows) is true.
// CollectOneRow is to CollectRows as QueryRow is to Query.
//
// This function closes the rows automatically on return.
func CollectOneRow[T any](rows *sql.Rows, fn RowToFunc[T]) (T, error) {
	defer func() {
		_ = rows.Close()
	}()

	var value T
	var err error

	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return value, err
		}
		return value, sql.ErrNoRows
	}

	value, err = fn(rows)
	if err != nil {
		return value, err
	}

	_ = rows.Close()
	return value, rows.Err()
}

// CollectExactlyOneRow calls fn for the first row in rows and returns the result.
//   - If no rows are found returns an error where errors.Is(ErrNoRows) is true.
//   - If more than 1 row is found returns an error where errors.Is(ErrTooManyRows) is true.
//
// This function closes the rows automatically on return.
func CollectExactlyOneRow[T any](rows *sql.Rows, fn RowToFunc[T]) (T, error) {
	defer func() {
		_ = rows.Close()
	}()

	var (
		err   error
		value T
	)

	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return value, err
		}

		return value, sql.ErrNoRows
	}

	value, err = fn(rows)
	if err != nil {
		return value, err
	}

	if rows.Next() {
		var zero T

		return zero, ErrTooManyRows
	}

	return value, rows.Err()
}
