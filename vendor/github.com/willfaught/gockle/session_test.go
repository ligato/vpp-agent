package gockle

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/maraino/go-mock"
)

const version = 4

const (
	ksCreate  = "create keyspace gockle_test with replication = {'class': 'SimpleStrategy', 'replication_factor': 1};"
	ksDrop    = "drop keyspace gockle_test"
	ksDropIf  = "drop keyspace if exists gockle_test"
	rowInsert = "insert into gockle_test.test (id, n) values (1, 2)"
	tabCreate = "create table gockle_test.test(id int primary key, n int)"
	tabDrop   = "drop table gockle_test.test"
)

func TestNewSession(t *testing.T) {
	if a, e := NewSession(nil), (session{}); a != e {
		t.Errorf("Actual session %v, expected %v", a, e)
	}

	var c = gocql.NewCluster("localhost")

	c.ProtoVersion = version

	var s, err = c.CreateSession()

	if err != nil {
		t.Skip(err)
	}

	if a, e := NewSession(s), (session{s: s}); a != e {
		t.Errorf("Actual session %v, expected %v", a, e)
	}
}

func TestNewSimpleSession(t *testing.T) {
	if s, err := NewSimpleSession(); err == nil {
		t.Error("Actual no error, expected error")
	} else if s != nil {
		t.Errorf("Actual session %v, expected nil", s)
		s.Close()
	}

	if a, err := NewSimpleSession("localhost"); err != nil {
		t.Skip(err)
	} else if a == nil {
		t.Errorf("Actual session nil, expected not nil")
	} else {
		a.Close()
	}
}

func TestSessionMetadata(t *testing.T) {
	var exec = func(s Session, q string) {
		if err := s.Exec(q); err != nil {
			t.Fatalf("Actual error %v, expected no error", err)
		}
	}

	var s = newSession(t)

	exec(s, ksDropIf)
	exec(s, ksCreate)

	defer exec(s, ksDrop)

	exec(s, tabCreate)

	defer exec(s, tabDrop)

	s = newSession(t)

	if a, err := s.Tables("gockle_test"); err == nil {
		if e := ([]string{"test"}); !reflect.DeepEqual(a, e) {
			t.Fatalf("Actual tables %v, expected %v", a, e)
		}
	} else {
		t.Fatalf("Actual error %v, expected no error", err)
	}

	if _, err := s.Tables("gockle_test_invalid"); err == nil {
		t.Errorf("Actual no error, expected error")
	}

	s.Close()

	if _, err := s.Tables("gockle_test"); err == nil {
		t.Errorf("Actual no error, expected error")
	}

	s = newSession(t)

	if a, err := s.Columns("gockle_test", "test"); err == nil {
		var ts = map[string]gocql.Type{"id": gocql.TypeInt, "n": gocql.TypeInt}

		if la, le := len(a), len(ts); la == le {
			for n, at := range a {
				if et, ok := ts[n]; ok {
					if at.Type() != et {
						t.Fatalf("Actual type %v, expected %v", at, et)
					}
				} else {
					t.Fatalf("Actual name %v invalid, expected valid", n)
				}
			}
		} else {
			t.Fatalf("Actual count %v, expected %v", la, le)
		}
	} else {
		t.Fatalf("Actual error %v, expected no error", err)
	}

	if _, err := s.Columns("gockle_test", "invalid"); err == nil {
		t.Error("Actual no error, expected error")
	}

	s.Close()

	if _, err := s.Columns("gockle_test", "test"); err == nil {
		t.Error("Actual no error, expected error")
	}
}

