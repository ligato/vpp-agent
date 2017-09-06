package gockle

import (
	"fmt"
	"reflect"
	"testing"
)

func TestBatch(t *testing.T) {
	var execs = newSession(t)

	var exec = func(q string) {
		if err := execs.Exec(q); err != nil {
			t.Fatalf("Actual error %v, expected no error", err)
		}
	}

	exec(ksDropIf)
	exec(ksCreate)
	exec(tabCreate)
	exec(rowInsert)

	defer execs.Close()
	defer exec(ksDrop)
	defer exec(tabDrop)

	// Exec
	var s = newSession(t)
	var b = s.Batch(BatchKind(0))

	if b == nil {
		t.Error("Actual batch nil, expected not nil")
	}

	b.Add("update gockle_test.test set n = 3 where id = 1 if n = 2")

	if err := b.Exec(); err != nil {
		t.Errorf("Actual error %v, expected no error", err)
	}

	// ExecTx
	b = s.Batch(BatchKind(0))
	b.Add("update gockle_test.test set n = 4 where id = 1 if n = 3")

	if a, err := b.ExecTx(); err == nil {
		if e := ([]map[string]interface{}{{"[applied]": true}}); !reflect.DeepEqual(a, e) {
			t.Errorf("Actual tx %v, expected %v", a, e)
		}
	} else {
		t.Errorf("Actual error %v, expected no error", err)
	}

	s.Close()

	if _, err := b.ExecTx(); err == nil {
		t.Error("Actual no error, expected error")
	}
}

func TestBatchMock(t *testing.T) {
	var m, e = &BatchMock{}, fmt.Errorf("e")

	testMock(t, m, &m.Mock, []struct {
		method    string
		arguments []interface{}
		results   []interface{}
	}{
		{"Add", []interface{}{"", []interface{}(nil)}, nil},
		{"Add", []interface{}{"a", []interface{}{1}}, nil},
		{"Exec", nil, []interface{}{nil}},
		{"Exec", nil, []interface{}{e}},
		{"ExecTx", nil, []interface{}{([]map[string]interface{})(nil), nil}},
		{"ExecTx", nil, []interface{}{[]map[string]interface{}{}, e}},
	})
}
