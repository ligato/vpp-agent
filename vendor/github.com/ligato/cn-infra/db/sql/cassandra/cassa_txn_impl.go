package cassandra

import (
	"github.com/ligato/cn-infra/db/sql"
)

// NewTxn creates a new Data Broker transaction. A transaction can
// hold multiple operations that are all committed to the data
// store together. After a transaction has been created, one or
// more operations (put or delete) can be added to the transaction
// before it is committed.
func (pdb *BrokerCassa) NewTxn() sql.Txn {
	// TODO Cassandra Batch/TXN
	panic("not implemented")
}
