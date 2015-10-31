package plan

import (
	"testing"

	"github.com/vladimirvivien/automi/sup"
)

func TestDefaultPlan_GraphSimpleSearch(t *testing.T) {
	tree := &node{
		proc: &sup.NoopProc{Name: "PA"},
		nodes: []*node{
			&node{proc: &sup.NoopProc{Name: "PB"}},
			&node{proc: &sup.NoopProc{Name: "P1"}},
			&node{proc: &sup.NoopProc{Name: "P2"}},
			&node{
				proc: &sup.NoopProc{Name: "P3"},
				nodes: []*node{
					&node{proc: &sup.NoopProc{Name: "P4"}},
				},
			},
		},
	}

	found := search(tree, "PA")
	if found == nil {
		t.Fatal("tree search failed")
	}

	found = search(tree, "P1")
	if found == nil {
		t.Fatal("tree search failed")
	}

	found = search(tree, "PB")
	if found == nil {
		t.Fatal("tree search failed")
	}
	found = search(tree, "P4")
	if found == nil {
		t.Fatal("tree search failed")
	}

	found = search(tree, "PZ")
	if found != nil {
		t.Fatal("tree search failed, no match should be found")
	}

}

func TestDefaultPlan_GraphGraph(t *testing.T) {
	tree := &node{
		proc: &sup.NoopProc{Name: "P1"},
		nodes: []*node{
			&node{proc: &sup.NoopProc{Name: "P2"}},
		},
	}

	tree = graph(tree, &node{
		proc: &sup.NoopProc{Name: "P1"},
		nodes: []*node{
			&node{proc: &sup.NoopProc{Name: "P3"}},
			&node{proc: &sup.NoopProc{Name: "P4"}},
		},
	})

	found := search(tree, "P1")

	if found == nil {
		t.Fatal("Search should not have failed.")
	}
	if len(found.nodes) != 2 {
		t.Fatal("Expected node with 2 children")
	}

    found = search(tree, "P4")
 	if found == nil {
		t.Fatal("tree search failed")
	}

	tree = graph(tree, &node{proc: &sup.NoopProc{Name: "P2"}})	
	found = search(tree, "P2")
 	if found == nil {
		t.Fatal("tree search failed")
	}
}

func TestDefaultPlan_GraphWalk(t *testing.T) {
	tree := &node{
		proc: &sup.NoopProc{Name: "P1"},
		nodes: []*node{
			&node{proc: &sup.NoopProc{Name: "P2"}},
			&node{proc: &sup.NoopProc{Name: "P3"}},
			&node{
				proc:  &sup.NoopProc{Name: "P4"},
				nodes: []*node{&node{proc: &sup.NoopProc{Name: "P5"}}},
			},
		},
	}

	count := 0
	walk(tree, func(n *node) { count++ })
	if count != 5 {
		t.Fatal("Walk failed to visit all expected nodes")
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