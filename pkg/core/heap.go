package core

type FloatHeapItem struct {
	Key   interface{}
	Value float64
}

// https://golang.org/pkg/container/heap
type FloatHeap []FloatHeapItem

func (h FloatHeap) Len() int           { return len(h) }
func (h FloatHeap) Less(i, j int) bool { return h[i].Value > h[j].Value }
func (h FloatHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *FloatHeap) Push(x interface{}) {
	*h = append(*h, x.(FloatHeapItem))
}

func (h *FloatHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
