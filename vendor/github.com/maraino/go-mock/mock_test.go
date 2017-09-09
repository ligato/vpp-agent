package mock

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type AStruct struct {
	Value int
}

type MockedStruct struct {
	Mock
}

type MockedInterface interface {
	FuncNoArgs() int
	FuncWithArgs(a int, b string) (int, string)
	FuncWithPointerArgs(a int, b *string) (int, *string)
	FuncWithRetArg(a int, b interface{}) int
	FuncVariadic(a int, args ...interface{}) (int, string)
}

func (m *MockedStruct) FuncNoArgs() int {
	ret := m.Called()
	return ret.Int(0)
}
func (m *MockedStruct) FuncWithArgs(a int, b string) (int, string) {
	ret := m.Called(a, b)
	return ret.Int(0), ret.String(1)
}

func (m *MockedStruct) FuncWithPointerArgs(a int, b *string) (r1 int, r2 *string) {
	ret := m.Called(a, b)
	r1 = ret.Int(0)
	r2 = ret.GetType(1, r2).(*string)
	return
}

func (m *MockedStruct) FuncWithRetArg(a int, b interface{}) int {
	ret := m.Called(a, b)
	return ret.Int(0)
}

func (m *MockedStruct) FuncVariadic(a int, args ...interface{}) (int, string) {
	ret := m.Called(a, args)
	return ret.Int(0), ret.String(1)
}

func (m *MockedStruct) FuncMockedInterface(a interface{}) (i MockedInterface) {
	ret := m.Called(a)
	i = ret.GetType(0, &MockedStruct{}).(MockedInterface)
	return
}

func (m *MockedStruct) FuncMockResult(a interface{}) *MockResult {
	return m.Called(a)
}

func (m *MockedStruct) Verify() bool {
	ret := m.Called()
	return ret.Bool(0)
}

func TestSanity(t *testing.T) {
	m := MockedStruct{}
	foo := 3
	m.When("FuncWithRetArg", 1, Any).Return(2).ReturnToArgument(1, &foo).Times(1)
	m.When("FuncWithRetArg", 2, AnyOfType("**int")).Return(3).ReturnToArgument(1, &foo).AtMost(2)
	m.When("FuncWithRetArg", 2, AnyOfType("*int")).Return(4).ReturnToArgument(1, 5).AtMost(2)
	m.When("Verify").Return(false).AtLeast(1).Timeout(100 * time.Millisecond)

	var b *int
	ret := m.FuncWithRetArg(1, &b)
	if ret != 2 {
		t.Errorf("Invalid return value. Expected: %d. Found: %d.", 2, ret)
	}
	if *b != 3 {
		t.Errorf("Invalid value for b. Expected: %d. Found: %d.", 3, *b)
	}

	ret = m.FuncWithRetArg(2, &b)
	if ret != 3 {
		t.Errorf("Invalid return value. Expected: %d. Found: %d.", 3, ret)
	}
	if *b != 3 {
		t.Errorf("Invalid value for b. Expected: %d. Found: %d.", 3, *b)
	}

	var c int
	ret = m.FuncWithRetArg(2, &c)
	if ret != 4 {
		t.Errorf("Invalid return value. Expected: %d. Found: %d.", 4, ret)
	}
	if c != 5 {
		t.Errorf("Invalid value for c. Expected: %d. Found: %d.", 5, *b)
	}

	ret = m.FuncWithRetArg(2, &c)
	if ret != 4 {
		t.Errorf("Invalid return value. Expected: %d. Found: %d.", 4, ret)
	}
	if c != 5 {
		t.Errorf("Invalid value for c. Expected: %d. Found: %d.", 5, *b)
	}

	ret2 := m.Verify()
	if ret2 != false {
		t.Errorf("Invalid return value: Expected: false. Found: %v.", ret2)
	}

	if ok, err := m.Mock.Verify(); !ok {
		t.Error(err)
	}
}

