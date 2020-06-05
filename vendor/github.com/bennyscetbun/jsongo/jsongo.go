// Copyright 2014 Benny Scetbun. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package Jsongo is a simple library to help you build Json without static struct
//
// Source code and project home:
// https://github.com/benny-deluxe/jsongo
//
//go:generate stringer -type=NodeType

package jsongo

import (
	"encoding/json"
	"errors"
	"reflect"
	//"fmt"
)

//ErrorKeyAlreadyExist error if a key already exist in current Node
var ErrorKeyAlreadyExist = errors.New("jsongo key already exist")

//ErrorMultipleType error if a Node already got a different type of value
var ErrorMultipleType = errors.New("jsongo this node is already set to a different NodeType")

//ErrorArrayNegativeValue error if you ask for a negative index in an array
var ErrorArrayNegativeValue = errors.New("jsongo negative index for array")

//ErrorArrayNegativeValue error if you ask for a negative index in an array
var ErrorAtUnsupportedType = errors.New("jsongo Unsupported Type as At argument")

//ErrorRetrieveUserValue error if you ask the value of a node that is not a value node
var ErrorRetrieveUserValue = errors.New("jsongo Cannot retrieve node's value which is not of type value")

//ErrorTypeUnmarshaling error if you try to unmarshal something in the wrong type
var ErrorTypeUnmarshaling = errors.New("jsongo Wrong type when Unmarshaling")

//ErrorUnknowType error if you try to use an unknow NodeType
var ErrorUnknowType = errors.New("jsongo Unknow NodeType")

//ErrorValNotPointer error if you try to use Val without a valid pointer
var ErrorValNotPointer = errors.New("jsongo: Val: arguments must be a pointer and not nil")

//ErrorGetKeys error if you try to get the keys from a Node that isnt a TypeMap or a TypeArray
var ErrorGetKeys = errors.New("jsongo: GetKeys: Node is not a TypeMap or TypeArray")

//ErrorDeleteKey error if you try to call DelKey on a Node that isnt a TypeMap
var ErrorDeleteKey = errors.New("jsongo: DelKey: This Node is not a TypeMap")

//ErrorCopyType error if you try to call Copy on a Node that isnt a TypeUndefined
var ErrorCopyType = errors.New("jsongo: Copy: This Node is not a TypeUndefined")

//Node Datastructure to build and maintain Nodes
type Node struct {
	m          map[string]*Node
	a          []Node
	v          interface{}
	vChanged   bool     //True if we changed the type of the value
	t          NodeType //Type of that Node 0: Not defined, 1: map, 2: array, 3: value
	dontExpand bool     //dont expand while Unmarshal
}

//NodeType is used to set, check and get the inner type of a Node
type NodeType uint

const (
	//TypeUndefined is set by default for empty Node
	TypeUndefined NodeType = iota
	//TypeMap is set when a Node is a Map
	TypeMap
	//TypeArray is set when a Node is an Array
	TypeArray
	//TypeValue is set when a Node is a Value Node
	TypeValue
	//typeError help us detect errors
	typeError
)

//At helps you move through your node by building them on the fly
//
//val can be string or int only
//
//strings are keys for TypeMap
//
//ints are index in TypeArray (it will make array grow on the fly, so you should start to populate with the biggest index first)*
func (that *Node) At(val ...interface{}) *Node {
	if len(val) == 0 {
		return that
	}
	switch vv := val[0].(type) {
	case string:
		return that.atMap(vv, val[1:]...)
	case int:
		return that.atArray(vv, val[1:]...)
	}
	panic(ErrorAtUnsupportedType)
}

//atMap return the Node in current map
func (that *Node) atMap(key string, val ...interface{}) *Node {
	if that.t != TypeUndefined && that.t != TypeMap {
		panic(ErrorMultipleType)
	}
	if that.m == nil {
		that.m = make(map[string]*Node)
		that.t = TypeMap
	}
	if next, ok := that.m[key]; ok {
		return next.At(val...)
	}
	that.m[key] = new(Node)
	return that.m[key].At(val...)
}

