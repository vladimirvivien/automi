package plan

import (
	"testing"
	"time"

	"github.com/vladimirvivien/automi/sup"
)

func TestDefaultPlan_Graph(t *testing.T) {
	tree := []*node{
		&node{
			proc: &sup.NoopProc{Name: "P1"},
			nodes: []*node{
				&node{proc: &sup.NoopProc{Name: "P2"}},
			},
		},
	}

	found := search(tree, &node{proc: &sup.NoopProc{Name: "P2"}})
	if found == nil {
		t.Fatal("tree search failed")
	}

	tree = graph(tree, &node{
		proc: &sup.NoopProc{Name: "P1"},
		nodes: []*node{
			&node{proc: &sup.NoopProc{Name: "P3"}},
			&node{proc: &sup.NoopProc{Name: "P4"}},
		},
	})

	found = search(tree, &node{proc: &sup.NoopProc{Name: "P1"}})

	if found == nil {
		t.Fatal("Search should not have failed.")
	}
	if len(found.nodes) != 2 {
		t.Fatal("Expected node with 2 children")
	}

	tree = update(
		tree,
		&node{proc: &sup.NoopProc{Name: "P4"}},
		&node{proc: &sup.NoopProc{Name: "P5"}},
	)
	found = search(tree, &node{proc: &sup.NoopProc{Name: "P4"}})

	if len(found.nodes) != 1 {
		t.Fatal("Expected children node not 1")
	}
}

func TestDefaultPlan_FromTo(t *testing.T) {
	node := From(&sup.NoopProc{Name: "P1"}).To(
		&sup.NoopProc{Name: "P2"},
		&sup.NoopProc{Name: "P3"},
	)
	if len(node.nodes) != 2 {
		t.Fatal("From.To builder not building node")
	}
}

func TestDefaultPlan_Flow(t *testing.T) {
	in := make(chan interface{})
	go func() {
		in <- 1
		in <- 2
		close(in)
	}()

	p1 := &sup.NoopProc{Name: "P1"}
	p1.SetInput(in)
	p2 := &sup.NoopProc{Name: "P2"}
	//p3 := sup.NoopProc{Name:"P3"}
	plan := New()
	node := From(p1).To(p2)
	plan.Flow(node)

	if p2.GetOutput() == nil {
		t.Fatal("Flow() not connecting components")
	}

	count := 0
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for _ = range p2.GetOutput() {
			count++
		}
	}()

	select {
	case <-wait:
		if count != 2 {
			t.Fatal("Flow not connecting properly. Expected 2 items, got", count)
		}
	case <-time.After(5 * time.Millisecond):
		t.Fatal("Waited too long, something is broken")
	}
}
