package datasync

import (
	"github.com/golang/protobuf/proto"
)

// ChangeEvent is used as the data type for the change channel
// (see the VPP Standard Plugins API). A data change event contains
// a key identifying where the change happened and two values for
// data stored under that key: the value *before* the change (previous
// value) and the value *after* the change (current value).
type ChangeEvent interface {
	CallbackResult

	ProtoWatchResp
}

// ResyncEvent is used as the data type for the resync channel
// (see the ifplugin API)
type ResyncEvent interface {
	CallbackResult

	GetValues() map[ /*keyPrefix*/ string]KeyValIterator
}

// CallbackResult can be used by an event receiver to indicate to the event producer
// whether an operation was successful (error is nil) or unsuccessful (error is
// not nil)
//
// DoneMethod is reused later. There are at least two implementations DoneChannel, DoneCallback
type CallbackResult interface {
	// Done allows plugins that are processing data change/resync to send feedback
	// If there was no error the Done(nil) needs to be called. Use the noError=nil
	// definition for better readability, for example:
	//     Done(noError).
	Done(error)
}

// ProtoWatchResp contains changed value
type ProtoWatchResp interface {
	ChangeValue
	WithKey
	WithPrevValue
}

// ChangeValue represents single propagated change.
type ChangeValue interface {
	LazyValueWithRev
	WithChangeType
}

// LazyValueWithRev defines value that is unmarshalled into proto message on demand with a revision.
// The reason for defining interface with only one method is primary to unify interfaces in this package
type LazyValueWithRev interface {
	LazyValue
	WithRevision
}

// WithKey is a helper interface which intent is to ensure that same
// method declaration is used in different interfaces (composition of interfaces)
type WithKey interface {
	// GetKey returns the key of the pair
	GetKey() string
}

// WithChangeType is a helper interface which intent is to ensure that same
// method declaration is used in different interfaces (composition of interfaces)
type WithChangeType interface {
	GetChangeType() PutDel
}

// WithRevision is a helper interface which intent is to ensure that same
// method declaration is used in different interfaces (composition of interfaces)
type WithRevision interface {
	// GetRevision gets revision of current value
	GetRevision() (rev int64)
}

// WithPrevValue is a helper interface which intent is to ensure that same
// method declaration is used in different interfaces (composition of interfaces)
type WithPrevValue interface {
	// GetPrevValue gets previous value in the data change event.
	// The caller must provide an address of a proto message buffer
	// for each value.
	// returns:
	// - prevValueExist flag is set to 'true' if prevValue was filled
	// - error if value argument can not be properly filled
	GetPrevValue(prevValue proto.Message) (prevValueExist bool, err error)
}

// LazyValue defines value that is unmarshalled into proto message on demand.
// The reason for defining interface with only one method is primary to unify interfaces in this package
type LazyValue interface {
	// GetValue gets the current in the data change event.
	// The caller must provide an address of a proto message buffer
	// for each value.
	// returns:
	// - revision associated with the latest change in the key-value pair
	// - error if value argument can not be properly filled
	GetValue(value proto.Message) error
}
