package kvscheduler

import (
	"github.com/gogo/protobuf/proto"
	"github.com/vishvananda/netns"
	"os"

	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

// checking of the original network namespace preservation
var defaultNs netns.NsHandle
var enableNetNsCheck = os.Getenv(checkNetNamespaceEnv) != ""

func init() {
	defaultNs, _ = netns.Get()
}

func checkNetNs() error {
	if enableNetNsCheck && defaultNs != -1 {
		ns, nsErr := netns.Get()
		if nsErr == nil {
			defer ns.Close()
		}
		if nsErr == nil && !defaultNs.Equal(ns) {
			return kvs.ErrEscapedNetNs
		}
	}
	return nil
}

// descriptorHandler handles access to descriptor methods (callbacks).
// For callback not provided, a default return value is returned.
type descriptorHandler struct {
	descriptor *kvs.KVDescriptor
}

// newDescriptorHandler is a constructor for descriptor handler
func newDescriptorHandler(descr *kvs.KVDescriptor) *descriptorHandler {
	return &descriptorHandler{
		descriptor: descr,
	}
}

// keyLabel by default returns the key itself.
func (h *descriptorHandler) keyLabel(key string) string {
	if h.descriptor == nil || h.descriptor.KeyLabel == nil {
		return key
	}
	defer trackDescMethod(h.descriptor.Name, "KeyLabel")()
	return h.descriptor.KeyLabel(key)
}

// equivalentValues by default uses proto.Equal().
func (h *descriptorHandler) equivalentValues(key string, oldValue, newValue proto.Message) bool {
	if h.descriptor == nil || h.descriptor.ValueComparator == nil {
		return proto.Equal(oldValue, newValue)
	}
	defer trackDescMethod(h.descriptor.Name, "ValueComparator")()
	return h.descriptor.ValueComparator(key, oldValue, newValue)
}

// validate return nil if Validate is not provided (optional method).
func (h *descriptorHandler) validate(key string, value proto.Message) error {
	if h.descriptor == nil || h.descriptor.Validate == nil {
		return nil
	}
	defer trackDescMethod(h.descriptor.Name, "Validate")()
	return h.descriptor.Validate(key, value)
}

// create returns ErrUnimplementedCreate if Create is not provided.
func (h *descriptorHandler) create(key string, value proto.Message) (metadata kvs.Metadata, err error) {
	if h.descriptor == nil {
		return
	}
	if h.descriptor.Create == nil {
		return nil, kvs.ErrUnimplementedCreate
	}
	defer trackDescMethod(h.descriptor.Name, "Create")()
	metadata, err = h.descriptor.Create(key, value)
	if nsErr := checkNetNs(); nsErr != nil {
		err = nsErr
	}
	return metadata, err

}

// update is not called if Update is not provided (updateWithRecreate() returns true).
func (h *descriptorHandler) update(key string, oldValue, newValue proto.Message, oldMetadata kvs.Metadata) (newMetadata kvs.Metadata, err error) {
	if h.descriptor == nil {
		return oldMetadata, nil
	}
	defer trackDescMethod(h.descriptor.Name, "Update")()
	newMetadata, err = h.descriptor.Update(key, oldValue, newValue, oldMetadata)
	if nsErr := checkNetNs(); nsErr != nil {
		err = nsErr
	}
	return newMetadata, err
}

// updateWithRecreate either forwards the call to UpdateWithRecreate if defined
// by the descriptor, or decides based on the availability of the Update operation.
func (h *descriptorHandler) updateWithRecreate(key string, oldValue, newValue proto.Message, metadata kvs.Metadata) bool {
	if h.descriptor == nil {
		return false
	}
	if h.descriptor.Update == nil {
		// without Update, re-creation is the only way
		return true
	}
	if h.descriptor.UpdateWithRecreate == nil {
		// by default it is assumed that any change can be applied using Update without
		// re-creation
		return false
	}
	defer trackDescMethod(h.descriptor.Name, "UpdateWithRecreate")()
	return h.descriptor.UpdateWithRecreate(key, oldValue, newValue, metadata)
}

// delete returns ErrUnimplementedDelete if Delete is not provided.
func (h *descriptorHandler) delete(key string, value proto.Message, metadata kvs.Metadata) error {
	if h.descriptor == nil {
		return nil
	}
	if h.descriptor.Delete == nil {
		return kvs.ErrUnimplementedDelete
	}
	defer trackDescMethod(h.descriptor.Name, "Delete")()
	err := h.descriptor.Delete(key, value, metadata)
	if nsErr := checkNetNs(); nsErr != nil {
		err = nsErr
	}
	return err
}

// isRetriableFailure first checks for errors returned by the handler itself.
// If descriptor does not define IsRetriableFailure, it is assumed any failure
// can be potentially fixed by retry.
func (h *descriptorHandler) isRetriableFailure(err error) bool {
	// first check for errors returned by the handler itself
	handlerErrs := []error{
		kvs.ErrUnimplementedCreate,
		kvs.ErrUnimplementedDelete,
		kvs.ErrEscapedNetNs}
	for _, handlerError := range handlerErrs {
		if err == handlerError {
			return false
		}
	}
	if h.descriptor == nil || h.descriptor.IsRetriableFailure == nil {
		return true
	}
	defer trackDescMethod(h.descriptor.Name, "IsRetriableFailure")()
	return h.descriptor.IsRetriableFailure(err)
}

// dependencies returns empty list if descriptor does not define any.
func (h *descriptorHandler) dependencies(key string, value proto.Message) (deps []kvs.Dependency) {
	if h.descriptor == nil || h.descriptor.Dependencies == nil {
		return
	}
	defer trackDescMethod(h.descriptor.Name, "Dependencies")()
	return h.descriptor.Dependencies(key, value)
}

// derivedValues returns empty list if descriptor does not define any.
func (h *descriptorHandler) derivedValues(key string, value proto.Message) (derives []kvs.KeyValuePair) {
	if h.descriptor == nil || h.descriptor.DerivedValues == nil {
		return
	}
	defer trackDescMethod(h.descriptor.Name, "DerivedValues")()
	return h.descriptor.DerivedValues(key, value)
}

// retrieve returns <ableToRetrieve> as false if descriptor does not implement Retrieve.
func (h *descriptorHandler) retrieve(correlate []kvs.KVWithMetadata) (values []kvs.KVWithMetadata, ableToRetrieve bool, err error) {
	if h.descriptor == nil || h.descriptor.Retrieve == nil {
		return values, false, nil
	}
	defer trackDescMethod(h.descriptor.Name, "Retrieve")()
	values, err = h.descriptor.Retrieve(correlate)
	if nsErr := checkNetNs(); nsErr != nil {
		err = nsErr
	}
	return values, true, err
}
