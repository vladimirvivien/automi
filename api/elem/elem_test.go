package elem

import "testing"

func TestElem_New(t *testing.T) {
	e := New("Hello String")
	if e.Val != "Hello String" {
		t.Fatal("Value not set proper for element")
	}
}

func TestElem_Slice(t *testing.T) {
	elems := Slice(3, 2, "Dog")
	if len(elems) != 3 {
		t.Fatal("Missing element from slice")
	}
	if elems[2].String() != "Dog" {
		t.Fatal("Expected value failed")
	}
}

func TestElem_ToString(t *testing.T) {
	e := New("Hello")
	if e.String() != "Hello" {
		t.Fatal("Elem not carrying string val")
	}
}

func TestElem_ToInt64(t *testing.T) {
	e := New(int64(12))
	if e.Int64() != int64(12) {
		t.Fatal("Elem not carrying int64 type value")
	}
}
