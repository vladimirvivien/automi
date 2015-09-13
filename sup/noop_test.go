package sup

import "testing"

func TestNoopProc_ProcGraph(t *testing.T) {
	p1 := &NoopProc{Name: "P1", Output: make(chan interface{})}
	p2 := &NoopProc{Name: "P2", Input: p1.GetOutput()}
	p3 := &NoopProc{Name: "P3", Input: p2.GetOutput()}
	if p1.Output != p2.Input && p2.Output != p3.Input {
		t.Fatal("Graph Path failed: p1 -> p2 ->")
	}
}
