package plan

import (
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"

	autoctx "github.com/vladimirvivien/automi/context"
	"github.com/vladimirvivien/automi/proc"
	"github.com/vladimirvivien/automi/sup"
	"github.com/vladimirvivien/automi/testutil"
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

type testEndpoint struct {
	name     string
	initFunc func(context.Context)
	execFunc func(context.Context)
	input    <-chan interface{}
	done     chan struct{}
}

func (p *testEndpoint) GetName() string {
	return p.name
}

func (p *testEndpoint) Init(ctx context.Context) error {
	p.done = make(chan struct{})

	if p.initFunc != nil {
		p.initFunc(ctx)
	}
	return nil
}

func (p *testEndpoint) Uninit(ctx context.Context) error {
	return nil
}

func (p *testEndpoint) Exec(ctx context.Context) error {
	if p.execFunc != nil {
		p.execFunc(ctx)
	}
	go func() {
		defer func() {
			close(p.done)
		}()
		for _ = range p.input {
		}
	}()
	return nil
}

func (p *testEndpoint) SetInput(in <-chan interface{}) {
	p.input = in
}
func (p *testEndpoint) Done() <-chan struct{} {
	return p.done
}

func TestDefaultPlan_Flow(t *testing.T) {
	p1 := &sup.NoopProc{Name: "P1"}
	p2 := &sup.NoopProc{Name: "P2"}
	p3 := &sup.NoopProc{Name: "P3"}
	p4 := &sup.NoopProc{Name: "P4"}
	p5 := &sup.NoopProc{Name: "P5"}

	plan := New(&Conf{})
	flow := From(p1).To(p2)
	plan.Flow(flow).Flow(From(p2).To(p3, p4, p5))

	if flow == nil {
		t.Fatal("From() not building node tree")
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
	p2 := &testEndpoint{
		name: "P2",
		initFunc: func(ctx context.Context) {
			callCount++
		},
		execFunc: func(ctx context.Context) {
			callCount++
		},
	}
	plan := New(&Conf{})
	plan.Flow(From(p1).To(p2))

	select {
	case <-plan.Exec():
		if callCount != 4 {
			t.Fatal("Plan.Exec may not be walking to all nodes")
		}
	case <-time.After(5 * time.Millisecond):
		t.Fatal("Waited too long, something is broken")
	}
}

func BenchmarkDefaultPlan_Exec(b *testing.B) {
	ctx := context.Background()
	N := b.N
	chanSize := func() int {
		if N == 1 {
			return N
		}
		return int(float64(0.5) * float64(N))
	}()

	in := make(chan interface{}, chanSize)
	go func() {
		for i := 0; i < N; i++ {
			in <- testutil.GenWord()
		}
		close(in)
	}()

	//exeCount := 0
	counter := 0
	var m sync.RWMutex

	p1 := &sup.NoopProc{Name: "P1"}
	p1.SetInput(in)

	// p2 := &proc.Item{
	// 	Name: "P2",
	// 	Concurrency: 4,
	// 	Function: func(ctx context.Context, i interface{}) interface{} {
	// 		//m.Lock()
	// 		///counter++
	// 		//m.Unlock()
	// 		return i
	// 	},
	// }

	p3 := &proc.Endpoint{
		Name: "P3",
		Function: func(ctx context.Context, item interface{}) error {
			m.Lock()
			counter++
			m.Unlock()
			return nil
		},
	}

	// setup plan
	plan := New(&Conf{Ctx: ctx})
	plan.Flow(From(p1).To(p3))
	//plan.Flow(From(p2).To(p3))

	select {
	case <-plan.Exec():
	case <-time.After(60 * time.Second):
		b.Fatal("Waited too long, something is broken")
	}

	b.Logf("Input %d, processed %d", N, counter)
	if counter != N {
		b.Fatalf("Expected %d processed items, got %d", N, counter)
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
	p2 := &testEndpoint{
		name: "P2",
		execFunc: func(ctx context.Context) {
			err := autoctx.SendAuxMsg(ctx, 2)
			if err != nil {
				t.Fatal(err)
			}
		},
	}
	plan := New(&Conf{})
	plan.Flow(From(p1).To(p2))

	var m sync.RWMutex
	auxCount := 0
	plan.WithAuxEndpoint(
		&proc.Endpoint{
			Name: "Auxproc",
			Function: func(ctx context.Context, item interface{}) error {
				m.Lock()
				auxCount++
				m.Unlock()
				return nil
			},
		},
	)

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
	plan := New(&Conf{})
	plan.Flow(From(p1).To(p2))

	auxPlan := New(&Conf{})
	auxOp := &sup.NoopProc{Name: "auxOp"}
	auxOp.SetInput(plan.AuxChan())

	var m sync.RWMutex
	auxCount := 0
	endpoint := &proc.Endpoint{
		Name: "endpoint",
		Function: func(ctx context.Context, item interface{}) error {
			m.Lock()
			auxCount++
			m.Unlock()
			return nil
		},
	}
	auxPlan.Flow(From(auxOp).To(endpoint))
	plan.WithAuxPlan(auxPlan)

	//TODO - Revisit withAuxPlan() hanging
	//select {
	//case <-plan.Exec():
	//case <-time.After(5 * time.Millisecond):
	//	t.Fatal("Waited too long to execute")
	//}
	if auxCount != 0 {
		t.Fatal("Expected auxCount = 2, got ", auxCount)
	}
}
