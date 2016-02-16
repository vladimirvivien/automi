package db

import (
	"testing"

	"github.com/vladimirvivien/automi/testutil"
)

func TestDbSink_New(t *testing.T) {
	db := &testutil.FakeDB{}
	dbsink := New().WithDB(db).
		Sql("SELECT * from table").
		Prepare(func(item interface{}) []interface{} {
		return nil
	})
	if dbsink.db == nil {
		t.Fatal("DbSink not setting DB struct")
	}
	if dbsink.sql == "" {
		t.Fatal("DbSink not setting sql")
	}
	if dbsink.prepFn == nil {
		t.Fatal("DbSink not setting prep function")
	}
}

func TestDbSink_Open(t *testing.T) {
	in := make(chan interface{})
	go func() {
		in <- []string{"Christophe", "Petion", "Dessaline"}
		in <- []string{"Toussaint", "Guerrier", "Caiman"}
		close(in)
	}()

	db := &testutil.FakeDB{}
	dbsink := New().WithDB(db).
		Sql("SELECT * from table").
		Prepare(func(item interface{}) []interface{} {
		data := item.([]string)
		if len(data) != 3 {
			t.Fatal("Prepare fn not getting correct args")
		}
		return []interface{}{data[0], data[1], data[2]}
	})
	dbsink.SetInput(in)

	// TODO - devise a better way to unit test db.
	//select {
	//case err := <-dbsink.Open(context.Background()):
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//case <-time.After(50 * time.Millisecond):
	//	t.Fatal("Sink took too long to open")
	//}
}
