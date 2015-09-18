package db

import (
	"database/sql"
	"testing"
)

func TestDbUpdate_Init(t *testing.T) {
	db := &DbUpdate{}
	if err := db.Init(); err == nil {
		t.Fatal("Error expected for missing attributes")
	}

	db = &DbUpdate{Name: "db"}
	if err := db.Init(); err == nil {
		t.Fatal("Error expected, missing attributes")
	}

	in := make(chan interface{})

	db = &DbUpdate{Name: "db", Input: in}
	if err := db.Init(); err == nil {
		t.Fatal("Error expected, missing attributes")
	}

	DB, _ := sql.Open("fake", "fake")
	db = &DbUpdate{Name: "db", Input: in, DB: DB}
	if err := db.Init(); err == nil {
		t.Fatal("Error expected, missing attributes")
	}
	db = &DbUpdate{
		Name:    "db",
		Input:   in,
		DB:      DB,
		SQL:     "DELTE from TBL where ? = ?",
		Prepare: func(d interface{}) []interface{} { return nil },
	}
	if err := db.Init(); err == nil {
		t.Fatal("Unexpected error:", err)
	}

	// TODO Look for better way to mock/test DB

}
