package GOPACKAGE

import "fmt"

var (
	resultA = map[int]string{
		1: "hello",
	}
)

const (
	resultX = 123
	_       = 1
)

type resultStruct struct {
	Val Data
}

func (s resultStruct) hello() {
	resultAdd()
	fmt.Println(resultX, resultA)
}
