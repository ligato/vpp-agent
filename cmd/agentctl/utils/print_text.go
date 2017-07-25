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

package utils

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"bytes"
	"github.com/ligato/cn-infra/statuscheck/model/status"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/model/interfaces"
	"github.com/logrusorgru/aurora.git"
)

const perLevelSpaces = 3

type pfx struct {
	printAsTree    bool
	perLevelSpaces int
}

var (
	prefixer   pfx
	treeWriter = NewTreeWriter(1, "├─", "│ ", "└─")
)

func newPrefixer(t bool, s int) pfx {
	return pfx{t, s}
}

func (p pfx) getPrefix(level int) string {
	if p.printAsTree {
		return fmt.Sprintf("%d^@%s", level, strings.Repeat(" ", level*p.perLevelSpaces))
	}
	return strings.Repeat(" ", level*p.perLevelSpaces)
}

// PrintDataAsText prints data from an EtcdDump repo in text
// format.
func (ed EtcdDump) PrintDataAsText(showEtcd bool, printAsTree bool) *bytes.Buffer {
	prefixer = newPrefixer(printAsTree, perLevelSpaces)
	keys := ed.getSortedKeys()

	nameFuncMap := template.FuncMap{
		"setBold": setBold,
		"pfx":     getPrefix,
	}

	nameTemplate, err := template.New("name").Funcs(nameFuncMap).Parse("{{setBold .}}:\n")
	if err != nil {
		panic(err)
	}

	stsFuncMap := template.FuncMap{
		"convertTime": convertTime,
		"setBold":     setBold,
		"setRed":      setRed,
		"setOsColor":  setOsColor,
		"pfx":         getPrefix,
	}

	stsTemplate, err := template.New("status").Funcs(stsFuncMap).Parse(
		"{{pfx 1}}STATUS:" +
			"{{$etcd := .ShowEtcd}}" +
			"{{range $index, $element := .Status}}\n{{pfx 2}}{{$index}}: {{setOsColor .State}}" +
			"{{if .LastUpdate}}, Updated: {{convertTime .LastUpdate | setBold}}{{end}}" +
			"{{if .LastChange}}, Changed: {{convertTime .LastChange}}{{end}}" +
			"{{if .BuildVersion}}\n{{pfx 3}}    Version: '{{.BuildVersion}}'{{end}}" +
			"{{if .BuildDate}}, Built: {{setBold .BuildDate}}{{end}}" +
			"{{if $etcd}}\n{{pfx 3}}    ETCD: Rev {{.Rev}}, Key '{{.Key}}'{{end}}" +
			"{{else}} {{setRed \"INACTIVE\"}}{{end}}\n")
	if err != nil {
		panic(err)
	}

	ifFuncMap := template.FuncMap{
		"convertTime":    convertTime,
		"setBold":        setBold,
		"setGreen":       setGreen,
		"setRed":         setRed,
		"isEnabled":      isEnabled,
		"setStsColor":    setStsColor,
		"getIfTypeInfo":  getIfTypeInfo,
		"getIpAddresses": getIPAddresses,
		"pfx":            getPrefix,
	}

	ifTemplate, err := template.New("interfaces").Funcs(ifFuncMap).Parse(
		"{{$etcd := .ShowEtcd}}{{with .Interfaces}}\n{{pfx 1}}INTERFACES:" +
			"{{$paLbl := \"PhAddr: \"}}" +

			// Interface status (combines status values from config and state)
			"{{range $index, $element := .}}\n{{pfx 2}}{{setBold $index}}" +
			"{{with .State}} ({{.InternalName}}, ifIdx {{.IfIndex}}){{end}}:\n" +
			"{{pfx 3}}Status: <" +
			"{{if .Config}}{{isEnabled .Config.Enabled}}{{else}}{{setRed \"NOT-IN-CONFIG\"}}: {{end}}" +
			"{{with .State}}{{setStsColor \"ADMIN\" .AdminStatus}}, {{setStsColor \"OPER\" .OperStatus}} {{else}}, {{setRed \"NOT-IN-VPP\"}}{{end}}" +
			">" +

			// Interface type
			"{{with .Config}}\n{{pfx 3}}IfType: {{.Type}}{{getIfTypeInfo .}}{{end}}" +

			// Interface MTU
			//TODO "{{with .Config}}\n{{pfx 3}}MTU: {{.Mtu}}{{end}}" +

			// IP Address and attributes (if configured)
			"{{with .Config}}\n{{pfx 3}}IpAddr: {{getIpAddresses .IpAddresses}}{{end}}" +

			// Physical (MAC) Address from both config and state
			// (if configured or available from state)
			"{{if .Config}}{{if .Config.PhysAddress}}\n{{pfx 3}}{{$paLbl}}{{.Config.PhysAddress}}" +
			"{{if .State}}{{if .State.PhysAddress}}, (s {{.State.PhysAddress}}){{end}}{{end}}" +
			"{{else if .State}}{{if .State.PhysAddress}}\n{{pfx 3}}{{$paLbl}}(s {{.State.PhysAddress}}){{end}}{{end}}" +
			"{{else if .State}}{{if .State.PhysAddress}}\n{{pfx 3}}{{$paLbl}}(s {{.State.PhysAddress}}){{end}}{{end}}" +

			// Link attributes (if available from state)
			"{{with .State}}{{if or .Mtu .Speed .Duplex}}\n{{pfx 3}}LnkAtr: {{with .Mtu}}mtu {{.}}{{end}}" +
			"{{with .Speed}}, speed {{.}}{{end}}{{with .Duplex}}, duplex {{.}}{{end}}{{end}}{{end}}" +

			// Interface statistics (if available from State)
			"{{with .State}}{{with .Statistics}}" +
			"{{if or .InPackets .InBytes .OutPackets .OutBytes .Ipv4Packets .Ipv6Packets .DropPackets .InErrorPackets .InMissPackets .PuntPackets .InNobufPackets}}\n" +
			"{{pfx 3}}Stats:" +
			"\n{{pfx 4}}In: pkt {{.InPackets}}, byte {{.InBytes}}, errPkt {{.InErrorPackets}}, nobufPkt {{.InNobufPackets}}, missPkt {{.InMissPackets}}" +
			"\n{{pfx 4}}Out: pkt {{.OutPackets}}, byte {{.OutBytes}}, errPkt {{.OutErrorPackets}}" +
			"\n{{pfx 4}}Misc: drop {{.DropPackets}}, punt {{.PuntPackets}}, ipv4 {{.Ipv4Packets}}, ipv6 {{.Ipv6Packets}}" +
			"{{end}}{{end}}{{end}}" +

			// Etcd metadata for both the config and state records
			"{{if $etcd}}\n{{pfx 3}}ETCD:" +
			"{{with .Config}}\n{{pfx 4}}Cfg: Rev {{.Rev}}, Key '{{.Key}}'{{end}}" +
			"{{with .State}}\n{{pfx 4}}Sts: Rev {{.Rev}}, Key '{{.Key}}'{{end}}" +
			"{{end}}\n" +
			"{{end}}" +
			"{{end}}")
	if err != nil {
		panic(err)
	}

	bdFuncMap := template.FuncMap{
		"setBold": setBold,
		"pfx":     getPrefix,
	}

	bdTemplate, err := template.New("bridgeDomains").Funcs(bdFuncMap).Parse(
		"{{$etcd := .ShowEtcd}}" +
			"{{$fibTableEntries := .FibTableEntries}}" +
			"{{with .BridgeDomains}}\n{{pfx 1}}BRIDGE DOMAINS:" +
			"{{range $bdKey, $element := .}}\n{{pfx 2}}{{setBold $bdKey}}:\n{{pfx 3}}Attributes: macAge {{.MacAge}}" +

			// Bridge domain attributes
			"{{if or .Flood .UnknownUnicastFlood .Forward .Learn .ArpTermination}}, <{{if .Flood}}FLOOD{{end}}" +
			"{{if .UnknownUnicastFlood}}{{if .Flood}},{{end}} UNKN-UNICAST-FLOOD{{end}}" +
			"{{if .Forward}}{{if or .Flood .UnknownUnicastFlood}},{{end}} FORWARD{{end}}" +
			"{{if .Learn}}{{if or .Flood .UnknownUnicastFlood .Forward}},{{end}} LEARN{{end}}" +
			"{{if .ArpTermination}}{{if or .Flood .UnknownUnicastFlood .Forward .Learn}},{{end}} ARP-TERMINATION{{end}}>" +
			"{{end}}" +

			// Interface table
			"{{with .Interfaces}}\n{{pfx 3}}Interfaces:" +
			"{{range $ifKey, $element := .}}\n{{pfx 4}}{{setBold $element.Name}} splitHorizonGrp {{.SplitHorizonGroup}}" +
			"{{if .BridgedVirtualInterface}}, <BVI>{{end}}" +
			"{{end}}" +
			"{{end}}" +

			// ARP termination table
			"{{with .ArpTerminationTable}}\n{{pfx 3}}ARP-Table:" +
			"{{range $arpKey, $arp := .}}\n{{pfx 4}}{{$arp.IpAddress}}: {{$arp.PhysAddress}}{{end}}" +
			"{{end}}" +

			// Etcd metadata
			"{{if $etcd}}\n{{pfx 3}}ETCD: Rev {{.Rev}}, Key '{{.Key}}'{{end}}\n" +
			"{{end}}" +
			// FIB table
			"{{with $fibTableEntries}}\n" +
			"{{with .FibTable}}" +
			"{{pfx 2}}FIB-Table:" +
			"{{range $fibKey, $fib := .}}\n" +
			"{{pfx 3}}{{$fib.PhysAddress}}" +
			"{{with $fib.OutgoingInterface}}, {{$fib.OutgoingInterface}}{{end}}" +
			"{{with $fib.BridgeDomain}}, {{$fib.BridgeDomain}}{{end}}" +
			"{{if $fib.StaticConfig}}, <STATIC>{{end}}" +
			"{{if $fib.BridgedVirtualInterface}}, <BVI>{{end}}" +
			"{{if eq $fib.Action 0}}, <FORWARD> {{else}}, <DROP>{{end}}" +
			"{{end}}" +
			"{{end}}" +
			"{{end}}" +
			"{{end}}\n\n")

	buffer := new(bytes.Buffer)
	if printAsTree {
		writer := treeWriter
		for _, key := range keys {
			vd, _ := ed[key]
			vd.ShowEtcd = showEtcd

			for _, bd := range vd.BridgeDomains {
				nl := []*string{}
				for _, bdi := range bd.Interfaces {
					nl = append(nl, &bdi.Name)
				}
				padRight(nl, ":")
			}
			nameTemplate.Execute(os.Stdout, key)
			stsTemplate.Execute(writer, vd)
			ifTemplate.Execute(writer, vd)
			bdTemplate.Execute(writer, vd)
			treeWriter.FlushTree()
			fmt.Println("")
		}
	} else {
		buffer.WriteTo(os.Stdout)
		for _, key := range keys {
			vd, _ := ed[key]
			vd.ShowEtcd = showEtcd
			for _, bd := range vd.BridgeDomains {
				nl := []*string{}
				for _, bdi := range bd.Interfaces {
					nl = append(nl, &bdi.Name)
				}
				padRight(nl, ":")
			}
			nameTemplate.Execute(buffer, key)
			stsTemplate.Execute(buffer, vd)
			ifTemplate.Execute(buffer, vd)
			bdTemplate.Execute(buffer, vd)
		}
	}
	return buffer
}

