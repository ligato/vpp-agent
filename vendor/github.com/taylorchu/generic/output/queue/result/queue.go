package result

type FIFO struct {
	items []int64
}

func New() *FIFO {
	return &FIFO{items: make([]int64, 0)}
}

func (q *FIFO) Enq(obj int64) *FIFO {
	q.items = append(q.items, obj)
	return q
}

func (q *FIFO) Deq() int64 {
	obj := q.items[0]
	q.items = q.items[1:]
	return obj
}

func (q *FIFO) Len() int {
	return len(q.items)
}