func TestReset(t *testing.T) {
	m := MockedStruct{}
	m.When("FuncNoArgs").Return(1)

	i := m.FuncNoArgs()
	if i != 1 {
		t.Error("fail")
	}

	m.Reset()

	if len(m.Functions) != 0 {
		t.Error("fail")
	}

	if m.order != 0 {
		t.Error("fail")
	}

	m.When("FuncNoArgs").Return(2)

	i = m.FuncNoArgs()
	if i != 2 {
		t.Error("fail")
	}
}

func TestPanic(t *testing.T) {
	m := MockedStruct{}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Panic not executed")
		} else if ok, err := m.Mock.Verify(); !ok {
			t.Error(err)
		}
	}()

	m.When("FuncWithArgs", 1, "foo").Panic("panic").Times(1)
	m.FuncWithArgs(1, "foo")
}

func TestMockMissingPanic(t *testing.T) {
	m := MockedStruct{}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Panic not executed")
		}
	}()

	m.FuncNoArgs()
}

func TestReturn(t *testing.T) {
	m := &MockedStruct{}
	m.When("FuncNoArgs").Return(1).Times(1)
	m.When("FuncNoArgs").Return(2).Times(1)
	m.When("FuncWithArgs", 1, "string").Return(2, "stringstring").Times(1)
	m.When("FuncWithArgs", 2, "string").Return(4, "stringstring").AtLeast(2)
	m.When("FuncMockedInterface", 3).Return(m).Times(1)

	i := m.FuncNoArgs()
	if i != 1 {
		t.Error("fail")
	}

	i = m.FuncNoArgs()
	if i != 2 {
		t.Error("fail")
	}

	a, b := m.FuncWithArgs(1, "string")
	if a != 2 || b != "stringstring" {
		t.Error("fail")
	}

	a, b = m.FuncWithArgs(2, "string")
	if a != 4 || b != "stringstring" {
		t.Error("fail")
	}

	a, b = m.FuncWithArgs(2, "string")
	if a != 4 || b != "stringstring" {
		t.Error("fail")
	}

	mi := m.FuncMockedInterface(3)
	if mm, ok := mi.(*MockedStruct); !ok || mm != m {
		t.Error("fail")
	}

	if ok, err := m.Mock.Verify(); !ok {
		t.Error(err)
	}
}

func TestAny(t *testing.T) {
	f1 := func(i interface{}) bool {
		ii, ok := i.(int)
		return ok && (ii == 5 || i == 6)
	}

	f2 := func(i interface{}) bool {
		ii, ok := i.(string)
		return ok && ii == "foo"
	}

	m := MockedStruct{}
	m.When("FuncWithArgs", 1, Any).Return(2, "booh").Times(1)
	m.When("FuncWithArgs", 2, Any).Return(4, "booh").Times(2)
	m.When("FuncWithArgs", AnyOfType("int"), AnyOfType("string")).Return(6, "booh").Times(2)
	m.When("FuncWithArgs", AnyIf(f1), AnyIf(f2)).Return(8, "booh").Times(2)

	a, b := m.FuncWithArgs(1, "string")
	if a != 2 || b != "booh" {
		t.Error("fail")
	}

	a, b = m.FuncWithArgs(2, "foo")
	if a != 4 || b != "booh" {
		t.Error("fail")
	}

	a, b = m.FuncWithArgs(2, "bar")
	if a != 4 || b != "booh" {
		t.Error("fail")
	}

	a, b = m.FuncWithArgs(3, "foo")
	if a != 6 || b != "booh" {
		t.Error("fail")
	}

	a, b = m.FuncWithArgs(4, "bar")
	if a != 6 || b != "booh" {
		t.Error("fail")
	}

	a, b = m.FuncWithArgs(5, "foo")
	if a != 8 || b != "booh" {
		t.Error("fail")
	}

	a, b = m.FuncWithArgs(6, "foo")
	if a != 8 || b != "booh" {
		t.Error("fail")
	}

	if ok, err := m.Mock.Verify(); !ok {
		t.Error(err)
	}
}

