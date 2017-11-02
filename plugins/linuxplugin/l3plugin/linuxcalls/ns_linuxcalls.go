// +build !windows,!darwin

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

package linuxcalls

import (
	"fmt"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/model/l3"
)

// ToGenericArpNs converts arp-type namespace to generic type namespace. Such an object can be used to call common
// namespace-related methods
func ToGenericArpNs(ns *l3.LinuxStaticArpEntries_ArpEntry_Namespace) *linuxcalls.Namespace {
	if ns == nil {
		return &linuxcalls.Namespace{}
	}
	return &linuxcalls.Namespace{Type: int32(ns.Type), Pid: ns.Pid, Microservice: ns.Microservice, Name: ns.Name, Filepath: ns.Filepath}
}

// ToGenericRouteNs converts route-type namespace to generic type namespace. Such an object can be used to call common
// namespace-related methods
func ToGenericRouteNs(ns *l3.LinuxStaticRoutes_Route_Namespace) *linuxcalls.Namespace {
	if ns == nil {
		return &linuxcalls.Namespace{}
	}
	return &linuxcalls.Namespace{Type: int32(ns.Type), Pid: ns.Pid, Microservice: ns.Microservice, Name: ns.Name, Filepath: ns.Filepath}
}

// ToRouteNs converts generic namespace to arp-type namespace
func ToRouteNs(ns *linuxcalls.Namespace) (*l3.LinuxStaticRoutes_Route_Namespace, error) {
	if ns == nil {
		return nil, fmt.Errorf("provided namespace is nil")
	}
	var namespaceType l3.LinuxStaticRoutes_Route_Namespace_NamespaceType
	switch ns.Type {
	case 0:
		namespaceType = l3.LinuxStaticRoutes_Route_Namespace_PID_REF_NS
	case 1:
		namespaceType = l3.LinuxStaticRoutes_Route_Namespace_MICROSERVICE_REF_NS
	case 2:
		namespaceType = l3.LinuxStaticRoutes_Route_Namespace_NAMED_NS
	case 3:
		namespaceType = l3.LinuxStaticRoutes_Route_Namespace_FILE_REF_NS
	}
	return &l3.LinuxStaticRoutes_Route_Namespace{Type: namespaceType, Pid: ns.Pid, Microservice: ns.Microservice, Name: ns.Name, Filepath: ns.Filepath}, nil
}

// ToArpNs converts generic namespace to arp-type namespace
func ToArpNs(ns *linuxcalls.Namespace) (*l3.LinuxStaticArpEntries_ArpEntry_Namespace, error) {
	if ns == nil {
		return nil, fmt.Errorf("provided namespace is nil")
	}
	var namespaceType l3.LinuxStaticArpEntries_ArpEntry_Namespace_NamespaceType
	switch ns.Type {
	case 0:
		namespaceType = l3.LinuxStaticArpEntries_ArpEntry_Namespace_PID_REF_NS
	case 1:
		namespaceType = l3.LinuxStaticArpEntries_ArpEntry_Namespace_MICROSERVICE_REF_NS
	case 2:
		namespaceType = l3.LinuxStaticArpEntries_ArpEntry_Namespace_NAMED_NS
	case 3:
		namespaceType = l3.LinuxStaticArpEntries_ArpEntry_Namespace_FILE_REF_NS
	}
	return &l3.LinuxStaticArpEntries_ArpEntry_Namespace{Type: namespaceType, Pid: ns.Pid, Microservice: ns.Microservice, Name: ns.Name, Filepath: ns.Filepath}, nil
}
