package gockle

import (
	"fmt"
	"reflect"
	"testing"
)

func TestIterator(t *testing.T) {
	var s = newSession(t)

	defer s.Close()

	var exec = func(q string) {
		if err := s.Exec(q); err != nil {
			t.Fatalf("Actual error %v, expected no error", err)
		}
	}

	exec(ksDropIf)
	exec(ksCreate)

	defer exec(ksDrop)

	exec(tabCreate)

	defer exec(tabDrop)

	exec(rowInsert)

	if i := s.ScanIterator("select * from gockle_test.test"); i == nil {
		t.Error("Actual iterator nil, expected not nil")
	} else {
		var id, n int

		if !i.Scan(&id, &n) {
			t.Errorf("Actual more false, expected true")
		}

		if id != 1 {
			t.Errorf("Actual id %v, expected 1", id)
		}

		if n != 2 {
			t.Errorf("Actual n %v, expected 2", n)
		}

		if err := i.Close(); err != nil {
			t.Errorf("Actual error %v, expected no error", err)
		}
	}

	if i := s.ScanIterator("select * from gockle_test.test"); i == nil {
		t.Error("Actual iterator nil, expected not nil")
	} else {
		var a = map[string]interface{}{}

		if !i.ScanMap(a) {
			t.Errorf("Actual more false, expected true")
		}

		if e := (map[string]interface{}{"id": 1, "n": 2}); !reflect.DeepEqual(a, e) {
			t.Errorf("Actual map %v, expected %v", a, e)
		}

		if err := i.Close(); err != nil {
			t.Errorf("Actual error %v, expected no error", err)
		}
	}
}

func TestIteratorMock(t *testing.T) {
	var m, e = &IteratorMock{}, fmt.Errorf("e")

	testMock(t, m, &m.Mock, []struct {
		method    string
		arguments []interface{}
		results   []interface{}
	}{
		{"Close", nil, []interface{}{nil}},
		{"Close", nil, []interface{}{e}},
		{"Scan", []interface{}{[]interface{}(nil)}, []interface{}{false}},
		{"Scan", []interface{}{[]interface{}{1}}, []interface{}{true}},
		{"ScanMap", []interface{}{map[string]interface{}(nil)}, []interface{}{false}},
		{"ScanMap", []interface{}{map[string]interface{}{"a": 1}}, []interface{}{true}},
	})
}
