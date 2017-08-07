package result

import "github.com/taylorchu/generic/fixture/vendoring"

type Struct struct {
	Val vendoring.Number
}

func add(a, b vendoring.Number) {
	_ = func(c vendoring.Number) {
	}
}