//atArray return the Node in current TypeArray (and make it grow if necessary)
func (that *Node) atArray(key int, val ...interface{}) *Node {
	if that.t == TypeUndefined {
		that.t = TypeArray
	} else if that.t != TypeArray {
		panic(ErrorMultipleType)
	}
	if key < 0 {
		panic(ErrorArrayNegativeValue)
	}
	if key >= len(that.a) {
		newa := make([]Node, key+1)
		for i := 0; i < len(that.a); i++ {
			newa[i] = that.a[i]
		}
		that.a = newa
	}
	return that.a[key].At(val...)
}

//Map Turn this Node to a TypeMap and/or Create a new element for key if necessary and return it
func (that *Node) Map(key string) *Node {
	if that.t != TypeUndefined && that.t != TypeMap {
		panic(ErrorMultipleType)
	}
	if that.m == nil {
		that.m = make(map[string]*Node)
		that.t = TypeMap
	}
	if _, ok := that.m[key]; ok {
		return that.m[key]
	}
	that.m[key] = &Node{}
	return that.m[key]
}

//Array Turn this Node to a TypeArray and/or set the array size (reducing size will make you loose data)
func (that *Node) Array(size int) *[]Node {
	if that.t == TypeUndefined {
		that.t = TypeArray
	} else if that.t != TypeArray {
		panic(ErrorMultipleType)
	}
	if size < 0 {
		panic(ErrorArrayNegativeValue)
	}
	var min int
	if size < len(that.a) {
		min = size
	} else {
		min = len(that.a)
	}
	newa := make([]Node, size)
	for i := 0; i < min; i++ {
		newa[i] = that.a[i]
	}
	that.a = newa
	return &(that.a)
}

//Val Turn this Node to Value type and/or set that value to val
func (that *Node) Val(val interface{}) {
	if that.t == TypeUndefined {
		that.t = TypeValue
	} else if that.t != TypeValue {
		panic(ErrorMultipleType)
	}
	rt := reflect.TypeOf(val)
	var finalval interface{}
	if val == nil {
		finalval = &val
		that.vChanged = true
	} else if rt.Kind() != reflect.Ptr {
		rv := reflect.ValueOf(val)
		var tmp reflect.Value
		if rv.CanAddr() {
			tmp = rv.Addr()
		} else {
			tmp = reflect.New(rt)
			tmp.Elem().Set(rv)
		}
		finalval = tmp.Interface()
		that.vChanged = true
	} else {
		finalval = val
	}
	that.v = finalval
}

//Get Return value of a TypeValue as interface{}
func (that *Node) Get() interface{} {
	if that.t != TypeValue {
		panic(ErrorRetrieveUserValue)
	}
	if that.vChanged {
		rv := reflect.ValueOf(that.v)
		return rv.Elem().Interface()
	}
	return that.v
}

// MustGetBool Return value of a TypeValue as bool
// will panic if cant convert the internal value or if the node is not a TypeValue
func (that *Node) MustGetBool() bool {
	return that.Get().(bool)
}

// MustGetString Return value of a TypeValue as string
// will panic if cant convert the internal value or if the node is not a TypeValue
func (that *Node) MustGetString() string {
	return that.Get().(string)
}

// MustGetInt Return value of a TypeValue as int
// will panic if cant convert the internal value or if the node is not a TypeValue
func (that *Node) MustGetInt() int {
	return (int)(that.Get().(float64))
}

// MustGetInt8 Return value of a TypeValue as int8
// will panic if cant convert the internal value or if the node is not a TypeValue
func (that *Node) MustGetInt8() int8 {
	return (int8)(that.Get().(float64))
}

// MustGetInt16 Return value of a TypeValue as int16
// will panic if cant convert the internal value or if the node is not a TypeValue
func (that *Node) MustGetInt16() int16 {
	return (int16)(that.Get().(float64))
}

