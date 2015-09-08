package steps

import (
	"testing"

	"xor/automi/api"
)

// This is to test ideas of automi's structure

func TestCreateStepGraph(t *testing.T) {
	s1 := &NoOp{Name: "S1"}
	s2 := &NoOp{Name: "S2", Input: s1}
	s3 := &NoOp{Name: "s3", Input: s2}
	count := 0
	testWalk(s2, &count)
	if count != 2 {
		t.Log("Expecting tree with 2")
		t.Fail()
	}
	count = 0
	testWalk(s3, &count)
	if count != 3 {
		t.Errorf("Expecting 3, got %d", count)
	}
}

func testWalk(s api.Step, c *int) {
	if s.GetInput() != nil {
		*c = *c + 1
		testWalk(s.GetInput(), c)
	} else {
		*c = *c + 1
		return
	}
}
