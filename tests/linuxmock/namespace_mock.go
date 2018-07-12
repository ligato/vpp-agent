package linuxmock

import (
	"github.com/ligato/vpp-agent/plugins/linux/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/linux/model/l3"
	"github.com/ligato/vpp-agent/plugins/linux/nsplugin"
)

// NamespacePluginMock allows to mock namespace plugin methods to manage namespaces and microservices
type NamespacePluginMock struct {
	responses []*whenNsResp
	respCurr  int
	respMax   int
}

// NewNamespacePluginMock creates new instance of the mock and initializes response list
func NewNamespacePluginMock() *NamespacePluginMock {
	return &NamespacePluginMock{
		responses: make([]*whenNsResp, 0),
	}
}

// Helper struct with single method call and desired response items
type whenNsResp struct {
	methodName string
	items      []interface{}
}

// When defines name of the related method. It creates a new instance of whenNsResp with provided method name and
// stores it to the mock.
func (mock *NamespacePluginMock) When(methodName string) *whenNsResp {
	resp := &whenNsResp{
		methodName: methodName,
	}
	mock.responses = append(mock.responses, resp)
	return resp
}

// ThenReturn receives array of items, which are desired to be returned in mocked method defined in "When". The full
// logic is:
// - When('someMethod').ThenReturn('values')
//
// Provided values should match return types of method. If method returns multiple values and only one is provided,
// mock tries to parse the value and returns it, while others will be nil or empty.
//
// If method is called several times, all cases must be defined separately, even if the return value is the same:
// - When('method1').ThenReturn('val1')
// - When('method1').ThenReturn('val1')
//
// All mocked methods are evaluated in same order they were assigned.
func (when *whenNsResp) ThenReturn(item ...interface{}) {
	when.items = item
}

// Auxiliary method returns next return value for provided method as generic type
func (mock *NamespacePluginMock) getReturnValues(name string) (response []interface{}) {
	for i, resp := range mock.responses {
		if resp.methodName == name {
			// Remove used response but retain order
			mock.responses = append(mock.responses[:i], mock.responses[i+1:]...)
			return resp.items
		}
	}
	// Return empty response
	return
}

/* Mocked netlink handler methods */ //todo define other

func (mock *NamespacePluginMock) IsNamespaceAvailable(ns *interfaces.LinuxInterfaces_Interface_Namespace) bool {
	items := mock.getReturnValues("IsNamespaceAvailable")
	return items[0].(bool)
}

func (mock *NamespacePluginMock) SwitchNamespace(ns *nsplugin.Namespace, ctx *nsplugin.NamespaceMgmtCtx) (func(), error) {
	items := mock.getReturnValues("SwitchNamespace")
	if len(items) == 1 {
		switch typed := items[0].(type) {
		case func():
			return typed, nil
		case error:
			return func() {}, typed
		}
	} else if len(items) == 2 {
		return items[0].(func()), items[1].(error)
	}
	return func() {}, nil
}

func (mock *NamespacePluginMock) SwitchToNamespace(nsMgmtCtx *nsplugin.NamespaceMgmtCtx, ns *interfaces.LinuxInterfaces_Interface_Namespace) (func(), error) {
	items := mock.getReturnValues("SwitchToNamespace")
	if len(items) == 1 {
		switch typed := items[0].(type) {
		case func():
			return typed, nil
		case error:
			return func() {}, typed
		}
	} else if len(items) == 2 {
		return items[0].(func()), items[1].(error)
	}
	return func() {}, nil
}

func (mock *NamespacePluginMock) SetInterfaceNamespace(ctx *nsplugin.NamespaceMgmtCtx, ifName string, namespace *interfaces.LinuxInterfaces_Interface_Namespace) error {
	items := mock.getReturnValues("SetInterfaceNamespace")
	if len(items) >= 1 {
		return items[0].(error)
	}
	return nil
}

func (mock *NamespacePluginMock) GetConfigNamespace() *interfaces.LinuxInterfaces_Interface_Namespace {
	items := mock.getReturnValues("GetConfigNamespace")
	if len(items) >= 1 {
		return items[0].(*interfaces.LinuxInterfaces_Interface_Namespace)
	}
	return nil
}

func (mock *NamespacePluginMock) IfaceNsToString(namespace *interfaces.LinuxInterfaces_Interface_Namespace) string {
	items := mock.getReturnValues("IfaceNsToString")
	if len(items) >= 1 {
		return items[0].(string)
	}
	return ""
}

func (mock *NamespacePluginMock) IfNsToGeneric(ns *interfaces.LinuxInterfaces_Interface_Namespace) *nsplugin.Namespace {
	items := mock.getReturnValues("IfNsToGeneric")
	if len(items) >= 1 {
		return items[0].(*nsplugin.Namespace)
	}
	return nil
}

func (mock *NamespacePluginMock) ArpNsToGeneric(ns *l3.LinuxStaticArpEntries_ArpEntry_Namespace) *nsplugin.Namespace {
	items := mock.getReturnValues("ArpNsToGeneric")
	if len(items) >= 1 {
		return items[0].(*nsplugin.Namespace)
	}
	return nil
}

func (mock *NamespacePluginMock) GenericToArpNs(ns *nsplugin.Namespace) (*l3.LinuxStaticArpEntries_ArpEntry_Namespace, error) {
	items := mock.getReturnValues("GenericToArpNs")
	if len(items) == 1 {
		switch typed := items[0].(type) {
		case *l3.LinuxStaticArpEntries_ArpEntry_Namespace:
			return typed, nil
		case error:
			return nil, typed
		}
	} else if len(items) == 2 {
		return items[0].(*l3.LinuxStaticArpEntries_ArpEntry_Namespace), items[1].(error)
	}
	return nil, nil
}

func (mock *NamespacePluginMock) RouteNsToGeneric(ns *l3.LinuxStaticRoutes_Route_Namespace) *nsplugin.Namespace {
	items := mock.getReturnValues("RouteNsToGeneric")
	if len(items) >= 1 {
		return items[0].(*nsplugin.Namespace)
	}
	return nil
}

func (mock *NamespacePluginMock) HandleMicroservices(ctx *nsplugin.MicroserviceCtx) {}
