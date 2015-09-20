package sup

import "testing"

func TestNoopProc_ProcGraph(t *testing.T) {
	p1 := &NoopProc{Name: "P1"}
	p1.SetInput(make(chan interface{}))
	p2 := &NoopProc{Name: "P2"}
	p2.SetInput(p1.GetOutput())
	p3 := &NoopProc{Name: "P3"}
	p3.SetInput(p2.GetOutput())
	if p1.GetOutput() != p2.input && p2.GetOutput() != p3.input {
		t.Fatal("Graph Path failed: p1 -> p2 ->p3")
	}
}
