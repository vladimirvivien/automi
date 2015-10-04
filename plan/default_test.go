package plan

import (
	"testing"
	"time"

	"golang.org/x/net/context"

	autoctx "github.com/vladimirvivien/automi/context"
	"github.com/vladimirvivien/automi/proc"
	"github.com/vladimirvivien/automi/sup"
)

type testProc struct {
	name     string
	initFunc func(context.Context)
	execFunc func(context.Context)
	input    <-chan interface{}
	output   chan interface{}
	done     chan struct{}
}

func (p *testProc) GetName() string {
	return p.name
}

func (p *testProc) Init(ctx context.Context) error {
	p.output = make(chan interface{})
	p.done = make(chan struct{})

	if p.initFunc != nil {
		p.initFunc(ctx)
	}

	return nil
}

func (p *testProc) Uninit(ctx context.Context) error {
	return nil
}

func (p *testProc) Exec(ctx context.Context) error {
	if p.execFunc != nil {
		p.execFunc(ctx)
	}
	go func() {
		defer func() {
			close(p.output)
			close(p.done)
		}()
		for in := range p.input {
			p.output <- in
		}
	}()
	return nil
}

func (p *testProc) GetOutput() <-chan interface{} {
	return p.output
}
func (p *testProc) SetInput(in <-chan interface{}) {
	p.input = in
}
func (p *testProc) Done() <-chan struct{} {
	return p.done
}

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

	tree = []*node{
		&node{
			proc: &sup.NoopProc{Name: "P1"},
			nodes: []*node{
				&node{proc: &sup.NoopProc{Name: "P2"}},
				&node{proc: &sup.NoopProc{Name: "P3"}},
				&node{
					proc:  &sup.NoopProc{Name: "P4"},
					nodes: []*node{&node{proc: &sup.NoopProc{Name: "P5"}}},
				},
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

func TestDefaultPlan_Flow(t *testing.T) {
	p1 := &sup.NoopProc{Name: "P1"}
	p2 := &sup.NoopProc{Name: "P2"}
	p3 := &sup.NoopProc{Name: "P3"}
	p4 := &sup.NoopProc{Name: "P4"}
	p5 := &sup.NoopProc{Name: "P5"}

	plan := New(Conf{})
	flow := From(p1).To(p2)
	plan.Flow(flow).Flow(From(p2).To(p3, p4, p5))

	if flow == nil {
		t.Fatal("From() not building prober node")
	}
	count := 0
	walk(
		plan.tree,
		func(n *node) {
			count++
		},
	)
	if count != 5 {
		t.Fatal("Unexpected node count of ", count)
	}

}

func TestDefaultPlan_Exec(t *testing.T) {
	in := make(chan interface{})
	go func() {
		in <- 1
		in <- 2
		close(in)
	}()

	callCount := 0
	p1 := &testProc{
		name: "P1",
		initFunc: func(ctx context.Context) {
			callCount++
		},
		execFunc: func(ctx context.Context) {
			callCount++
		},
		input: in,
	}
	p1.SetInput(in)
	p2 := &testProc{
		name: "P2",
		initFunc: func(ctx context.Context) {
			callCount++
		},
		execFunc: func(ctx context.Context) {
			callCount++
		},
	}
	plan := New(Conf{})
	plan.Flow(From(p1).To(p2))

	plan.Exec()

	if p2.GetOutput() == nil {
		t.Fatal("Flow() not connecting components")
	}

	// make sure all data made it
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
		if callCount != 4 {
			t.Fatal("Plan.Exec may not be walking to all nodes")
		}
		if count != 2 {
			t.Fatal("Flow not connecting properly. Expected 2 items, got", count)
		}
	case <-time.After(5 * time.Millisecond):
		t.Fatal("Waited too long, something is broken")
	}
}

func TestDefaultPlan_WithAuxEndpoint(t *testing.T) {
	in := make(chan interface{})
	go func() {
		in <- 1
		in <- 2
		close(in)
	}()

	p1 := &testProc{
		name: "P1",
		execFunc: func(ctx context.Context) {
			err := autoctx.SendAuxMsg(ctx, 1)
			if err != nil {
				t.Fatal(err)
			}
		},
		input: in,
	}

	p1.SetInput(in)
	p2 := &testProc{
		name: "P2",
		execFunc: func(ctx context.Context) {
			err := autoctx.SendAuxMsg(ctx, 2)
			if err != nil {
				t.Fatal(err)
			}
		},
	}
	plan := New(Conf{})
	plan.Flow(From(p1).To(p2))

	auxCount := 0
	plan.WithAuxEndpoint(
		&proc.Endpoint{
			Name: "Auxproc",
			Function: func(ctx context.Context, item interface{}) error {
				auxCount++
				return nil
			},
		},
	)

	go func() {
		for _ = range p2.GetOutput() {
		}
	}()

	select {
	case <-plan.Exec():
	case <-time.After(5 * time.Millisecond):
		t.Fatal("Waited too long to execute")
	}
	if auxCount != 2 {
		t.Fatal("Auxiliary not processed properly")
	}

}

func TestDefaultPlan_WithAuxPlan(t *testing.T) {
	in := make(chan interface{})
	go func() {
		in <- 1
		in <- 2
		close(in)
	}()

	p1 := &testProc{
		name: "P1",
		execFunc: func(ctx context.Context) {
			err := autoctx.SendAuxMsg(ctx, 1)
			if err != nil {
				t.Fatal(err)
			}
		},
		input: in,
	}

	p1.SetInput(in)
	p2 := &testProc{
		name: "P2",
		execFunc: func(ctx context.Context) {
			err := autoctx.SendAuxMsg(ctx, 2)
			if err != nil {
				t.Fatal(err)
			}
		},
	}
	plan := New(Conf{})
	plan.Flow(From(p1).To(p2))

	auxPlan := New(Conf{})
	auxOp := &sup.NoopProc{Name: "auxOp"}
	auxOp.SetInput(plan.AuxChan())

	auxCount := 0
	endpoint := &proc.Endpoint{
		Name: "endpoint",
		Function: func(ctx context.Context, item interface{}) error {
			auxCount++
			return nil
		},
	}
	auxPlan.Flow(From(auxOp).To(endpoint))
	plan.WithAuxPlan(auxPlan)

	go func() {
		for _ = range p2.GetOutput() {
		}
	}()

	select {
	case <-plan.Exec():
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Waited too long to execute")
	}
	if auxCount != 2 {
		t.Fatal("Expected auxCount = 2, got ", auxCount)
	}
}