func TestNil(t *testing.T) {
	m := MockedStruct{}
	var s = "string"
	m.When("FuncWithRetArg", 1, nil).Return(2).Times(1)
	m.When("FuncWithPointerArgs", 2, nil).Return(3, nil).Times(1)
	m.When("FuncWithPointerArgs", 3, &s).Return(4, &s).Times(1)

	var aNil *string
	a := m.FuncWithRetArg(1, aNil)
	if a != 2 {
		t.Error("fail")
	}

	b, c := m.FuncWithPointerArgs(2, aNil)
	if b != 3 && c != nil {
		t.Error("fail")
	}

	b, c = m.FuncWithPointerArgs(3, &s)
	if b != 4 && *c != s {
		t.Error("fail")
	}

	if ok, err := m.Mock.Verify(); !ok {
		t.Error(err)
	}
}

func TestFindMultipleMatches(t *testing.T) {
	m := MockedStruct{}
	m.When("FuncWithArgs", 1, "string").Return(1, "").Times(1)
	m.When("FuncWithArgs", Any, "string").Return(2, "").AtMost(2)
	m.When("FuncWithArgs", 1, Any).Return(3, "").Times(1)
	m.When("FuncWithArgs", Any, Any).Return(4, "").Between(1, 3)
	m.When("FuncWithArgs", AnyOfType("int"), AnyOfType("string")).Return(5, "booh").AtLeast(1)

	results := []int{1, 2, 2, 3, 4, 4, 4, 5, 5, 5, 5}
	for _, r := range results {
		a, _ := m.FuncWithArgs(1, "string")
		if a != r {
			t.Errorf("Invalid return value. Found: %d, Expected: %d.", a, r)
		}
	}

	if ok, err := m.Mock.Verify(); !ok {
		t.Error(err)
	}
}

func TestNotFound(t *testing.T) {
	m := MockedStruct{}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Panic not executed")
		}
	}()

	m.When("FuncWithArgs", 1, "string").Return(2, "stringstring")
	m.FuncWithArgs(1, "foo")
}

func TestAnyIfNotFound(t *testing.T) {
	m := MockedStruct{}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Panic not executed")
		}
	}()

	f1 := func(i interface{}) bool {
		ii, ok := i.(int)
		return ok && ii == 2
	}

	m.When("FuncWithArgs", AnyIf(f1), "foo").Return(2, "stringstring")
	m.FuncWithArgs(1, "foo")
}

func TestReturnToArgument(t *testing.T) {
	m := MockedStruct{}
	ref := "a ref"
	m.When("FuncWithRetArg", 1, Any).Return(2).ReturnToArgument(1, 1)
	m.When("FuncWithRetArg", 11, Any).Return(2).ReturnToArgument(1, 1)
	m.When("FuncWithRetArg", 2, Any).Return(2).ReturnToArgument(1, "a string")
	m.When("FuncWithRetArg", 22, Any).Return(2).ReturnToArgument(1, "a string")
	m.When("FuncWithRetArg", 3, Any).Return(2).ReturnToArgument(1, &ref)
	m.When("FuncWithRetArg", 33, Any).Return(2).ReturnToArgument(1, &ref)
	m.When("FuncWithRetArg", 4, Any).Return(2).ReturnToArgument(1, true)
	m.When("FuncWithRetArg", 44, Any).Return(2).ReturnToArgument(1, true)
	m.When("FuncWithRetArg", 5, Any).Return(2).ReturnToArgument(1, AStruct{5})
	m.When("FuncWithRetArg", 55, Any).Return(2).ReturnToArgument(1, AStruct{55})
	m.When("FuncWithRetArg", 6, Any).Return(2).ReturnToArgument(1, &AStruct{6})
	m.When("FuncWithRetArg", 66, Any).Return(2).ReturnToArgument(1, &AStruct{66})

	var a1 int
	ret := m.FuncWithRetArg(1, &a1)
	if ret != 2 || a1 != 1 {
		t.Error("fail")
	}

	var a11 *int
	ret = m.FuncWithRetArg(11, &a11)
	if ret != 2 || *a11 != 1 {
		t.Error("fail")
	}

	var a2 string
	ret = m.FuncWithRetArg(2, &a2)
	if ret != 2 || a2 != "a string" {
		t.Error("fail")
	}

	var a22 *string
	ret = m.FuncWithRetArg(22, &a22)
	if ret != 2 || *a22 != "a string" {
		t.Error("fail")
	}

	var a3 string
	ret = m.FuncWithRetArg(3, &a3)
	if ret != 2 || a3 != "a ref" {
		t.Error("fail")
	}

	var a33 *string
	ret = m.FuncWithRetArg(33, &a33)
	if ret != 2 || *a33 != "a ref" {
		t.Error("fail")
	}

	var a4 bool
	ret = m.FuncWithRetArg(4, &a4)
	if ret != 2 || a4 != true {
		t.Error("fail")
	}

	var a44 *bool
	ret = m.FuncWithRetArg(44, &a44)
	if ret != 2 || *a44 != true {
		t.Error("fail")
	}

	var a5 AStruct
	ret = m.FuncWithRetArg(5, &a5)
	if ret != 2 || a5.Value != 5 {
		t.Error("fail")
	}

	var a55 *AStruct
	ret = m.FuncWithRetArg(55, &a55)
	if ret != 2 || a55.Value != 55 {
		t.Error("fail")
	}

	var a6 AStruct
	ret = m.FuncWithRetArg(6, &a6)
	if ret != 2 || a6.Value != 6 {
		t.Error("fail")
	}

	var a66 *AStruct
	ret = m.FuncWithRetArg(66, &a66)
	if ret != 2 || a66.Value != 66 {
		t.Error("fail")
	}
}

