package testutil

import (
	"database/sql"
	"database/sql/driver"
)

type FakeDB struct{}

func (d *FakeDB) Begin() (*sql.Tx, error) { return nil, nil }
func (d *FakeDB) Close() error            { return nil }
func (d *FakeDB) Driver() driver.Driver   { return nil }
func (d *FakeDB) Exec(string, ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (d *FakeDB) Ping() error                       { return nil }
func (d *FakeDB) Prepare(string) (*sql.Stmt, error) { return nil, nil }
func (d *FakeDB) Query(string, ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (d *FakeDB) QueryRow(string, ...interface{}) *sql.Row {
	return nil
}
func (d *FakeDB) SetMaxIdleConns(int) {}