func getPrefix(level int) string {
	return prefixer.getPrefix(level)
}

func isEnabled(enabled bool) string {
	if enabled {
		return fmt.Sprintf("%s", aurora.Green("ENABLED"))
	}
	return fmt.Sprintf("%s", aurora.Red("DISABLED"))
}

func convertTime(t int64) string {
	return time.Unix(t, 0).Format("2006-01-02 15:04:05")
}
func setRed(attr interface{}) string {
	return fmt.Sprintf("%s", aurora.Red(attr))
}
func setGreen(attr interface{}) string {
	return fmt.Sprintf("%s", aurora.Green(attr))
}
func setYellow(attr interface{}) string {
	return fmt.Sprintf("%s", aurora.Brown(attr))
}
func setBold(attr interface{}) string {
	return fmt.Sprintf("%s", aurora.Bold(attr))
}

// setOsColor sets the color for the Operational State
func setOsColor(arg status.OperationalState) string {
	switch arg {
	case status.OperationalState_OK:
		return setGreen(arg)
	case status.OperationalState_INIT:
		return setYellow(arg)
	case status.OperationalState_ERROR:
		return setRed(arg)
	default:
		return arg.String()
	}
}

func setStsColor(kind string, arg interfaces.InterfacesState_Interface_Status) string {
	sts := fmt.Sprintf("%s-%s", kind, arg)
	switch arg {
	case interfaces.InterfacesState_Interface_UP:
		return setGreen(sts)
	case interfaces.InterfacesState_Interface_DOWN:
		return setRed(sts)
	default:
		return sts
	}
}

