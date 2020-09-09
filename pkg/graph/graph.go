package graph

import (
	"sync"
)

type Node interface {
	Children() []Node
}

func Walk(node Node, fn func(Node)) {
	for _, n := range node.Children() {
		Walk(n, fn)
	}

	fn(node)
}

func AsyncWalk(node Node, fn func(Node)) {
	wg := sync.WaitGroup{}
	for _, n := range node.Children() {
		n := n

		wg.Add(1)
		go func() {
			AsyncWalk(n, fn)
			wg.Done()
		}()
	}

	wg.Wait()
	fn(node)
}
