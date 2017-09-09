package gockle

import (
	"fmt"

	"github.com/maraino/go-mock"
)

var mySession = &SessionMock{}

func ExampleIterator_ScanMap() {
	var iteratorMock = &IteratorMock{}

	iteratorMock.When("ScanMap", mock.Any).Call(func(m map[string]interface{}) bool {
		m["id"] = 1
		m["name"] = "alex"

		return false
	})

	iteratorMock.When("Close").Return(nil)

	var sessionMock = &SessionMock{}

	const query = "select * from users"

	sessionMock.When("ScanIterator", query, mock.Any).Return(iteratorMock)
	sessionMock.When("Close").Return()

	var session Session = sessionMock
	var iterator = session.ScanIterator(query)
	var row = map[string]interface{}{}

	for more := true; more; {
		more = iterator.ScanMap(row)

		fmt.Printf("id = %v, name = %v\n", row["id"], row["name"])
	}

	if err := iterator.Close(); err != nil {
		fmt.Println(err)
	}

	session.Close()

	// Output: id = 1, name = alex
}

func ExampleSession_Batch() {
	var batchMock = &BatchMock{}

	batchMock.When("Add", "insert into users (id, name) values (1, 'alex')", mock.Any).Return()
	batchMock.When("Exec").Return(fmt.Errorf("invalid"))

	var sessionMock = &SessionMock{}

	sessionMock.When("Batch", BatchLogged).Return(batchMock)
	sessionMock.When("Close").Return()

	var session Session = sessionMock
	var batch = session.Batch(BatchLogged)

	batch.Add("insert into users (id, name) values (1, 'alex')")

	if err := batch.Exec(); err != nil {
		fmt.Println(err)
	}

	session.Close()

	// Output: invalid
}

func ExampleSession_ScanMapSlice() {
	var sessionMock = &SessionMock{}

	const query = "select * from users"

	sessionMock.When("ScanMapSlice", query, mock.Any).Return([]map[string]interface{}{{"id": 1, "name": "alex"}}, nil)
	sessionMock.When("Close").Return()

	var session Session = sessionMock
	var rows, _ = session.ScanMapSlice(query)

	for _, row := range rows {
		fmt.Printf("id = %v, name = %v\n", row["id"], row["name"])
	}

	session.Close()

	// Output: id = 1, name = alex
}
