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

// TestPut1_convenient is most convenient way of putting one entity to cassandra
func TestPut1_convenient(t *testing.T) {
	gomega.RegisterTestingT(t)

	session := mockSession()
	defer session.Close()
	db := cassandra.NewBrokerUsingSession(session)

	mockPut(session, "UPDATE User SET id = ?, first_name = ?, last_name = ? WHERE id = ?",
		[]interface{}{
			"James Bond",
			"James",
			"Bond",
		})

	err := db.Put(sql.FieldEQ(&JamesBond.ID), JamesBond)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}

// TestPut2_EQ is most convenient way of putting one entity to cassandra
func TestPut2_EQ(t *testing.T) {
	gomega.RegisterTestingT(t)

	session := mockSession()
	defer session.Close()
	db := cassandra.NewBrokerUsingSession(session)

	mockPut(session, "UPDATE User SET id = ?, first_name = ?, last_name = ? WHERE id = ?",
		[]interface{}{
			"James Bond",
			"James",
			"Bond",
		})

	err := db.Put(sql.Field(&JamesBond.ID, sql.EQ(JamesBond.ID)), JamesBond)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}
