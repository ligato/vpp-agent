package queue

type Type string

// TypeQueue represents a queue of Type types.
type TypeQueue struct {
	items []Type
}

// New makes a new empty Type queue.
func New() *TypeQueue {
	return &TypeQueue{items: make([]Type, 0)}
}

// Enq adds an item to the queue.
func (q *TypeQueue) Enq(obj Type) *TypeQueue {
	q.items = append(q.items, obj)
	return q
}

// Deq removes and returns the next item in the queue.
func (q *TypeQueue) Deq() Type {
	obj := q.items[0]
	q.items = q.items[1:]
	return obj
}

// Len gets the current number of Type items in the queue.
func (q *TypeQueue) Len() int {
	return len(q.items)
}
