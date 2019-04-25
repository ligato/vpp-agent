package utils

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/ligato/vpp-agent/api/models/linux/interfaces"
	"github.com/ligato/vpp-agent/api/models/linux/l3"

	vpp_ipsec "github.com/ligato/vpp-agent/api/models/vpp/ipsec"

	"github.com/ligato/vpp-agent/api/models/vpp/acl"
	"github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/api/models/vpp/l2"
	"github.com/ligato/vpp-agent/api/models/vpp/l3"
	"github.com/ligato/vpp-agent/api/models/vpp/nat"

	"github.com/ligato/cn-infra/health/statuscheck/model/status"
	"github.com/logrusorgru/aurora.git"

	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
)

const perLevelSpaces = 3

type pfx struct {
	printAsTree    bool
	perLevelSpaces int
}

var (
	prefixer pfx
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

func (ed EtcdDump) PrintStatus(showConf bool) (*bytes.Buffer, error) {
	prefixer = newPrefixer(false, perLevelSpaces)

	stsFuncMap := template.FuncMap{
		"convertTime": convertTime,
		"setBold":     setBold,
		"setRed":      setRed,
		"setOsColor":  setOsColor,
		"pfx":         getPrefix,
	}

	stsTemplate := template.Must(template.New("status").Funcs(stsFuncMap).Parse(
		"{{pfx 1}}{{setBold \"STATUS\"}}:" +
			"{{$etcd := .ShowEtcd}}" +
			// Iterate over status.
			"{{range $statusName, $statusData := .Status}}\n{{pfx 2}}{{$statusName}}: {{setOsColor .State}}" +
			"{{if .LastUpdate}}, Updated: {{convertTime .LastUpdate | setBold}}{{end}}" +
			"{{if .LastChange}}, Changed: {{convertTime .LastChange}}{{end}}" +
			"{{if .BuildVersion}}\n{{pfx 3}}    Version: '{{.BuildVersion}}'{{end}}" +
			"{{if .BuildDate}}, Built: {{setBold .BuildDate}}{{end}}" +
			"{{if $etcd}}\n{{pfx 3}}    ETCD: Rev {{.Rev}}, Key '{{.Key}}'{{end}}" +
			// In case there is no status
			"{{else}} {{setRed \"INACTIVE\"}}" +
			// Iterate over status - end of loop
			"{{end}}\n"))

	templates := []*template.Template{}
	templates = append(templates, stsTemplate)

	return ed.textRenderer(showConf, templates)
}

func (ed EtcdDump) PrintConfig(showConf bool) (*bytes.Buffer, error) {
	prefixer = newPrefixer(false, perLevelSpaces)

	ifTemplate := createInterfaceTemplate()
	aclTemplate := createACLTemplate()
	bdTemplate := createBridgeTemplate()
	fibTemplate := createFibTableTemplate()
	xconnectTemplate := createXconnectTableTemplate()
	arpTemplate := createArpTableTemplate()
	routeTemplate := createRouteTableTemplate()
	proxyarpTemplate := createProxyArpTemplate()
	ipneighbor := createIPScanNeightTemplate()
	//nat := createNATTemplate()
	//dnat := createDNATTemplate()
	spolicy := createIPSecPolicyTemplate()
	sassociation := createIPSecAssociationTemplate()

	linterface := createlInterfaceTemplate()
	larp := createlARPTemplate()
	lroute := createlRouteTemplate()

	templates := []*template.Template{}
	// Keep template order
	templates = append(templates, ifTemplate,
		aclTemplate, bdTemplate, fibTemplate, xconnectTemplate,
		arpTemplate, routeTemplate, proxyarpTemplate, ipneighbor,
		/*nat, dnat,*/ spolicy, sassociation, linterface, larp, lroute)

	return ed.textRenderer(showConf, templates)
}

func createACLTemplate() *template.Template {

	FuncMap := template.FuncMap{
		"setBold":        setBold,
		"isEnabled":      isEnabled,
		"getIpAddresses": getIPAddresses,
		"pfx":            getPrefix,
	}

	Template := template.Must(template.New("acl").Funcs(FuncMap).Parse(
		"{{$conf := .ShowConf}}{{$print := .PrintConf}}" +
			"{{with .Config}}{{with .VppConfig}}" +
			"{{with .Acls}}\n{{pfx 1}}{{setBold \"ACL\"}}" +

			"{{if $print}}:" +

			// Iterate over ACL.
			"{{range .}}\n{{pfx 2}}{{.Name}}" +
			"{{if $conf}}" +

			//// Iterate over ACL rule.
			"{{range .Rules}}" +
			"\n{{pfx 4}}Action: {{.Action}}" +

			//Start IP Rule
			"{{with .IpRule}}\n{{pfx 4}}IP Rule:" +

			//Start IP
			"{{with .Ip}}\n{{pfx 5}}IP:" +
			"\n{{pfx 6}}Destination Network: {{.DestinationNetwork}}" +
			"\n{{pfx 6}}Source Network: {{.SourceNetwork}}" +
			// End IP
			"{{end}}" +

			"{{with .Icmp}}\n{{pfx 5}}ICMP:" +
			"\n{{pfx 6}}Is ICMPv6{{isEnabled .Icmpv6}}" +

			//Start ICMP Code Range
			"{{with .IcmpCodeRange}}\n{{pfx 6}}ICMP Code Range:" +
			"\n{{pfx 7}}First: {{.First}}" +
			"\n{{pfx 7}}Last: {{.Last}}" +
			//End ICMP Code Range
			"{{end}}" +

			//Start ICMP Type Range
			"{{with .IcmpTypeRange}}\n{{pfx 6}}ICMP Type Range" +
			"\n{{pfx 7}}First: {{.First}}" +
			"\n{{pfx 7}}Last: {{.Last}}" +
			//End ICMP Type Range
			"{{end}}" +

			//End ICMP
			"{{end}}" +

			//Start TCP
			"{{with .Tcp}}\n{{pfx 5}}TCP:" +

			"{{with .DestinationPortRange}}\n{{pfx 6}}Destination Port Range:" +
			"\n{{pfx 7}}Lower Port: {{.LowerPort}}" +
			"\n{{pfx 7}}Upper Port: {{.UpperPort}}" +
			//End DestinationPortRange
			"{{end}}" +

			"{{with .SourcePortRange}}\n{{pfx 6}}Source Port Range" +
			"\n{{pfx 7}}Lower Port: {{.LowerPort}}" +
			"\n{{pfx 7}}Upper Port: {{.UpperPort}}" +

			//End SourcePortRange
			"{{end}}" +

			"\n{{pfx 6}}TCP Flags Mask: {{.TcpFlagsMask}}" +
			"\n{{pfx 6}}TCP Flags Value: {{.TcpFlagsValue}}" +

			//End TCP
			"{{end}}" +

			//Start UDP
			"{{with .Udp}}\n{{pfx 5}}UDP:" +
			//Start UDP Destinaton Port
			"{{with .DestinationPortRange}}\n{{pfx 6}}Destination Port Range:" +
			"\n{{pfx 7}}Lower Port: {{.LowerPort}}" +
			"\n{{pfx 7}}Upper Port: {{.UpperPort}}" +

			//End DestinationPortRange
			"{{end}}" +

			//Start UDP Source Port
			"{{with .SourcePortRange}}\n{{pfx 6}}Source Port Range:" +
			"\n{{pfx 7}}Lower Port: {{.LowerPort}}" +
			"\n{{pfx 7}}Upper Port: {{.UpperPort}}" +

			//End SourcePortRange
			"{{end}}" +

			//End UDP
			"{{end}}" +

			//End IP Rule
			"{{end}}" +

			//Macip Rule
			"{{with .MacipRule}}\n{{pfx 4}}Macip Rule:" +
			"\n{{pfx 5}}Source Address: {{.SourceAddress}}/{{.SourceAddressPrefix}}" +

			"\n{{pfx 5}}Source MacAddress: {{.SourceMacAddress}}/{{.SourceMacAddressMAsk}}" +

			//End Macip Rule
			"{{end}}" +

			//End Rule range
			"{{end}}" +

			//Interfaces
			"{{with .Interfaces}}\n{{pfx 3}} Interfaces:" +
			"\n{{pfx 4}}Egress:" +
			"{{range $Id, $Value := .Egress}}\n{{pfx 5}}{{$Value}}{{end}}" +
			"\n{{pfx 4}}Ingress:" +
			"{{range $Id, $Value := .Ingress}}\n{{pfx 5}}{{$Value}}{{end}}" +
			"{{end}}" +

			// End Config
			"{{end}}" +

			// End Iterate
			"{{end}}" +

			// End print
			"\n{{end}}" +

			"{{end}}" +
			"{{end}}{{end}}"))

	return Template
}

func createInterfaceTemplate() *template.Template {

	ifFuncMap := template.FuncMap{
		"convertTime":    convertTime,
		"setBold":        setBold,
		"setGreen":       setGreen,
		"setRed":         setRed,
		"isEnabled":      isEnabled,
		"setStsColor":    setStsColor,
		"getIpAddresses": getIPAddresses,
		"pfx":            getPrefix,
	}

	Template := template.Must(template.New("interfaces").Funcs(ifFuncMap).Parse(
		"{{$conf := .ShowConf}}{{$print := .PrintConf}}" +
			"{{with .Config}}{{with .VppConfig}}" +
			"{{with .Interfaces}}\n{{pfx 1}}{{setBold \"Interfaces\"}}" +

			"{{if $print}}:" +

			// Iterate over interfaces.
			"{{range .}}" +
			"\n{{pfx 2}}{{.Name}}" +

			"{{if $conf}}" +
			// Interface overall status
			" {{isEnabled .Enabled}} " +

			// Interface type
			"\n{{pfx 3}}IfType: {{.Type}}" +

			"\n{{pfx 3}}PhyAddr: {{.PhysAddress}}" +

			// IP Address and attributes (if configured)
			"{{with .IpAddresses}}\n{{pfx 3}}IpAddr: " +
			"{{getIpAddresses .}}" +
			"{{end}}" +

			"\n{{pfx 3}}Mtu: {{.Mtu}}" +
			"\n{{pfx 3}}Vrf: {{.Vrf}}" +
			"\n{{pfx 3}}Dhcp Client: {{isEnabled .SetDhcpClient}}" +

			// End Config
			"{{end}}" +

			// End print
			"\n{{end}}" +

			"{{end}}{{end}}" +
			"{{end}}{{end}}"))

	return Template
}

func createBridgeTemplate() *template.Template {

	FuncMap := template.FuncMap{
		"setBold":        setBold,
		"isEnabled":      isEnabled,
		"getIpAddresses": getIPAddresses,
		"pfx":            getPrefix,
	}

	Template := template.Must(template.New("bridgeDomains").Funcs(FuncMap).Parse(
		"{{$conf := .ShowConf}}{{$print := .PrintConf}}" +
			"{{with .Config}}{{with .VppConfig}}" +
			"{{with .BridgeDomains}}\n{{pfx 1}}{{setBold \"Bridge\"}}" +

			"{{if $print}}:" +

			"{{range .}}" +
			"\n{{pfx 2}}{{.Name}}" +
			"{{if $conf}}" +
			"\n{{pfx 3}}Flood: {{isEnabled .Flood}}" +
			"\n{{pfx 3}}Unknown Unicast Flood: {{isEnabled .UnknownUnicastFlood}}" +
			"\n{{pfx 3}}Forward: {{isEnabled .Forward}}" +
			"\n{{pfx 3}}Learn: {{isEnabled .Learn}}" +
			"\n{{pfx 3}}ArpTermination: {{isEnabled .ArpTermination}}" +
			"\n{{pfx 3}}: {{.MacAge}}" +

			//Interate over interfaces.
			"\n{{pfx 3}}Interfaces:" +
			"{{range .Interfaces}}" +
			"\n{{pfx 4}}Name: {{.Name}}" +
			"\n{{pfx 4}}BridgedVirtualInterface: {{isEnabled .BridgedVirtualInterface}}" +
			"\n{{pfx 4}}SplitHorizonGroup: {{.SplitHorizonGroup}}" +

			//End interate interfaces.
			"{{end}}" +

			//Interate over ArpTerminationTable.
			"\n{{pfx 3}}Arp Termination Table:" +
			"{{range .ArpTerminationTable}}" +
			"\n{{pfx 4}}IP Address: {{.IpAddress}}" +
			"\n{{pfx 4}}MAC: {{.PhysAddress}}" +

			//End interate ArpTerminationTable.
			"{{end}}" +

			//End Config
			"{{end}}" +

			// End print
			"\n{{end}}" +

			"{{end}}{{end}}" +
			"{{end}}{{end}}"))

	return Template
}

func createFibTableTemplate() *template.Template {

	FuncMap := template.FuncMap{
		"convertTime":    convertTime,
		"setBold":        setBold,
		"setGreen":       setGreen,
		"setRed":         setRed,
		"isEnabled":      isEnabled,
		"setStsColor":    setStsColor,
		"getIpAddresses": getIPAddresses,
		"pfx":            getPrefix,
	}

	Template := template.Must(template.New("fibTable").Funcs(FuncMap).Parse(
		"{{$print := .PrintConf}}" +
			"{{with .Config}}{{with .VppConfig}}" +
			"{{with .Fibs}}\n{{pfx 1}}{{setBold \"Fib Table\"}}" +

			"{{if $print}}:" +

			"{{range .}}" +

			"\n{{pfx 3}}MAC address: {{.PhysAddress}}" +
			"\n{{pfx 3}}Bridge Domain: {{.BridgeDomain}}" +
			"\n{{pfx 3}}Outgoing Interface: {{.OutgoingInterface}}" +
			"\n{{pfx 3}}Action: {{.Action}}" +
			"\n{{pfx 3}}Static Config: {{isEnabled .StaticConfig}}" +
			"\n{{pfx 3}}Bridge Virtual Interface: {{isEnabled .BridgedVirtualInterface}}" +

			"{{end}}" +

			// End print
			"\n{{end}}" +

			"{{end}}{{end}}" +
			"{{end}}"))

	return Template
}

func createXconnectTableTemplate() *template.Template {

	FuncMap := template.FuncMap{
		"setBold": setBold,
		"pfx":     getPrefix,
	}

	Template := template.Must(template.New("xconnect").Funcs(FuncMap).Parse(
		"{{$conf := .ShowConf}}{{$print := .PrintConf}}" +
			"{{with .Config}}{{with .VppConfig}}" +
			"{{with .XconnectPairs}}\n{{pfx 1}}{{setBold \"Xconnect\"}}" +

			"{{if $print}}:" +

			// Iterate over xconnect.
			"{{range .XConnectPairs}}" +
			"{{if $conf}}" +

			"{{with .Config}}{{with .Xconnect}}" +

			"\n{{pfx 3}}Receive Interface: {{.ReceiveInterface}}" +
			"\n{{pfx 3}}Transmit Interface: {{.TransmitInterface}}" +

			//End Xconnect
			"{{end}}{{end}}" +
			//End Conf
			"{{end}}" +

			// End print
			"\n{{end}}" +

			"{{end}}{{end}}" +
			"{{end}}{{end}}"))

	return Template
}

func createArpTableTemplate() *template.Template {

	FuncMap := template.FuncMap{
		"setBold":        setBold,
		"getIpAddresses": getIPAddresses,
		"isEnabled":      isEnabled,
		"pfx":            getPrefix,
	}

	Template := template.Must(template.New("arp").Funcs(FuncMap).Parse(
		"{{$print := .PrintConf}}" +
			"{{with .Config}}{{with .VppConfig}}" +
			"{{with .Arps}}\n{{pfx 1}}{{setBold \"ARP\"}}" +

			"{{if $print}}:" +

			"{{range .}}" +

			"\n{{pfx 2}}Interface: {{.Interface}}" +
			"\n{{pfx 2}}IP address: {{.IpAddress}}" +
			"\n{{pfx 2}}MAC: {{.PhysAddress}}" +
			"\n{{pfx 2}}Static: {{isEnabled .Static}}" +

			"{{end}}" +

			// End print
			"\n{{end}}" +

			//End
			"{{end}}" +
			"{{end}}{{end}}"))

	return Template
}

func createRouteTableTemplate() *template.Template {

	FuncMap := template.FuncMap{
		"setBold":        setBold,
		"getIpAddresses": getIPAddresses,
		"pfx":            getPrefix,
	}

	Template := template.Must(template.New("routetable").Funcs(FuncMap).Parse(
		"{{$print := .PrintConf}}" +
			"{{with .Config}}{{with .VppConfig}}" +
			"{{with .Routes}}\n{{pfx 1}}{{setBold \"Route Table\"}}:" +

			"{{if $print}}:" +

			"{{range .}}" +

			"\n{{pfx 2}}Type: {{.Type}}" +
			"\n{{pfx 2}}VrfId: {{.VrfId}}" +
			"\n{{pfx 2}}Destination Address: {{.DstNetwork}}" +
			"\n{{pfx 2}}Next Hop Address: {{.NextHopAddr}}" +
			"\n{{pfx 2}}Out going Interface: {{.OutgoingInterface}}" +
			"\n{{pfx 2}}Weight: {{.Weight}}" +
			"\n{{pfx 2}}Preference: {{.Preference}}" +
			"\n{{pfx 2}}ViaVrfId: {{.ViaVrfId}}" +

			//End
			"{{end}}" +

			// End print
			"\n{{end}}" +

			"{{end}}" +

			"{{end}}{{end}}"))

	return Template
}

func createProxyArpTemplate() *template.Template {

	FuncMap := template.FuncMap{
		"setBold": setBold,
		"pfx":     getPrefix,
	}

	Template := template.Must(template.New("proxyarp").Funcs(FuncMap).Parse(
		"{{$print := .PrintConf}}" +
			"{{with .Config}}{{with .VppConfig}}" +
			"{{with .ProxyArp}}\n{{pfx 1}}{{setBold \"Proxy ARP\"}}" +

			"{{if $print}}:" +

			//Iterate over Interfaces
			"{{range .Interfaces}}" +
			"\n{{pfx 2}}{{.Name}}" +

			//End iterate
			"{{end}}" +

			//Iterate over Proxy Ranges
			"\n{{pfx 2}}Range:" +
			"{{range .Ranges}}" +
			"\n{{pfx 3}}{{.FirstIpAddr}} - {{.LastIpAddr}}" +

			//End iterate
			"{{end}}" +

			// End print
			"\n{{end}}" +

			//End
			"{{end}}" +

			"{{end}}{{end}}"))

	return Template
}

func createIPScanNeightTemplate() *template.Template {

	FuncMap := template.FuncMap{
		"setBold": setBold,
		"pfx":     getPrefix,
	}

	Template := template.Must(template.New("ipscanneigh").Funcs(FuncMap).Parse(
		"{{$print := .PrintConf}}" +
			"{{with .Config}}{{with .VppConfig}}" +
			"{{with .IpscanNeighbor}}\n{{pfx 1}}{{setBold \"IP Neighbor\"}}" +

			"{{if $print}}:" +

			"\n{{pfx 2}}Mode: {{.Mode}}" +
			"\n{{pfx 2}}Scan Interval: {{.ScanInterval}}" +
			"\n{{pfx 2}}Max Proc Time: {{.MaxProcTime}}" +
			"\n{{pfx 2}}Max Proc Time: {{.MaxProcTime}}" +
			"\n{{pfx 2}}Max Uptime: {{.MaxUpdate}}" +
			"\n{{pfx 2}}Scan Int Delay: {{.ScanIntDelay}}" +
			"\n{{pfx 2}}Stale Threshold: {{.StaleThreshold}}" +

			// End print
			"\n{{end}}" +

			//End
			"{{end}}" +

			"{{end}}{{end}}"))

	return Template
}

func createNATTemplate() *template.Template {

	FuncMap := template.FuncMap{
		"setBold": setBold,
		"pfx":     getPrefix,
	}

	Template := template.Must(template.New("nat").Funcs(FuncMap).Parse(
		"{{$print := .PrintConf}}" +
			"{{with .Config}}{{with .VppConfig}}" +
			"{{with .NAT}}\n{{pfx 1}}{{setBold \"NAT\"}}" +

			"{{if $print}}:" +

			"{{with .Label}}\n{{pfx 2}}Label: {{.}}{{end}}" +

			//Iterate over StMappings
			"{{range $StMapName, $StMapData:= .StMappings}}" +
			"{{with .ExternalInterface}}\n{{pfx 3}}External Interface: {{.}}{{end}}" +
			"{{with .ExternalIp}}\n{{pfx 3}}External IP: {{.}}{{end}}" +
			"{{with .ExternalPort}}\n{{pfx 3}}External Port: {{.}}{{end}}" +

			//Iterate over Local IPs
			"{{range $LocalIPName, $LocalIPData := .LocalIps}}" +
			"{{with .VrfId}}\n{{pfx 4}}VrfID: {{.}}{{end}}" +
			"{{with .LocalIp}}\n{{pfx 4}}Local IP: {{.}}{{end}}" +
			"{{with .LocalPort}}\n{{pfx 4}}Local Port: {{.}}{{end}}" +
			"{{with .Probability}}\n{{pfx 4}}Probability: {{.}}{{end}}" +

			//End over Local IPs
			"{{end}}" +

			"{{with .Protocol}}\n{{pfx 3}}Protocol: {{.}}{{end}}" +
			"{{with .TwiceNat}}\n{{pfx 3}}Twice Nat: {{.}}{{end}}" +
			"{{with .SessionAffinity}}\n{{pfx 3}}Session Affinity: {{.}}{{end}}" +

			//End StMappings
			"{{end}}" +

			// End print
			"\n{{end}}" +

			//End
			"{{end}}" +

			"{{end}}{{end}}"))

	return Template
}

func createDNATTemplate() *template.Template {

	FuncMap := template.FuncMap{
		"setBold": setBold,
		"pfx":     getPrefix,
	}

	Template := template.Must(template.New("dnat").Funcs(FuncMap).Parse(
		"{{$print := .PrintConf}}" +
			"{{with .Config}}{{with .VppConfig}}" +
			"{{with .DNAT}}\n{{pfx 1}}{{setBold \"DNAT\"}}" +

			"{{if $print}}:" +

			"{{with .Label}}\n{{pfx 2}}Label: {{.}}{{end}}" +

			//Iterate over StMappings
			"{{range $StMapName, $StMapData:= .StMappings}}" +
			"{{with .ExternalInterface}}\n{{pfx 3}}External Interface: {{.}}{{end}}" +
			"{{with .ExternalIp}}\n{{pfx 3}}External IP: {{.}}{{end}}" +
			"{{with .ExternalPort}}\n{{pfx 3}}External Port: {{.}}{{end}}" +

			//Iterate over Local IPs
			"{{range $LocalIPName, $LocalIPData := .LocalIps}}" +
			"{{with .VrfId}}\n{{pfx 4}}VrfID: {{.}}{{end}}" +
			"{{with .LocalIp}}\n{{pfx 4}}Local IP: {{.}}{{end}}" +
			"{{with .LocalPort}}\n{{pfx 4}}Local Port: {{.}}{{end}}" +
			"{{with .Probability}}\n{{pfx 4}}Probability: {{.}}{{end}}" +

			//End over Local IPs
			"{{end}}" +

			"{{with .Protocol}}\n{{pfx 3}}Protocol: {{.}}{{end}}" +
			"{{with .TwiceNat}}\n{{pfx 3}}Twice Nat: {{.}}{{end}}" +
			"{{with .SessionAffinity}}\n{{pfx 3}}Session Affinity: {{.}}{{end}}" +

			//End StMappings
			"{{end}}" +

			//Iterate over StMappings
			"{{range $IdMapName, $IDMapData:= .IdMappings}}" +
			"{{with .VrfId}}\n{{pfx 4}}VrfID: {{.}}{{end}}" +
			"{{with .Interface}}\n{{pfx 4}}Interface: {{.}}{{end}}" +
			"{{with .IPAddress}}\n{{pfx 4}}IP Address: {{.}}{{end}}" +
			"{{with .Port}}\n{{pfx 4}}Port: {{.}}{{end}}" +
			"{{with .Protocol}}\n{{pfx 4}}Protocol: {{.}}{{end}}" +
			//End over StMappings
			"{{end}}" +

			// End print
			"\n{{end}}" +

			//End
			"{{end}}" +

			"{{end}}{{end}}"))

	return Template
}

func createIPSecPolicyTemplate() *template.Template {

	FuncMap := template.FuncMap{
		"setBold": setBold,
		"pfx":     getPrefix,
	}

	Template := template.Must(template.New("ipsecpolicy").Funcs(FuncMap).Parse(
		"{{$print := .PrintConf}}" +
			"{{with .Config}}{{with .VppConfig}}" +
			"{{with .IpsecSpds}}\n{{pfx 1}}{{setBold \"Security policy database\"}}" +

			"{{if $print}}:" +

			// Iterate over Policy.
			"{{range .}}" +

			"\n{{pfx 2}}Index: {{.Index}}" +

			// Iterate over Interfaces.
			"\n{{pfx 2}}Interfaces:" +
			"{{range .Interfaces}}" +
			"\n{{pfx 3}}Name: {{.Name}}" +

			// End iterate over Interfaces.
			"{{end}}" +

			//Iterate over Interfaces.
			"\n{{pfx 2}}PolicyEntries:" +
			"{{range .PolicyEntries}}" +
			"\n{{pfx 3}}SaIndex: {{.SaIndex}}" +
			"\n{{pfx 3}}Priority: {{.Priority}}" +
			"\n{{pfx 3}}Is Outbound: {{.IsOutbound}}" +
			"\n{{pfx 3}}Remote Addr Start: {{.RemoteAddrStart}}" +
			"\n{{pfx 3}}Remote Addr Stop: {{.RemoteAddrStop}}" +
			"\n{{pfx 3}}Local Addr Start: {{.LocalAddrStart}}" +
			"\n{{pfx 3}}Local Addr Stop: {{.LocalAddrStop}}" +
			"\n{{pfx 3}}Protocol: {{.Protocol}}" +
			"\n{{pfx 3}}Remote Port Start: {{.RemotePortStart}}" +
			"\n{{pfx 3}}Remote Port Stop: {{.RemotePortStop}}" +
			"\n{{pfx 3}}Local Port Start: {{.LocalPortStart}}" +
			"\n{{pfx 3}}Local Port Stop: {{.LocalPortStop}}" +
			"\n{{pfx 3}}Action: {{.Action}}" +

			// End iterate over Interface.
			"{{end}}" +

			// End Iterate over Policy.
			"{{end}}" +

			// End print
			"\n{{end}}" +

			"{{end}}" +

			"{{end}}{{end}}"))

	return Template
}

func createIPSecAssociationTemplate() *template.Template {

	FuncMap := template.FuncMap{
		"setBold": setBold,
		"pfx":     getPrefix,
	}

	Template := template.Must(template.New("ipsecassociation").Funcs(FuncMap).Parse(
		"{{$print := .PrintConf}}" +
			"{{with .Config}}{{with .VppConfig}}" +
			"{{with .IpsecSas}}\n{{pfx 1}}{{setBold \"Security associations\"}}" +

			"{{if $print}}:" +

			// Iterate over Association.
			"{{range .}}" +

			"\n{{pfx 3}}Index: {{.Index}}" +
			"\n{{pfx 3}}Spi: {{.Spi}}" +
			"\n{{pfx 3}}Protocol: {{.Protocol}}" +
			"\n{{pfx 3}}Crypto Alg: {{.CryptoAlg}}" +
			"\n{{pfx 3}}Crypto Key: {{.CryptoKey}}" +
			"\n{{pfx 3}}Integ Alg: {{.IntegAlg}}" +
			"\n{{pfx 3}}Integ Key: {{.IntegKey}}" +
			"\n{{pfx 3}}Use Esn: {{.UseEsn}}" +
			"\n{{pfx 3}}Use Anti Replay: {{.UseAntiReplay}}" +
			"\n{{pfx 3}}Tunnel Src Addr: {{.TunnelSrcAddr}}" +
			"\n{{pfx 3}}Tunnel Dst Addr: {{.TunnelDstAddr}}" +
			"\n{{pfx 3}}Enable Udp Encap: {{.EnableUdpEncap}}" +

			// End iterate over Association.
			"{{end}}" +

			// End print
			"\n{{end}}" +

			"{{end}}" +

			"{{end}}{{end}}"))

	return Template
}

func createlInterfaceTemplate() *template.Template {

	FuncMap := template.FuncMap{
		"setBold":        setBold,
		"getIpAddresses": getIPAddresses,
		"pfx":            getPrefix,
	}

	Template := template.Must(template.New("linterface").Funcs(FuncMap).Parse(
		"{{$conf := .ShowConf}}{{$print := .PrintConf}}" +
			"{{with .Config}}{{with .LinuxConfig}}" +
			"{{with .Interfaces}}\n{{pfx 1}}{{setBold \"Linux interface\"}}" +

			"{{if $print}}:" +

			// Iterate over interface.
			"{{range .}}" +
			"\n{{pfx 2}} Linux Interface" +
			"{{if $conf}}" +

			"\n{{pfx 3}}Name: {{.Name}}" +

			"\n{{pfx 3}}Type: {{.Type}}" +

			"{{with .Namespace}}\n{{pfx 3}}Name Space:" +
			"\n{{pfx 4}}Type: {{.Type}}" +
			"\n{{pfx 4}}Reference: {{.Reference}}" +
			"{{end}}" +

			"\n{{pfx 3}}Host IfName: {{.HostIfName}}" +
			"\n{{pfx 3}}Enabled: {{.Enabled}}" +
			"\n{{pfx 3}}Ip Addresses: {{getIpAddresses .IpAddresses}}" +
			"\n{{pfx 3}}PhysAddress: {{.PhysAddress}}" +
			"\n{{pfx 3}}Mtu: {{.Mtu}}" +
			"\n{{pfx 3}}Link: {{.Link}}" +

			//End
			"{{end}}" +

			//End iterate over interface.
			"{{end}}" +

			// End print
			"\n{{end}}" +

			"{{end}}" +

			"{{end}}{{end}}"))

	return Template
}

func createlARPTemplate() *template.Template {

	FuncMap := template.FuncMap{
		"setBold": setBold,
		"pfx":     getPrefix,
	}

	Template := template.Must(template.New("larp").Funcs(FuncMap).Parse(
		"{{$conf := .ShowConf}}{{$print := .PrintConf}}" +
			"{{with .Config}}{{with .LinuxConfig}}" +
			"{{with .ArpEntries}}\n{{pfx 1}}{{setBold \"Linux ARP\"}}" +

			"{{if $print}}:" +

			// Iterate over interface.
			"{{range .}}\n{{pfx 2}}" +

			"\n{{pfx 3}}Interface: {{.Interface}}" +
			"\n{{pfx 3}}Ip Address: {{.IpAddress}}" +
			"\n{{pfx 3}}HwAddress: {{.HwAddress}}" +

			//End iterate over interface.
			"{{end}}" +

			// End print
			"\n{{end}}" +

			"{{end}}" +

			"{{end}}{{end}}"))

	return Template
}

func createlRouteTemplate() *template.Template {

	FuncMap := template.FuncMap{
		"setBold": setBold,
		"pfx":     getPrefix,
	}

	Template := template.Must(template.New("lroute").Funcs(FuncMap).Parse(
		"{{$print := .PrintConf}}" +
			"{{with .Config}}{{with .LinuxConfig}}" +
			"{{with .Routes}}\n{{pfx 1}}{{setBold \"Linux Route\"}}" +

			"{{if $print}}:" +

			// Iterate over Route.
			"{{range .}}" +

			"\n{{pfx 3}}Outgoing Interface: {{.OutgoingInterface}}" +
			"\n{{pfx 3}}Scope: {{.Scope}}" +
			"\n{{pfx 3}}Dst Network: {{.DstNetwork}}" +
			"\n{{pfx 3}}Gw Addr: {{.GwAddr}}" +
			"\n{{pfx 3}}Metric: {{.Metric}}" +

			//End
			"{{end}}" +

			//End iterate over interface.
			"{{end}}" +

			// End print
			"\n{{end}}" +

			"{{end}}{{end}}"))

	return Template
}

func printList(data *VppData, buffer *bytes.Buffer) {
	vppdata := data.Config.GetVppConfig()
	linuxData := data.Config.GetLinuxConfig()

	if vppdata.GetAcls() != nil {
		fmt.Fprintf(buffer, "vpp %s\n", vpp_acl.ModelACL.Type)
	}

	if vppdata.GetArps() != nil {
		fmt.Fprintf(buffer, "vpp %s\n", vpp_l3.ModelARPEntry.Type)
	}

	if vppdata.GetBridgeDomains() != nil {
		fmt.Fprintf(buffer, "vpp %s\n", vpp_l2.ModelBridgeDomain.Type)
	}

	if vppdata.GetDnat44S() != nil {
		fmt.Fprintf(buffer, "vpp %s\n", vpp_nat.ModelDNat44.Type)
	}

	if vppdata.GetFibs() != nil {
		fmt.Fprintf(buffer, "vpp %s\n", vpp_l2.ModelFIBEntry.Type)
	}

	if vppdata.GetInterfaces() != nil {
		fmt.Fprintf(buffer, "vpp %s\n", vpp_interfaces.ModelInterface.Type)
	}

	if vppdata.GetIpscanNeighbor() != nil {
		fmt.Fprintf(buffer, "vpp %s\n", vpp_l3.ModelIPScanNeighbor.Type)
	}

	if vppdata.GetIpsecSas() != nil {
		fmt.Fprintf(buffer, "vpp %s\n", vpp_ipsec.ModelSecurityAssociation.Type)
	}

	if vppdata.GetIpsecSpds() != nil {
		fmt.Fprintf(buffer, "vpp %s\n", vpp_ipsec.ModelSecurityPolicyDatabase.Type)
	}

	if vppdata.GetNat44Global() != nil {
		fmt.Fprintf(buffer, "vpp %s\n", vpp_nat.ModelNat44Global.Type)
	}

	if vppdata.GetProxyArp() != nil {
		fmt.Fprintf(buffer, "vpp %s\n", vpp_l3.ModelProxyARP.Type)
	}

	if vppdata.GetRoutes() != nil {
		fmt.Fprintf(buffer, "vpp %s\n", vpp_l3.ModelRoute.Type)
	}

	if vppdata.GetXconnectPairs() != nil {
		fmt.Fprintf(buffer, "vpp %s\n", vpp_l2.ModelXConnectPair.Type)
	}

	if linuxData.GetRoutes() != nil {
		fmt.Fprintf(buffer, "linux %s\n", linux_l3.ModelRoute.Type)
	}

	if linuxData.GetInterfaces() != nil {
		fmt.Fprintf(buffer, "linux %s\n", linux_interfaces.ModelInterface.Type)
	}

	if linuxData.GetArpEntries() != nil {
		fmt.Fprintf(buffer, "linux %s\n", linux_l3.ModelARPEntry.Type)
	}
}

// Render data according to templates in text form.
func (ed EtcdDump) textRenderer(showConf bool, templates []*template.Template) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	for _, key := range ed.getSortedKeys() {
		vd, _ := ed[key]
		vd.ShowConf = showConf
		vd.PrintConf = showConf

		if showConf {
			for _, templateVal := range templates {
				err := templateVal.Execute(buffer, vd)
				if err != nil {
					return nil, err
				}
			}
		} else {
			printList(vd, buffer)
		}
	}
	return buffer, nil
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

// setOsColor sets the color for the Operational State.
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

func setStsColor(kind string, arg interfaces.InterfaceState_Status) string {
	sts := fmt.Sprintf("%s-%s", kind, arg)
	switch arg {
	case interfaces.InterfaceState_UP:
		return setGreen(sts)
	case interfaces.InterfaceState_DOWN:
		return setRed(sts)
	default:
		return sts
	}
}

// getIPAddresses gets a list of IPv4 addresses configured on an
// interface. The parameters are returned as a formatted string
// ready to be printed out.
func getIPAddresses(addrs []string) string {
	return strings.Join(addrs, ", ")
}