func TestSessionMock(t *testing.T) {
	var m, e = &SessionMock{}, fmt.Errorf("e")

	testMock(t, m, &m.Mock, []struct {
		method    string
		arguments []interface{}
		results   []interface{}
	}{
		{"Close", nil, nil},
		{"Columns", []interface{}{"", ""}, []interface{}{map[string]gocql.TypeInfo(nil), nil}},
		{"Columns", []interface{}{"a", "b"}, []interface{}{map[string]gocql.TypeInfo{"c": gocql.NativeType{}}, e}},
		{"Batch", []interface{}{BatchKind(0)}, []interface{}{(*batch)(nil)}},
		{"Batch", []interface{}{BatchKind(1)}, []interface{}{&batch{}}},
		{"Exec", []interface{}{"", []interface{}(nil)}, []interface{}{nil}},
		{"Exec", []interface{}{"a", []interface{}{1}}, []interface{}{e}},
		{"Scan", []interface{}{"", []interface{}(nil), []interface{}(nil)}, []interface{}{nil}},
		{"Scan", []interface{}{"a", []interface{}{1}, []interface{}{1}}, []interface{}{e}},
		{"ScanIterator", []interface{}{"", []interface{}(nil)}, []interface{}{(*iterator)(nil)}},
		{"ScanIterator", []interface{}{"a", []interface{}{1}}, []interface{}{iterator{}}},
		{"ScanMap", []interface{}{"", map[string]interface{}(nil), []interface{}(nil)}, []interface{}{nil}},
		{"ScanMap", []interface{}{"a", map[string]interface{}{"b": 2}, []interface{}{1}}, []interface{}{e}},
		{"ScanMapSlice", []interface{}{"", []interface{}(nil)}, []interface{}{[]map[string]interface{}(nil), nil}},
		{"ScanMapSlice", []interface{}{"a", []interface{}{1}}, []interface{}{[]map[string]interface{}{{"b": 2}}, e}},
		{"ScanMapTx", []interface{}{"", map[string]interface{}(nil), []interface{}(nil)}, []interface{}{false, nil}},
		{"ScanMapTx", []interface{}{"a", map[string]interface{}{"b": 2}, []interface{}{1}}, []interface{}{true, e}},
		{"Tables", []interface{}{""}, []interface{}{[]string(nil), nil}},
		{"Tables", []interface{}{"a"}, []interface{}{[]string{"b"}, e}},
	})
}

func TestSessionQuery(t *testing.T) {
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

	// Batch
	if s.Batch(BatchKind(0)) == nil {
		t.Error("Actual batch nil, expected not nil")
	}

	// ScanIterator
	if s.ScanIterator("select * from gockle_test.test") == nil {
		t.Error("Actual iterator nil, expected not nil")
	}

	// Scan
	var id, n int

	if err := s.Scan("select id, n from gockle_test.test", []interface{}{&id, &n}); err == nil {
		if id != 1 {
			t.Errorf("Actual id %v, expected 1", id)
		}

		if n != 2 {
			t.Errorf("Actual n %v, expected 2", n)
		}
	} else {
		t.Errorf("Actual error %v, expected no error", err)
	}

	// ScanMap
	var am, em = map[string]interface{}{}, map[string]interface{}{"id": 1, "n": 2}

	if err := s.ScanMap("select id, n from gockle_test.test", am); err == nil {
		if !reflect.DeepEqual(am, em) {
			t.Errorf("Actual map %v, expected %v", am, em)
		}
	} else {
		t.Errorf("Actual error %v, expected no error", err)
	}

	// ScanMapTx
	am = map[string]interface{}{}

	if b, err := s.ScanMapTx("update gockle_test.test set n = 3 where id = 1 if n = 2", am); err == nil {
		if !b {
			t.Error("Actual applied false, expected true")
		}

		if l := len(am); l != 0 {
			t.Errorf("Actual length %v, expected 0", l)
		}
	} else {
		t.Errorf("Actual error %v, expected no error", err)
	}

	// ScanMapSlice
	var es = []map[string]interface{}{{"id": 1, "n": 3}}

	if as, err := s.ScanMapSlice("select * from gockle_test.test"); err == nil {
		if !reflect.DeepEqual(as, es) {
			t.Errorf("Actual rows %v, expected %v", as, es)
		}
	} else {
		t.Errorf("Actual error %v, expected no error", err)
	}
}

func newSession(t *testing.T) Session {
	var c = gocql.NewCluster("localhost")

	c.ProtoVersion = version
	c.Timeout = 5 * time.Second

	var s, err = c.CreateSession()

	if err != nil {
		t.Skip(err)
	}

	return NewSession(s)
}

func testMock(t *testing.T, i interface{}, m *mock.Mock, tests []struct {
	method    string
	arguments []interface{}
	results   []interface{}
}) {
	var v = reflect.ValueOf(i)

	for _, test := range tests {
		t.Log("Test:", test)
		m.Reset()
		m.When(test.method, test.arguments...).Return(test.results...)

		var vs []reflect.Value

		for _, a := range test.arguments {
			vs = append(vs, reflect.ValueOf(a))
		}

		var method = v.MethodByName(test.method)

		if method.Type().IsVariadic() {
			vs = method.CallSlice(vs)
		} else {
			vs = method.Call(vs)
		}

		var is []interface{}

		for _, v := range vs {
			is = append(is, v.Interface())
		}

		if !reflect.DeepEqual(is, test.results) {
			t.Errorf("Actual %v, expected %v", is, test.results)
		}
	}
}
