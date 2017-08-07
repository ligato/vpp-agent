package GOPACKAGE

import (
	"container/list"
	"net"
)

type Data struct {
	field  unresolvedType
	field2 Element
	field3 Listener
}

type Element list.Element
type Listener net.Listener

var (
	_ = unresolvedFunc()
)
