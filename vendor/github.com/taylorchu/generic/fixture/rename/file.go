package rename

import "fmt"

var (
	A = map[int]string{
		1: "hello",
	}
)

const (
	X = 123
	_ = 1
)

type Struct struct {
	Val Type
}

func (s Struct) hello() {
	add()
	fmt.Println(X, A)
}
