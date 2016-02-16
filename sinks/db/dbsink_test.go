package db

import (
	"testing"

	"github.com/vladimirvivien/automi/testutil"
)

func TestDbSink_New(t *testing.T) {
	db := &testutil.FakeDB{}
	dbsink := New().WithDB(db)
	if dbsink.db == nil {
		t.Fatal("DbSink not setting DB struct")
	}
}
