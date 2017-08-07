package GOPACKAGE

type resultTypeQueue struct {
	items []Data
}

func resultNew() *resultTypeQueue {
	return &resultTypeQueue{items: make([]Data, 0)}
}

func (q *resultTypeQueue) Enq(obj Data) *resultTypeQueue {
	q.items = append(q.items, obj)
	return q
}

func (q *resultTypeQueue) Deq() Data {
	obj := q.items[0]
	q.items = q.items[1:]
	return obj
}

func (q *resultTypeQueue) Len() int {
	return len(q.items)
}
