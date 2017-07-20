// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cassandra_test

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/ligato/cn-infra/db/sql"
	"github.com/ligato/cn-infra/db/sql/cassandra"
)

// TestListValues1_convenient is most convenient way of selecting slice of entities
// User of the API does not need to write SQL string (string is calculated from the entity type.
// User of the API does not need to use the Iterator (user gets directly slice of th entity type - reflection needed).
func TestListValues1_convenient(t *testing.T) {
	gomega.RegisterTestingT(t)

	session := mockSession()
	defer session.Close()
	db := cassandra.NewBrokerUsingSession(session)

	query := sql.FROM(UserTable, sql.WHERE(sql.Field(&UserTable.LastName, sql.EQ("Bond"))))
	mockQuery(session, query, cells(JamesBond), cells(PeterBond))

	users := &[]User{}
	err := sql.SliceIt(users, db.ListValues(query))

	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(users).ToNot(gomega.BeNil())
	gomega.Expect(users).To(gomega.BeEquivalentTo(&[]User{*JamesBond, *PeterBond}))
}

/*
// TestListValues2_constFieldname let's user to write field name in where statement in old way using string constant.
// All other is same as in TestListValues1
func TestListValues2_constFieldname(t *testing.T) {
	gomega.RegisterTestingT(t)

	session := mockSession()
	defer session.Close()
	db := cassandra.NewBrokerUsingSession(session)

	query := sql.SelectFrom(UserTable) + sql.Where(sql.EQ("last_name", "Bond"))
	mockQuery(session, query, cells(JamesBond), cells(PeterBond))

	users := &[]User{}
	err := sql.SliceIt(users, db.ListValues(query))

	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(users).ToNot(gomega.BeNil())
	gomega.Expect(users).To(gomega.BeEquivalentTo(&[]User{JamesBond, PeterBond}))
}

// TestListValues3_customSQL let's user to write part of the SQL statement/query
// All other is same as in TestListValues1
func TestListValues3_customSQL(t *testing.T) {
	gomega.RegisterTestingT(t)

	session := mockSession()
	defer session.Close()
	db := cassandra.NewBrokerUsingSession(session)

	query := sql.SelectFrom(UserTable) + "WHERE last_name = 'Bond'"
	mockQuery(session, query, cells(JamesBond), cells(PeterBond))

	users := &[]User{}
	err := sql.SliceIt(users, db.ListValues(query))

	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(users).ToNot(gomega.BeNil())
	gomega.Expect(users).To(gomega.BeEquivalentTo(&[]User{JamesBond, PeterBond}))
}

// TestListValues4_iterator does not use reflection to fill slice of users (but the iterator)
// All other is same as in TestListValues1
func TestListValues4_iterator(t *testing.T) {
	gomega.RegisterTestingT(t)

	session := mockSession()
	defer session.Close()
	db := cassandra.NewBrokerUsingSession(session)

	query := sql.SelectFrom(UserTable) + sql.Where(sql.Field(&UserTable.LastName, UserTable, "Bond"))
	mockQuery(session, query, cells(JamesBond), cells(PeterBond))

	users := []*User{}
	it := db.ListValues(query)
	for {
		user := &User{}
		stop := it.GetNext(user)
		if stop {
			break
		}
		users = append(users, user)
	}
	err := it.Close()

	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(users).ToNot(gomega.BeNil())
	gomega.Expect(users).To(gomega.BeEquivalentTo([]*User{&JamesBond, &PeterBond}))
}


// TestListValues4_iteratorScanMap does not use reflection to fill slice of users (but the iterator)
// All other is same as in TestListValues1
func TestListValues4_iteratorScanMap(t *testing.T) {
	gomega.RegisterTestingT(t)

	session := mockSession()
	defer session.Close()
	db := cassandra.NewBrokerUsingSession(session)

	query := sql.SelectFrom(UserTable) + sql.Where(sql.Field(&UserTable.LastName, UserTable, "Bond"))
	mockQuery(session, query, cells(JamesBond), cells(PeterBond))

	it := db.ListValues(query)
	for {
		user := map[string]interface{}{}
		stop := it.GetNext(user)
		if stop {
			break
		}
		fmt.Println("user: ", user)
	}
	err := it.Close()

	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	//gomega.Expect(users).ToNot(gomega.BeNil())
	//gomega.Expect(users).To(gomega.BeEquivalentTo([]*User{&JamesBond, &PeterBond}))
}
*/
