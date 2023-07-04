package processinginfra

import "github.com/jmoiron/sqlx"

type DB interface {
	Select(dest interface{}, query string, args ...interface{}) error
	sqlx.Execer
	sqlx.Queryer
}