// MustGetInt32 Return value of a TypeValue as int32
// will panic if cant convert the internal value or if the node is not a TypeValue
func (that *Node) MustGetInt32() int32 {
	return (int32)(that.Get().(float64))
}

// MustGetInt64 Return value of a TypeValue as int64
// will panic if cant convert the internal value or if the node is not a TypeValue
func (that *Node) MustGetInt64() int64 {
	return (int64)(that.Get().(float64))
}

// MustGetUint Return value of a TypeValue as uint
// will panic if cant convert the internal value or if the node is not a TypeValue
func (that *Node) MustGetUint() uint {
	return (uint)(that.Get().(float64))
}

// MustGetUint8 Return value of a TypeValue as uint8
// will panic if cant convert the internal value or if the node is not a TypeValue
func (that *Node) MustGetUint8() uint8 {
	return (uint8)(that.Get().(float64))
}

// MustGetUint16 Return value of a TypeValue as uint16
// will panic if cant convert the internal value or if the node is not a TypeValue
func (that *Node) MustGetUint16() uint16 {
	return (uint16)(that.Get().(float64))
}

// MustGetUint32 Return value of a TypeValue as uint32
// will panic if cant convert the internal value or if the node is not a TypeValue
func (that *Node) MustGetUint32() uint32 {
	return (uint32)(that.Get().(float64))
}

// MustGetUint64 Return value of a TypeValue as uint64
// will panic if cant convert the internal value or if the node is not a TypeValue
func (that *Node) MustGetUint64() uint64 {
	return (uint64)(that.Get().(float64))
}

// MustGetFloat32 Return value of a TypeValue as float32
// will panic if cant convert the internal value or if the node is not a TypeValue
func (that *Node) MustGetFloat32() float32 {
	return (float32)(that.Get().(float64))
}

// MustGetFloat64 Return value of a TypeValue as float64
// will panic if cant convert the internal value or if the node is not a TypeValue
func (that *Node) MustGetFloat64() float64 {
	return that.Get().(float64)
}

//GetKeys Return a slice interface that represent the keys to use with the At fonction (Works only on TypeMap and TypeArray)
func (that *Node) GetKeys() []interface{} {
	var ret []interface{}
	switch that.t {
	case TypeMap:
		nb := len(that.m)
		ret = make([]interface{}, nb)
		for key := range that.m {
			nb--
			ret[nb] = key
		}
	case TypeArray:
		nb := len(that.a)
		ret = make([]interface{}, nb)
		for nb > 0 {
			nb--
			ret[nb] = nb
		}
	default:
		panic(ErrorGetKeys)
	}
	return ret
}

//Len Return the length of the current Node
//
// if TypeUndefined return 0
//
// if TypeValue return 1
//
// if TypeArray return the size of the array
//
// if TypeMap return the size of the map
func (that *Node) Len() int {
	var ret int
	switch that.t {
	case TypeMap:
		ret = len(that.m)
	case TypeArray:
		ret = len(that.a)
	case TypeValue:
		ret = 1
	}
	return ret
}

//SetType Is use to set the Type of a node and return the current Node you are working on
func (that *Node) SetType(t NodeType) *Node {
	if that.t != TypeUndefined && that.t != t {
		panic(ErrorMultipleType)
	}
	if t >= typeError {
		panic(ErrorUnknowType)
	}
	that.t = t
	switch t {
	case TypeMap:
		that.m = make(map[string]*Node, 0)
	case TypeArray:
		that.a = make([]Node, 0)
	case TypeValue:
		that.Val(nil)
	}
	return that
}

//GetType Is use to Get the Type of a node
func (that *Node) GetType() NodeType {
	return that.t
}