func TestTimes(t *testing.T) {
	m := MockedStruct{}
	m.When("FuncWithArgs", 1, "times").Return(1, "stringstring").Times(1)
	m.When("FuncWithArgs", 1, "atleast2").Return(2, "stringstring").AtLeast(2)
	m.When("FuncWithArgs", 1, "atleast3").Return(3, "stringstring").AtLeast(2)
	m.When("FuncWithArgs", 1, "atmost0").Return(4, "stringstring").AtMost(2)
	m.When("FuncWithArgs", 1, "atmost1").Return(5, "stringstring").AtMost(2)
	m.When("FuncWithArgs", 1, "atmost2").Return(5, "stringstring").AtMost(2)
	m.When("FuncWithArgs", 1, "between1").Return(5, "stringstring").Between(1, 3)
	m.When("FuncWithArgs", 1, "between2").Return(5, "stringstring").Between(1, 3)
	m.When("FuncWithArgs", 1, "between3").Return(5, "stringstring").Between(1, 3)

	m.FuncWithArgs(1, "times")
	m.FuncWithArgs(1, "atleast2")
	m.FuncWithArgs(1, "atleast2")
	m.FuncWithArgs(1, "atleast3")
	m.FuncWithArgs(1, "atleast3")
	m.FuncWithArgs(1, "atleast3")
	m.FuncWithArgs(1, "atmost1")
	m.FuncWithArgs(1, "atmost2")
	m.FuncWithArgs(1, "atmost2")
	m.FuncWithArgs(1, "between1")
	m.FuncWithArgs(1, "between2")
	m.FuncWithArgs(1, "between2")
	m.FuncWithArgs(1, "between3")
	m.FuncWithArgs(1, "between3")
	m.FuncWithArgs(1, "between3")

	if ok, err := m.Mock.Verify(); !ok {
		t.Error(err)
	}
}

func TestTimesFail(t *testing.T) {
	m := MockedStruct{}
	m.When("FuncWithArgs", 1, "times").Return(1, "stringstring").Times(1)
	if ok, _ := m.Mock.Verify(); ok {
		t.Error("Error expected and not found")
	}

	m = MockedStruct{}
	m.When("FuncWithArgs", 1, "atleast").Return(2, "stringstring").AtLeast(2)
	m.FuncWithArgs(1, "atleast")
	if ok, _ := m.Mock.Verify(); ok {
		t.Error("Error expected and not found")
	}

	m = MockedStruct{}
	m.When("FuncWithArgs", 1, "atmost").Return(5, "stringstring").AtMost(2)
	m.FuncWithArgs(1, "atmost")
	m.FuncWithArgs(1, "atmost")
	m.FuncWithArgs(1, "atmost")
	if ok, _ := m.Mock.Verify(); ok {
		t.Error("Error expected and not found")
	}

	m = MockedStruct{}
	m.When("FuncWithArgs", 1, "between1").Return(5, "stringstring").Between(1, 3)
	if ok, _ := m.Mock.Verify(); ok {
		t.Error("Error expected and not found")
	}

	m = MockedStruct{}
	m.When("FuncWithArgs", 1, "between2").Return(5, "stringstring").Between(1, 3)
	m.FuncWithArgs(1, "between2")
	m.FuncWithArgs(1, "between2")
	m.FuncWithArgs(1, "between2")
	m.FuncWithArgs(1, "between2")
	if ok, _ := m.Mock.Verify(); ok {
		t.Error("Error expected and not found")
	}
}

