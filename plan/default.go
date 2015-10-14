package plan

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/context"
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
		panic("Unable to search, node is not a Process")
	}

	// width search at root level first
	for _, n := range root {
		nproc, ok := n.proc.(api.Process)
		if !ok {
			panic("Unable to search, node is not a Process")
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
		panic("Update failed, node not found")
	}
	found.nodes = append(found.nodes, branch...)
	return t
}

// walks the tree and exec f() for each node
func walk(t tree, f func(*node)) {
	if t == nil {
		panic("Unable to walk nil tree")
	}
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
func (n *node) To(sinks ...interface{}) *node {
	for _, sink := range sinks {
		dest, ok := sink.(api.Sink)
		if !ok {
			panic("To() param must be a Sink")
		}
		n.nodes = append(n.nodes, &node{proc: dest})
	}
	return n
}

// From is a builder function used to create a node
// of form From(proc).To(proc)
func From(from interface{}) *node {
	src, ok := from.(api.Source)
	if !ok {
		panic("From() param must be a Source")
	}
	return &node{proc: src}
}

// Conf is a configuration struct for creating new Plan.
type Conf struct {
	PlanName string
	Log      *logrus.Entry
	Ctx      context.Context
}

// DefaultPlan is an implementation of the Plan type.
// It allows the construction of a process graph using a simple
// fluent API.
// Flow(From(src).To(snk))
type DefaultPlan struct {
	tree        tree
	ctx         context.Context
	auxChan     chan interface{}
	auxPlan     *DefaultPlan
	auxEndpoint interface{}
}

// New creates a default plan that can be used
// to assemble process nodes and execute them.
func New(conf Conf) *DefaultPlan {
	log := func() *logrus.Entry {
		if conf.Log == nil {
			entry := logrus.WithField("Plan", "Default")
			if conf.PlanName != "" {
				entry = logrus.WithField("Plan", conf.PlanName)
			}
			conf.Log = entry
		}
		return conf.Log
	}()

	auxChan := make(chan interface{}, 1024)
	ctx := func() context.Context {
		if conf.Ctx == nil {
			ctx := autoctx.WithLogEntry(context.TODO(), log)
			ctx = autoctx.WithAuxChan(ctx, auxChan)
			return ctx
		}
		return conf.Ctx
	}()

	return &DefaultPlan{
		ctx:     ctx,
		tree:    make(tree, 0),
		auxChan: auxChan,
	}
}

// Flow is the building block used to assemble plan
// elements using the idiom:
// plan.Flow(From(<proc1>).To(<proc2>))
func (p *DefaultPlan) Flow(n *node) *DefaultPlan {
	// validate node before admitting
	if n == nil {
		panic("Flow() param is nil")
	}
	if n.proc == nil {
		panic("Flow() param does not resolve to a Process")
	}
	if n.nodes == nil || len(n.nodes) == 0 {
		panic("Flow must from source node to child node(s)")
	}

	// validate types for parent node/children
	_, ok := n.proc.(api.Source)
	if !ok {
		panic("Flow must start with a Source node")
	}

	for _, sink := range n.nodes {
		_, ok := sink.proc.(api.Sink)
		if !ok {
			panic("Children nodes must be Sink node(s)")
		}
	}

	p.tree = graph(p.tree, n) // insert new node
	return p
}

// WithAuxPlan accepts an auxiliary plan that is launched
// to process items on the auxiliary channel after the main plan is
// done. Set either WithAuxPlan or WithAuxEndpoint but not both.
func (p *DefaultPlan) WithAuxPlan(auxPlan *DefaultPlan) *DefaultPlan {
	p.auxPlan = auxPlan
	return p
}

// WithAuxEndpoint accepts an endpoint processor that will be used to
// process items from the auxiliary channel after the main plan is done.
// Set either WithAuxPlan or WithAuxEndpoint but not both.
func (p *DefaultPlan) WithAuxEndpoint(endpoint interface{}) *DefaultPlan {
	if _, ok := endpoint.(api.Endpoint); !ok {
		panic("WithAuxEndpoint param must implement api.Endpoint")
	}
	if sink, ok := endpoint.(api.Sink); ok {
		sink.SetInput(p.AuxChan())
	} else {
		panic("WithAuxEndpoint param must implement api.Sink")
	}
	p.auxEndpoint = endpoint
	return p
}

// AuxChan returns a readonly instance of the auxiliary channel
// This channel can be used as input to a separate flow for futher processing
// of auxialiary items (see DefaultPlan.WithAuxPlan)
func (p *DefaultPlan) AuxChan() <-chan interface{} {
	return p.auxChan
}

func (p *DefaultPlan) init() {
	if p == nil {
		panic("Internal node tree is nil")
	}
	walk(
		p.tree,
		func(n *node) {
			nproc, ok := n.proc.(api.Process)
			if !ok {
				panic(fmt.Sprintf("Init failed, expects Process type, got %T", nproc))
			}
			if err := nproc.Init(p.ctx); err != nil {
				panic(fmt.Sprintf("Unable to init node: %s", err))
			}
			if n.nodes != nil {
				srcproc, ok := n.proc.(api.Source)
				if !ok {
					panic("Nodes with children are expected to be Source")
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

	// walk tree and execute each process
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

	// go and complete plan
	go func() {
		defer close(done)
		wg.Wait()

		// handle auxiliary chan
		//close(p.auxChan)
		p.processAux()
	}()

	return done
}

func (p *DefaultPlan) processAux() {
	// process auxiliary plan if any
	if p.auxPlan != nil {
		close(p.auxChan)
		wait := make(chan struct{})
		go func() {
			<-p.auxPlan.Exec()
			close(wait)
		}()
		<-wait
	}

	// process auxliary endpoint if any
	if p.auxEndpoint != nil {
		close(p.auxChan)
		proc, ok := p.auxEndpoint.(api.Process)
		if !ok {
			panic("Aux endpoint must be an api.Process")
		}
		if err := proc.Init(p.ctx); err != nil {
			panic(err)
		}
		if err := proc.Exec(p.ctx); err != nil {
			panic(err)
		}
		ep, ok := p.auxEndpoint.(api.Endpoint)
		if !ok {
			panic("Aux endpoint must be an api.Endpoint")
		}
		<-ep.Done()
	}

}
