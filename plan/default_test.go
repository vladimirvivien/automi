package plan

import (
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/vladimirvivien/automi/sup"
)

type testProc struct {
	name     string
	initFunc func(context.Context)
	execFunc func(context.Context)
	input    <-chan interface{}
	output   <-chan interface{}
}

func (p *testProc) GetName() string {
	return p.name
}

func (p *testProc) Init(ctx context.Context) error {
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
	return nil
}

func (p *testProc) GetOutput() <-chan interface{} {
	return p.output
}
func (p *testProc) SetInput(in <-chan interface{}) {
	p.output = in
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

	walk(tree, func(n *node) { t.Log(n) })
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
	plan := New()
	plan.Flow(From(p1).To(p2))

	if p2.GetOutput() == nil {
		t.Fatal("Flow() not connecting components")
	}

	plan.Exec()

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
		t.Log("callCount", callCount)
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