func TestTimeout(t *testing.T) {
	m := &MockedStruct{}
	m.When("FuncWithArgs", 1, "string").Return(1, "stringstring").Timeout(200 * time.Millisecond)
	m.When("FuncWithArgs", 2, "string").Return(2, "stringstring")

	// Function with timeout
	t1 := time.Now()
	a, b := m.FuncWithArgs(1, "string")
	t2 := time.Now()

	if a != 1 || b != "stringstring" {
		t.Error("fail")
	}

	if t1.Add(200 * time.Millisecond).After(t2) {
		t.Error("fail", t1, t2)
	}

	// Function without timeout
	t1 = time.Now()
	a, b = m.FuncWithArgs(2, "string")
	t2 = time.Now()

	if a != 2 || b != "stringstring" {
		t.Error("fail")
	}

	if t1.Add(200 * time.Millisecond).Before(t2) {
		t.Error("fail")
	}
}

func TestCall(t *testing.T) {
	m := &MockedStruct{}
	m.When("FuncWithArgs", 1, "string").Call(func(a int, b string) (int, string) {
		return a * 2, b + b
	}).Times(1)
	m.When("FuncWithArgs", 2, "string").Call(func(a int) int {
		return a * 2
	}).Times(1)
	m.When("FuncWithArgs", 3, "string").Call(func() {
		return
	}).Times(1)
	m.When("FuncVariadic", 4, []interface{}{"foo", "bar"}).Call(func(a int, args ...interface{}) (int, string) {
		b1, b2 := args[0].(string), args[1].(string)
		return a * 2, b1 + b2
	}).Times(1)
	m.When("FuncWithRetArg", 5, Any).Call(func(a int, b *string) int {
		*b = "foobar"
		return a * 2
	}).Times(1)
	m.When("FuncWithArgs", 6, "string").Call(func(a int, b string, c int) (int, string) {
		return a * c, b + b
	}).Times(1)

	a, b := m.FuncWithArgs(1, "string")
	if a != 2 || b != "stringstring" {
		t.Error("fail")
	}

	a, b = m.FuncWithArgs(2, "string")
	if a != 4 || b != "" {
		t.Error("fail")
	}

	a, b = m.FuncWithArgs(3, "string")
	if a != 0 || b != "" {
		t.Error("fail")
	}

	a, b = m.FuncVariadic(4, "foo", "bar")
	if a != 8 || b != "foobar" {
		t.Error("fail")
	}

	a = m.FuncWithRetArg(5, &b)
	if a != 10 || b != "foobar" {
		t.Error("fail")
	}

	a, b = m.FuncWithArgs(6, "string")
	if a != 0 || b != "stringstring" {
		t.Error("fail")
	}

	if ok, err := m.Mock.Verify(); !ok {
		t.Error(err)
	}
}