// getIfTypeInfo gets type-specific parameters for an interface.
// The parameters are returned as a formatted string ready to be
// printed out.
func getIfTypeInfo(ifc *IfconfigWithMD) string {
	switch ifc.Type {
	case interfaces.InterfaceType_MEMORY_INTERFACE:
		if ifc.Memif.Master {
			return fmt.Sprintf("; <MASTER>, id %d, bufSize %d, rngSize %d, socketFN '%s', secret '%s', rxQueues '%d', txQueues '%d'",
				ifc.Memif.Id, ifc.Memif.BufferSize, ifc.Memif.RingSize, ifc.Memif.SocketFilename, ifc.Memif.Secret,
				ifc.Memif.RxQueues, ifc.Memif.TxQueues)
		}
		return fmt.Sprintf("; id %d, bufSize %d, rngSize %d, socketFN '%s', secret '%s', rxQueues '%d', txQueues '%d'",
			ifc.Memif.Id, ifc.Memif.BufferSize, ifc.Memif.RingSize, ifc.Memif.SocketFilename, ifc.Memif.Secret,
			ifc.Memif.RxQueues, ifc.Memif.TxQueues)

	case interfaces.InterfaceType_VXLAN_TUNNEL:
		return fmt.Sprintf("; srcIp %s, dstIp %s, vni %d",
			ifc.Vxlan.SrcAddress, ifc.Vxlan.DstAddress, ifc.Vxlan.Vni)
	case interfaces.InterfaceType_AF_PACKET_INTERFACE:
		return fmt.Sprintf("; hostName %s", ifc.Afpacket.HostIfName)
	default:
		return ""
	}
}

// getIPAddresses gets a list of IPv4 addresses configured on an
// interface. The parameters are returned as a formatted string
// ready to be printed out.
func getIPAddresses(addrs []string) string {
	return strings.Join(addrs, ", ")
}
