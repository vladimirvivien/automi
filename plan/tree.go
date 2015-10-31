package plan

import(
	"fmt"

    "github.com/vladimirvivien/automi/api"
)

type node struct {
	proc  interface{}
	nodes []*node
}

func (n *node) procName() string{
	proc, ok := n.proc.(api.Process)
	if !ok {
		return ""
	}
	return proc.GetName()
}

func (n *node) String() string {
	return fmt.Sprintf("%v -> %v", n.proc, n.nodes)
}

func (n *node) eq(other *node) bool {
	// pull out process for node
	proc1, isProcType1 := n.proc.(api.Process)

	// pull other process
	proc2, isProcType2 := other.proc.(api.Process)

	// if proc type and name are equal
	if isProcType1 && isProcType2 && proc1.GetName() == proc2.GetName() {
		return true
	}
	return false
}

// search bails after first match
func search(root *node, key string) *node {
	if root == nil || key == "" {
		return nil
	}

	// pull n-process
	nproc, ok := root.proc.(api.Process)
	if !ok {
		return nil
	}

	if root.procName() == key{
		return root
	}	

	
	// match is found
	if nproc.GetName()  == key {
		return root
	} else {
		// no match, continue
		if root.nodes != nil {
			for _, n := range root.nodes {
				result := search(n, key)
				if result != nil && result.procName() == key {
					return result
				}
			}
		}
	}
	return nil
}


// finds existing node and overwrites it
func graph(t *node, n *node) *node {
	if n == nil {
		return t
	}

	if t == nil {
		return n
	}

	// overwrite
	found := search(t, n.procName())
	if found != nil {
		*found = *n
		return t
	}

	// insert as a branch
	if t != nil && found == nil {
		t.nodes = append(t.nodes, n)
	}
	return t
}

// walks the tree and exec f() for each node
func walk(t *node, f func(*node)) {
	if t == nil {
		return
	}

	if f != nil {
		f(t)
	}
	if t.nodes != nil {
		for _, n := range t.nodes {
			walk(n, f)
		}
	}

}