func TestMockResult(t *testing.T) {
	var mNil *MockedStruct
	m := &MockedStruct{}
	m.When("FuncMockResult", "struct").Return(m)
	m.When("FuncMockResult", "nil").Return(nil)
	m.When("FuncMockResult", "nil.with.type").Return(mNil)
	m.When("FuncMockResult", "bool").Return(true)
	m.When("FuncMockResult", "byte").Return(byte('a'))
	m.When("FuncMockResult", "bytes").Return([]byte("abc"))
	m.When("FuncMockResult", "bytes.nil").Return(nil)
	m.When("FuncMockResult", "error").Return(errors.New("error"))
	m.When("FuncMockResult", "float32").Return(float32(1.32))
	m.When("FuncMockResult", "float64").Return(float64(1.64))
	m.When("FuncMockResult", "int").Return(int(1))
	m.When("FuncMockResult", "int8").Return(int8(8))
	m.When("FuncMockResult", "int16").Return(int16(16))
	m.When("FuncMockResult", "int32").Return(int32(32))
	m.When("FuncMockResult", "int64").Return(int64(64))
	m.When("FuncMockResult", "string").Return("string")
	m.When("FuncMockResult", "chan").Return(make(chan int))

	ret := m.FuncMockResult("struct")
	if ret.Contains(0) != true {
		t.Error("fail")
	}
	if ret.Get(0).(*MockedStruct) != m {
		t.Error("fail")
	}
	if ret.GetType(0, m).(*MockedStruct) != m {
		t.Error("fail")
	}

	ret = m.FuncMockResult("nil")
	if ret.GetType(0, m).(*MockedStruct) != nil {
		t.Error("fail")
	}

	ret = m.FuncMockResult("nil.with.type")
	if ret.GetType(0, m).(*MockedStruct) != nil {
		t.Error("fail")
	}

	ret = m.FuncMockResult("bool")
	if ret.Bool(0) != true {
		t.Error("fail")
	}

	ret = m.FuncMockResult("byte")
	if ret.Byte(0) != byte('a') {
		t.Error("fail")
	}

	ret = m.FuncMockResult("bytes")
	if string(ret.Bytes(0)) != "abc" {
		t.Error("fail")
	}

	ret = m.FuncMockResult("bytes.nil")
	if ret.Bytes(0) != nil {
		t.Error("fail")
	}

	ret = m.FuncMockResult("error")
	if ret.Error(0).Error() != "error" {
		t.Error("fail")
	}

	ret = m.FuncMockResult("float32")
	if ret.Float32(0) != 1.32 {
		t.Error("fail")
	}

	ret = m.FuncMockResult("float64")
	if ret.Float64(0) != 1.64 {
		t.Error("fail")
	}

	ret = m.FuncMockResult("int")
	if ret.Int(0) != 1 {
		t.Error("fail")
	}

	ret = m.FuncMockResult("int8")
	if ret.Int8(0) != 8 {
		t.Error("fail")
	}

	ret = m.FuncMockResult("int16")
	if ret.Int16(0) != 16 {
		t.Error("fail")
	}

	ret = m.FuncMockResult("int32")
	if ret.Int32(0) != 32 {
		t.Error("fail")
	}

	ret = m.FuncMockResult("int64")
	if ret.Int64(0) != 64 {
		t.Error("fail")
	}

	ret = m.FuncMockResult("string")
	if ret.String(0) != "string" {
		t.Error("fail")
	}

	ret = m.FuncMockResult("chan")
	c := ret.Get(0)
	v := reflect.ValueOf(c)
	if v.Type().String() != "chan int" {
		t.Error("fail")
	}
}

func TestMockResultDefaults(t *testing.T) {
	m := &MockedStruct{}
	m.When("FuncMockResult", 1)

	ret := m.FuncMockResult(1)

	// Test all cases
	if ret.Contains(0) != false {
		t.Error("fail")
	}
	if ret.Get(0) != nil {
		t.Error("fail")
	}
	if ret.Bool(0) != false {
		t.Error("fail")
	}
	if ret.Byte(0) != 0 {
		t.Error("fail")
	}
	if ret.Bytes(0) != nil {
		t.Error("fail")
	}
	if ret.Error(0) != nil {
		t.Error("fail")
	}
	if ret.Float32(0) != 0 {
		t.Error("fail")
	}
	if ret.Float64(0) != 0 {
		t.Error("fail")
	}
	if ret.Int(0) != 0 {
		t.Error("fail")
	}
	if ret.Int8(0) != 0 {
		t.Error("fail")
	}
	if ret.Int16(0) != 0 {
		t.Error("fail")
	}
	if ret.Int32(0) != 0 {
		t.Error("fail")
	}
	if ret.Int64(0) != 0 {
		t.Error("fail")
	}
	if ret.String(0) != "" {
		t.Error("fail")
	}
}

