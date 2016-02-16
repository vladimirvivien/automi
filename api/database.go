package api

import (
	"database/sql"
	"database/sql/driver"
)

// Database is a stand-in interface that maps to database/sql.DB struct.
// This is a workaround to make it easier to test database related code
// without the of actual database trivers.
type Database interface {
	Begin() (*sql.Tx, error)
	Close() error
	Driver() driver.Driver
	Exec(string, ...interface{}) (sql.Result, error)
	Ping() error
	Prepare(string) (*sql.Stmt, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	SetMaxIdleConns(int)
}
