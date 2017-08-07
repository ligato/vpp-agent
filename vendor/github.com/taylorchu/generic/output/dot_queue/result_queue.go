package GOPACKAGE

type FIFO struct {
	items []Data
}

func resultNew() *FIFO {
	return &FIFO{items: make([]Data, 0)}
}

func (q *FIFO) Enq(obj Data) *FIFO {
	q.items = append(q.items, obj)
	return q
}

func (q *FIFO) Deq() Data {
	obj := q.items[0]
	q.items = q.items[1:]
	return obj
}

func (q *FIFO) Len() int {
	return len(q.items)
}