func TestVerifyMocks(t *testing.T) {
	good := &Mock{}
	bad1 := &Mock{}
	bad2 := &Mock{}
	bad1.When("foo").Times(1)
	bad1.When("bar").Times(1)
	if ok, err := VerifyMocks(good, good, good); !ok {
		t.Error(err)
	}
	ok, err := VerifyMocks(good, bad1, bad2)
	if ok {
		t.Fail()
	}
	_, bad1Error := bad1.Verify()
	if err.Error() != bad1Error.Error() {
		t.Errorf("Expected verification error %s, found %s", bad1Error, err)
	}
}

func TestSlice(t *testing.T) {
	var m = &MockedStruct{}

	t.Log("Match")

	for _, test := range []struct {
		a, e []interface{}
	}{
		{nil, nil},
		{nil, []interface{}{}},
		{nil, []interface{}{Rest}},
		{[]interface{}{}, nil},
		{[]interface{}{}, []interface{}{}},
		{[]interface{}{}, []interface{}{Rest}},

		{[]interface{}{1}, []interface{}{1}},
		{[]interface{}{1}, []interface{}{Rest}},
		{[]interface{}{1}, []interface{}{1, Rest}},

		{[]interface{}{1, 2}, []interface{}{1, 2}},
		{[]interface{}{1, 2}, []interface{}{Rest}},
		{[]interface{}{1, 2}, []interface{}{1, Rest}},
		{[]interface{}{1, 2}, []interface{}{1, 2, Rest}},

		{[]interface{}{1, 2, 3}, []interface{}{1, 2, 3}},
		{[]interface{}{1, 2, 3}, []interface{}{Rest}},
		{[]interface{}{1, 2, 3}, []interface{}{1, Rest}},
		{[]interface{}{1, 2, 3}, []interface{}{1, 2, Rest}},
		{[]interface{}{1, 2, 3}, []interface{}{1, 2, 3, Rest}},
	} {
		t.Log("Test:", test)

		m.Reset()
		m.When("FuncVariadic", 1, Slice(test.e...))
		m.FuncVariadic(1, test.a...)
	}

	t.Log("No match")

	var try = func(f func()) (panicked bool) {
		defer func() {
			if v := recover(); v != nil {
				panicked = true
			}

			return
		}()

		f()

		return false
	}

	for _, test := range []struct {
		a, e []interface{}
	}{
		{nil, []interface{}{1}},
		{nil, []interface{}{1, Rest}},
		{nil, []interface{}{1, 2}},
		{nil, []interface{}{1, 2, Rest}},
		{nil, []interface{}{1, 2, 3}},
		{nil, []interface{}{1, 2, 3, Rest}},

		{[]interface{}{}, []interface{}{1}},
		{[]interface{}{}, []interface{}{1, Rest}},
		{[]interface{}{}, []interface{}{1, 2}},
		{[]interface{}{}, []interface{}{1, 2, Rest}},
		{[]interface{}{}, []interface{}{1, 2, 3}},
		{[]interface{}{}, []interface{}{1, 2, 3, Rest}},

		{[]interface{}{1}, []interface{}{1, 2}},
		{[]interface{}{1}, []interface{}{1, 2, Rest}},
		{[]interface{}{1}, []interface{}{1, 2, 3}},
		{[]interface{}{1}, []interface{}{1, 2, 3, Rest}},

		{[]interface{}{1, 2}, []interface{}{1, 2, 3}},
		{[]interface{}{1, 2}, []interface{}{1, 2, 3, Rest}},

		{[]interface{}{1, 2}, []interface{}{2, 1}},
		{[]interface{}{1, 2}, []interface{}{2, 1, Rest}},

		{[]interface{}{1, 2, 3}, []interface{}{3, 2, 1}},
		{[]interface{}{1, 2, 3}, []interface{}{3, 2, 1, Rest}},
	} {
		t.Log("Test:", test)

		m.Reset()
		m.When("FuncVariadic", 1, Slice(test.e...))

		if !try(func() { m.FuncVariadic(1, test.a...) }) {
			t.Errorf("Actual no panic, expected panic")
		}
	}
}
