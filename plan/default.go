package plan

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"

	"github.com/vladimirvivien/automi/api"
)

type node struct {
	proc  interface{}
	nodes []*node
}

func (n *node) String() string {
	return fmt.Sprintf("%v -> %v", n.proc, n.nodes)
}

type tree []*node

// search bails after first match
func search(root tree, key *node) *node {
	if root == nil || key == nil {
		panic("Unable to search for nil tree or key")
	}

	kproc, ok := key.proc.(api.Process)
	if !ok {
		panic("Unable to search, Node.proc not a valid Process")
	}

	// width search at root level first
	for _, n := range root {
		nproc, ok := n.proc.(api.Process)
		if !ok {
			panic("Unable to search, Node.proc not a valid Process")
		}
		if kproc.GetName() == nproc.GetName() {
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

// walks the tree and exec f() for each node
func walk(t tree, f func(*node)) {
	for _, n := range t {
		if f != nil {
			f(n)
		}
		if n.nodes != nil {
			walk(n.nodes, f)
		}
	}
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
	ctx  context.Context
}

// New creates a default plan that can be used
// to assemble process nodes and execute them.
func New() *DefaultPlan {
	return &DefaultPlan{
		ctx:  context.Background(),
		tree: make(tree, 0),
	}
}

// Flow is the building block used to assemble plan
// elements using the idiom:
// plan.Flow(From(<proc1>).To(<proc2>))
func (p *DefaultPlan) Flow(n *node) *DefaultPlan {
	// validate node before admitting
	if n == nil {
		panic("Node in flow is nil")
	}
	if n.proc == nil {
		panic("Node does not have an a process")
	}
	if n.nodes == nil || len(n.nodes) == 0 {
		panic("Node must flow to child node(s)")
	}

	// link proc with downstream
	_, ok := n.proc.(api.Source)
	if !ok {
		panic("Flow must start with a Source node")
	}

	// link components
	for _, sink := range n.nodes {
		_, ok := sink.proc.(api.Sink)
		if !ok {
			panic("Flow must flow into a Sink node")
		}
	}

	p.tree = graph(p.tree, n) // insert new node
	return p
}

func (p *DefaultPlan) init() {
	walk(
		p.tree,
		func(n *node) {
			nproc, ok := n.proc.(api.Process)
			if !ok {
				panic("Node not a process, unable to initialize")
			}
			if err := nproc.Init(p.ctx); err != nil {
				panic(fmt.Sprintf("Unable to init node: %s", err))
			}
			if n.nodes != nil {
				srcproc, ok := n.proc.(api.Source)
				if !ok {
					panic("Node with children are expected to be Source")
				}
				// link components
				for _, sink := range n.nodes {
					sinkproc, ok := sink.proc.(api.Sink)
					if !ok {
						panic("Children nodes are expected to be Sink nodes")
					}
					sinkproc.SetInput(srcproc.GetOutput())
				}
			}

		},
	)

}

// Exec walks the node graph and calls
// Init() and Exec() on each element
func (p *DefaultPlan) Exec() <-chan struct{} {
	defer func() {
		if re := recover(); re != nil {
			fmt.Println("[Failure]", re)
		}
	}()

	p.init() // initialize nodes

	// walk and look for leaves
	// ensure they are of type api.Endpoint
	// if so, collect them
	endpoints := make([]api.Endpoint, 0)
	walk(
		p.tree,
		func(n *node) {
			if n.nodes == nil {
				ep, ok := n.proc.(api.Endpoint)
				if !ok {
					panic("Leave nodes must be of type Endpoint")
				}
				endpoints = append(endpoints, ep)
			}
		},
	)

	if len(endpoints) == 0 {
		panic("Graph must be terminated with Endpoint nodes")
	}

	walk(
		p.tree,
		func(n *node) {
			nproc, ok := n.proc.(api.Process)
			if !ok {
				panic("Plan can only execute nodes that implement Process")
			}
			if err := nproc.Exec(p.ctx); err != nil {
				panic(fmt.Sprintf("Failed to Exec node %s: %s", nproc.GetName(), err))
			}
		},
	)

	// wait for end points to finish
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(len(endpoints))
	go func(wait *sync.WaitGroup) {
		for _, ep := range endpoints {
			<-ep.Done()
			wait.Done()
		}
	}(&wg)

	go func() {
		wg.Wait()
		close(done)
	}()

	return done
}