//Copy Will set this node like the one in argument. this node must be of type TypeUndefined
//
//if deepCopy is true we will copy all the children recursively else we will share the children
//
//return the current Node
func (that *Node) Copy(other *Node, deepCopy bool) *Node {
	if that.t != TypeUndefined {
		panic(ErrorCopyType)
	}

	if other.t == TypeValue {
		*that = *other
	} else if other.t == TypeArray {
		if !deepCopy {
			*that = *other
		} else {
			that.Array(len(other.a))
			for i := range other.a {
				that.At(i).Copy(other.At(i), deepCopy)
			}
		}
	} else if other.t == TypeMap {
		that.SetType(other.t)
		if !deepCopy {
			for val := range other.m {
				that.m[val] = other.m[val]
			}
		} else {
			for val := range other.m {
				that.Map(val).Copy(other.At(val), deepCopy)
			}
		}
	}
	return that
}

//Unset Will unset everything in the Node. All the children data will be lost
func (that *Node) Unset() {
	*that = Node{}
}

//DelKey will remove a key in the map.
//
//return the current Node.
func (that *Node) DelKey(key string) *Node {
	if that.t != TypeMap {
		panic(ErrorDeleteKey)
	}
	delete(that.m, key)
	return that
}

//UnmarshalDontExpand set or not if Unmarshall will generate anything in that Node and its children
//
//val: will change the expanding rules for this node
//
//- The type wont be change for any type
//
//- Array wont grow
//
//- New keys wont be added to Map
//
//- Values set to nil "*.Val(nil)*" will be turn into the type decide by Json
//
//- It will respect any current mapping and will return errors if needed
//
//recurse: if true, it will set all the children of that Node with val
func (that *Node) UnmarshalDontExpand(val bool, recurse bool) *Node {
	that.dontExpand = val
	if recurse {
		switch that.t {
		case TypeMap:
			for k := range that.m {
				that.m[k].UnmarshalDontExpand(val, recurse)
			}
		case TypeArray:
			for k := range that.a {
				that.a[k].UnmarshalDontExpand(val, recurse)
			}
		}
	}
	return that
}

//MarshalJSON Make Node a Marshaler Interface compatible
func (that *Node) MarshalJSON() ([]byte, error) {
	var ret []byte
	var err error
	switch that.t {
	case TypeMap:
		ret, err = json.Marshal(that.m)
	case TypeArray:
		ret, err = json.Marshal(that.a)
	case TypeValue:
		ret, err = json.Marshal(that.v)
	default:
		ret, err = json.Marshal(nil)
	}
	if err != nil {
		return nil, err
	}
	return ret, err
}

func (that *Node) unmarshalMap(data []byte) error {
	tmp := make(map[string]json.RawMessage)
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	for k := range tmp {
		if _, ok := that.m[k]; ok {
			err := json.Unmarshal(tmp[k], that.m[k])
			if err != nil {
				return err
			}
		} else if !that.dontExpand {
			err := json.Unmarshal(tmp[k], that.Map(k))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (that *Node) unmarshalArray(data []byte) error {
	var tmp []json.RawMessage
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	for i := len(tmp) - 1; i >= 0; i-- {
		if !that.dontExpand || i < len(that.a) {
			err := json.Unmarshal(tmp[i], that.At(i))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (that *Node) unmarshalValue(data []byte) error {
	if that.v != nil {
		return json.Unmarshal(data, that.v)
	}
	var tmp interface{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	that.Val(tmp)
	return nil
}

//UnmarshalJSON Make Node a Unmarshaler Interface compatible
func (that *Node) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if that.dontExpand && that.t == TypeUndefined {
		return nil
	}
	if that.t == TypeValue {
		return that.unmarshalValue(data)
	}
	if data[0] == '{' {
		if that.t != TypeMap && that.t != TypeUndefined {
			return ErrorTypeUnmarshaling
		}
		return that.unmarshalMap(data)
	}
	if data[0] == '[' {
		if that.t != TypeArray && that.t != TypeUndefined {
			return ErrorTypeUnmarshaling
		}
		return that.unmarshalArray(data)

	}
	if that.t == TypeUndefined {
		return that.unmarshalValue(data)
	}
	return ErrorTypeUnmarshaling
}
