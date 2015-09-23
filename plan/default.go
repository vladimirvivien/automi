package plan

import (
	"fmt"

	"github.com/vladimirvivien/automi/api"
)

type node struct {
	proc  api.Processor
	nodes []*node
}

func (n *node) String() string {
	return fmt.Sprintf("%v -> %v", n.proc, n.nodes)
}

type tree []*node

// search bails after first match
func search(root tree, key *node) *node {
	if root == nil {
		return nil
	}
	// width search at root level first
	for _, n := range root {
		if key.proc.GetName() == n.proc.GetName() {
			return n
		}
	}
	// depth search
	for _, n := range root {
		if n != nil {
			return search(n.nodes, key)
		}
	}

	return nil
}

// finds existing node and overwrites it
func graph(t tree, n *node) tree {
	found := search(t, n)
	if found != nil {
		*found = *n
		return t
	}
	t = append(t, n)
	return t
}

// branch inserts node branch as a child of node
func update(t tree, node *node, branch ...*node) tree {
	found := search(t, node)
	if found == nil {
		panic("Branch op failed, node not found")
	}
	found.nodes = append(found.nodes, branch...)
	return t
}

// To used as part of the From().To() builder
func (n *node) To(sinks ...api.Processor) *node {
	for _, sink := range sinks {
		n.nodes = append(n.nodes, &node{proc: sink})
	}
	return n
}

// From is a builder function used to create a node
// of form From(proc).To(proc)
func From(from api.Processor) *node {
	return &node{proc: from}
}

// DefaultPlan is an implementation of the Plan type.
// It allows the construction of a process graph using a simple
// fluent API.
// Flow(From(src).To(snk))
type DefaultPlan struct {
	tree tree
}

func New() *DefaultPlan {
	return &DefaultPlan{tree: make(tree, 0)}
}

func (p *DefaultPlan) Flow(n *node) *DefaultPlan {
	// validate node before admitting
	if n == nil {
		panic("Node in flow is nil")
	}
	if n.proc == nil {
		panic("Node does not have an a process")
	}
	if n.nodes == nil || len(n.nodes) == 0 {
		panic("Node is not linked to other nodes")
	}

	// link components
	for _, sink := range n.nodes {
		sink.proc.SetInput(n.proc.GetOutput())
	}

	p.tree = graph(p.tree, n) // insert new node
	return p
}
