package utils

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

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
	prefixer   pfx
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

func (ed EtcdDump) PrintTest(showConf bool) (*bytes.Buffer, error) {
	prefixer = newPrefixer(false, perLevelSpaces)


 	ifTemplate := createInterfaceTemplate()
 	aclTemplate := createAclTemplate()
 	bdTemplate := createBridgeTemplate()
 	fibTemplate := createFibTableTemplate()
 	xconnectTemplate := createXconnectTableTemplate()
 	arpTemplate := createArpTableTemplate()
 	routeTemplate := createRouteTableTemplate()
	proxyarpTemplate := createProxyArpTemplate()
	ipneighbor := createIPScanNeightTemplate()
	nat := createNATTemplate()
	dnat := createDNATTemplate()

	templates := []*template.Template{}
	// Keep template order
	templates = append(templates, ifTemplate, aclTemplate, bdTemplate, fibTemplate,
		xconnectTemplate, arpTemplate, routeTemplate, proxyarpTemplate, ipneighbor,
		nat, dnat)

	return ed.textRenderer(showConf, templates)
}

func createAclTemplate() (*template.Template) {

	FuncMap := template.FuncMap{
		"setBold":        setBold,
		"isEnabled":      isEnabled,
		"getIpAddresses": getIPAddresses,
		"pfx":            getPrefix,
	}

	Template, err := template.New("acl").Funcs(FuncMap).Parse(
		"{{$conf := .ShowConf}}" +
			"{{with .ACL}}\n{{pfx 1}}ACL:" +

		// Iterate over ACL.
			"{{range $ACLName, $ACLData := .}}\n{{pfx 2}}{{setBold $ACLName}}" +
			"{{if $conf}}" +
			"{{with .Config}}{{with .ACL}}" +

		//Start Rules
			"{{with .Rules}}\n{{pfx 3}}ACL Rules:" +
		// Iterate over ACL rule.
			"{{range $RuleName, $RuleData := .}}" +
			"{{with .Action}}\n{{pfx 4}}Action: {{.}}{{end}}" +

		//Start IP Rule
			"{{with .IpRule}}\n{{pfx 4}}IP Rule:" +

			//Start IP
			"{{with .Ip}}\n{{pfx 5}}IP:" +

			"{{with .DestinationNetwork}}\n{{pfx 6}}Destination Network:" +
			"{{.}}{{end}}" +

			"{{with .SourceNetwork}}\n{{pfx 6}}Source Network:" +
			"{{.}}{{end}}" +

		// End IP
			"{{end}}" +

			"{{with .Icmp}}\n{{pfx 5}}ICMP:" +
			"{{with .Icmpv6}}\n{{pfx 6}}Is ICMPv6{{isEnabled .}}{{end}}" +

		//Start ICMP Code Range
			"{{with .IcmpCodeRange}}\n{{pfx 6}}ICMP Code Range:" +
			"{{with .First}}\n{{pfx 7}}First: {{.}}{{end}}" +
			"{{with .Last}}\n{{pfx 7}}Last: {{.}}{{end}}" +

		//End ICMP Code Range
			"{{end}}" +

		//Start ICMP Type Range
			"{{with .IcmpTypeRange}}\n{{pfx 6}}ICMP Type Range" +
			"{{with .First}}\n{{pfx 7}}First: {{.}}{{end}}" +
			"{{with .Last}}\n{{pfx 7}}Last: {{.}}{{end}}" +

		//End ICMP Type Range
			"{{end}}" +

		//End ICMP
			"{{end}}" +

		//Start TCP
			"{{with .Tcp}}\n{{pfx 5}}TCP:" +
			"{{with .DestinationPortRange}}\n{{pfx 6}}Destination Port Range:" +
			"{{with .LowerPort}}\n{{pfx 7}}Lower Port: {{.}}{{end}}" +
			"{{with .UpperPort}}\n{{pfx 7}}Upper Port: {{.}}{{end}}" +

		//End DestinationPortRange
			"{{end}}" +

			"{{with .SourcePortRange}}\n{{pfx 6}}Source Port Range" +
			"{{with .LowerPort}}\n{{pfx 7}}Lower Port: {{.}}{{end}}" +
			"{{with .UpperPort}}\n{{pfx 7}}Upper Port: {{.}}{{end}}" +

		//End SourcePortRange
			"{{end}}" +

			"{{with .TcpFlagsMask}}\n{{pfx 6}}TCP Flags Mask: {{.}}{{end}}" +
			"{{with .TcpFlagsValue}}\n{{pfx 6}}TCP Flags Value: {{.}}{{end}}" +

		//End TCP
			"{{end}}" +

		//Start UDP
			"{{with .Udp}}\n{{pfx 5}}UDP:" +
		//Start UDP Destinaton Port
			"{{with .DestinationPortRange}}\n{{pfx 6}}Destination Port Range:" +
			"{{with .LowerPort}}\n{{pfx 7}}Lower Port: {{.}}{{end}}" +
			"{{with .UpperPort}}\n{{pfx 7}}Upper Port: {{.}}{{end}}" +

		//End DestinationPortRange
			"{{end}}" +

		//Start UDP Source Port
			"{{with .SourcePortRange}}\n{{pfx 6}}Source Port Range:" +
			"{{with .LowerPort}}\n{{pfx 7}}Lower Port: {{.}}{{end}}" +
			"{{with .UpperPort}}\n{{pfx 7}}Upper Port: {{.}}{{end}}" +

		//End SourcePortRange
			"{{end}}" +

		//End UDP
			"{{end}}" +

		//End IP Rule
			"{{end}}" +

		//Macip Rule
			"{{with .MacipRule}}\n{{pfx 4}}Macip Rule:" +
			"{{with .SourceAddress}}\n{{pfx 5}}Source Address: {{getIpAddresses .}}{{end}}" +
			"{{with .SourceAddressPrefix}}/{{.}}{{end}}" +

			"{{with .SourceMacAddress}}\n{{pfx 5}}Source MacAddress: {{.}}{{end}}" +
			"{{with .SourceMacAddressMAsk}}/{{.}}{{end}}" +

		//End Macip Rule
			"{{end}}" +

		//End Rule range
			"{{end}}" +

		// End iterate over ACL rule.
			"{{end}}" +

			//Interfaces
			"{{with .Interfaces}}\n{{pfx 3}} Interfaces:" +
			"{{with .Egress}}\n{{pfx 4}}Egress: {{.}}{{end}}" +
			"{{with .Ingress}}\n{{pfx 4}}Ingress: {{.}}{{end}}" +
			"{{end}}" +

			"{{end}}{{end}}" +

		// End Config
			"{{end}}" +

			"{{end}}{{end}}")

	if err != nil {
		panic(err)
	}

	return Template
}

func createInterfaceTemplate() (*template.Template) {

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

	Template, err := template.New("interfaces").Funcs(ifFuncMap).Parse(
		"{{$conf := .ShowConf}}" +
			"{{with .Interfaces}}\n{{pfx 1}}INTERFACES:" +

		// Iterate over interfaces.
			"{{range $ifaceName, $ifaceData := .}}\n{{pfx 2}}{{setBold $ifaceName}}\n" +

			"{{if $conf}}" +
			// Interface overall status
			"{{with .Config}}{{with .Interface}}{{pfx 3}}{{isEnabled .Enabled}}{{end}} " +
			// 'with .Config' else
			"{{else}}{{setRed \"NOT-IN-CONFIG\"}}{{end}}" +

			// Interface type
			"{{with .Config}}{{with .Interface}}\n{{pfx 3}}IfType: {{.Type}}" +
			"{{end}}" +

			"{{with .Interface}}\n{{pfx 3}}PhyAddr: {{.PhysAddress}}" +
			"{{end}}" +

			// IP Address and attributes (if configured)
			"{{with .Interface}}{{with .IpAddresses}}\n{{pfx 3}}IpAddr: {{getIpAddresses .}}" +
			"{{end}}{{end}}" +

			"{{with .Interface}}\n{{pfx 3}}Mtu: {{.Mtu}}" +
			"{{end}}" +

			"{{end}}" +
			// End Config
			"{{end}}" +

			"{{end}}{{end}}\n")

	if err != nil {
		panic(err)
	}

	return Template
}

func createBridgeTemplate() (*template.Template) {

	FuncMap := template.FuncMap{
		"setBold":        setBold,
		"isEnabled":      isEnabled,
		"getIpAddresses": getIPAddresses,
		"pfx":            getPrefix,
	}

	Template, err := template.New("bridgeDomains").Funcs(FuncMap).Parse(
		"{{$conf := .ShowConf}}" +
			"{{with .BridgeDomains}}\n{{pfx 1}}Bridge:" +

			"{{range $BridgeName, $BridgeData := .}}\n{{pfx 2}}{{setBold $BridgeName}}" +
			"{{if $conf}}" +
			"{{with .Config}}{{with .BridgeDomain}}" +

			"{{with .Flood}}\n{{pfx 3}}Flood: {{isEnabled .}}{{end}}" +
			"{{with .UnknownUnicastFlood}}\n{{pfx 3}}Unknown Unicast Flood: {{isEnabled .}}{{end}}" +
			"{{with .Forward}}\n{{pfx 3}}Forward: {{isEnabled .}}{{end}}" +
			"{{with .Learn}}\n{{pfx 3}}Learn: {{isEnabled .}}{{end}}" +
			"{{with .ArpTermination}}\n{{pfx 3}}ArpTermination: {{isEnabled .}}{{end}}" +
			"{{with .MacAge}}\n{{pfx 3}}: {{.}}{{end}}" +

			//Interate over interfaces.
			"{{range $InterfaceName, $InterfaceData := .Interfaces}}\n{{pfx 3}}{{setBold $InterfaceName}}" +
			"{{with .BridgedVirtualInterface}}\n{{pfx 4}}BridgedVirtualInterface: {{isEnabled .}}{{end}}" +
			"{{with .SplitHorizonGroup}}\n{{pfx 4}}SplitHorizonGroup: {{.}} {{end}}" +

			//End interate interfaces.
			"{{end}}" +

		//Interate over ArpTerminationTable.
			"{{range $ArpTableName, $ArpTableData := .ArpTerminationTable}}\n{{pfx 3}}{{setBold $ArpTableName}}" +
			"{{with .IpAddress}}\n{{pfx 4}}IP Address: {{getIpAddresses .}} {{end}}" +
			"{{with .PhysAddress}}\n{{pfx 4}}: {{.}} {{end}}" +

			//End interate ArpTerminationTable.
			"{{end}}" +

			"{{end}}{{end}}" +
			//End Config
			"{{end}}" +

			"{{end}}{{end}}")


	if err != nil {
		panic(err)
	}

	return Template
}

func createFibTableTemplate() (*template.Template) {

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

	Template, err := template.New("fibTable").Funcs(FuncMap).Parse(
		"{{$conf := .ShowConf}}" +
			"{{with .FibTableEntries}}\n{{pfx 1}}Fib Table:" +

			"{{if $conf}}" +
			"{{with .Config}}{{with .FIBEntry}}" +

			"{{with .PhysAddress}}\n{{pfx 3}}MAC address: {{.}}{{end}}" +
			"{{with .BridgeDomain}}\n{{pfx 3}}Bridge Domain: {{.}}{{end}}" +
			"{{with .OutgoingInterface}}\n{{pfx 3}}Outgoing Interface: {{.}}{{end}}" +
			"{{with .Action}}\n{{pfx 3}}Action: {{.Action}}{{end}}" +
			"{{with .StaticConfig}}\n{{pfx 3}}Static Config: {{isEnabled .}}{{end}}" +
			"{{with .BridgedVirtualInterface}}{{pfx 3}}Bridge Virtual Interface: {{isEnabled .}}{{end}}" +

			"{{end}}{{end}}" +

		//End config
			"{{end}}" +
			"{{end}}")

	if err != nil {
		panic(err)
	}

	return Template
}

func createXconnectTableTemplate() (*template.Template) {

	FuncMap := template.FuncMap{
		"setBold":        setBold,
		"pfx":            getPrefix,
	}

	Template, err := template.New("xconnect").Funcs(FuncMap).Parse(
		"{{$conf := .ShowConf}}" +
			"{{with .XConnectPairs}}\n{{pfx 1}}Xconnect:" +

		// Iterate over xconnect.
			"{{range $XconnectName, $XconnectData := .}}\n{{pfx 2}}{{setBold $XconnectName}}" +
			"{{if $conf}}" +

			"{{with .Config}}{{with .Xconnect}}" +

			"{{with .ReceiveInterface}}\n{{pfx 3}}Receive Interface: {{.}}{{end}}" +
			"{{with .TransmitInterface}}\n{{pfx 3}}Transmit Interface: {{.}}{{end}}" +

			//End Xconnect
			"{{end}}{{end}}" +
			//End Conf
			"{{end}}" +
			"{{end}}{{end}}")

	if err != nil {
		panic(err)
	}

	return Template
}

func createArpTableTemplate() (*template.Template) {

	FuncMap := template.FuncMap{
		"setBold":        setBold,
		"getIpAddresses": getIPAddresses,
		"isEnabled":      isEnabled,
		"pfx":            getPrefix,
	}

	Template, err := template.New("arp").Funcs(FuncMap).Parse(
			"{{with .ARP}}\n{{pfx 1}}ARP:" +
			"{{with .Config}}{{with .ARPEntry}}" +

			"{{with .Interface}}\n{{pfx 2}}Interface: {{.}}{{end}}" +
			"{{with .IpAddress}}\n{{pfx 2}}IP address : {{.}}{{end}}" +
			"{{with .PhysAddress}}\n{{pfx 2}}MAC: {{.}}{{end}}" +
			"{{with .Static}}\n{{pfx 2}}Static: {{isEnabled .}}{{end}}" +

			//End
			"{{end}}{{end}}" +
			"{{end}}")

	if err != nil {
		panic(err)
	}

	return Template
}

func createRouteTableTemplate() (*template.Template) {

	FuncMap := template.FuncMap{
		"setBold":        setBold,
		"getIpAddresses": getIPAddresses,
		"pfx":            getPrefix,
	}

	Template, err := template.New("routetable").Funcs(FuncMap).Parse(
		"{{with .StaticRoutes}}\n{{pfx 1}}Route Table:" +

			"{{with .Config}}{{with .Route}}" +
			"{{with .Type}}\n{{pfx 2}}Type: {{.}}{{end}}" +
			"{{with .VrfId}}\n{{pfx 2}}VrfId: {{.}}{{end}}" +
			"{{with .DstNetwork}}\n{{pfx 2}}Destination Address: {{.}}{{end}}" +
			"{{with .NextHopAddr}}\n{{pfx 2}}Next Hop Address : {{.}}{{end}}" +
			"{{with .OutgoingInterface}}\n{{pfx 2}}Out going Interface: {{.}}{{end}}" +
			"{{with .Weight}}\n{{pfx 2}}Weight: {{.}}{{end}}" +
			"{{with .Preference}}\n{{pfx 2}}Preference: {{.}}{{end}}" +
			"{{with .ViaVrfId}}\n{{pfx 2}}ViaVrfId: {{.}}{{end}}" +

			//End
			"{{end}}{{end}}" +

			"{{end}}")

	if err != nil {
		panic(err)
	}

	return Template
}

func createProxyArpTemplate() (*template.Template) {

	FuncMap := template.FuncMap{
		"setBold":        setBold,
		"pfx":            getPrefix,
	}

	Template, err := template.New("proxyarp").Funcs(FuncMap).Parse(
		"{{with .ProxyARP}}\n{{pfx 1}}Proxy ARP:" +

			"{{with .Config}}{{with .ProxyARP}}" +

		//Iterate over Interfaces
			"{{range $InterfacesName, $InterfaceData := .Interfaces}}" +
			"{{with .Name}}\n{{pfx 2}}{{.}}{{end}}" +

		//End iterate
			"{{end}}" +

		//Iterate over Proxy Ranges
			"{{range $ProxyName, $ProxyData := .Ranges}}" +
			"{{with .FirstIpAddr}}\n{{pfx 2}}First IP Address: {{.}}{{end}}" +
			"{{with .LastIpAddr}}\n{{pfx 2}}Last IP Address: {{.}}{{end}}" +

		//End iterate
			"{{end}}" +

			//End
			"{{end}}{{end}}" +

			"{{end}}")

	if err != nil {
		panic(err)
	}

	return Template
}

func createIPScanNeightTemplate() (*template.Template) {

	FuncMap := template.FuncMap{
		"setBold":        setBold,
		"pfx":            getPrefix,
	}

	Template, err := template.New("ipscanneigh").Funcs(FuncMap).Parse(
		"{{with .IPScanNeight}}\n{{pfx 1}}IP Neighbor:" +

			"{{with .Config}}{{with .IPScanNeighbor}}" +

			"{{with .Mode}}\n{{pfx 2}}Mode: {{.}}{{end}}" +
			"{{with .ScanInterval}}\n{{pfx 2}}Scan Interval: {{.}}{{end}}" +
			"{{with .MaxProcTime}}\n{{pfx 2}}Max Proc Time: {{.}}{{end}}" +
			"{{with .MaxProcTime}}\n{{pfx 2}}Max Proc Time: {{.}}{{end}}" +
			"{{with .MaxUpdate}}\n{{pfx 2}}Max Uptime: {{.}}{{end}}" +
			"{{with .ScanIntDelay}}\n{{pfx 2}}Scan Int Delay: {{.}}{{end}}" +
			"{{with .StaleThreshold}}\n{{pfx 2}}Stale Threshold: {{.}}{{end}}" +

			//End
			"{{end}}{{end}}" +

			"{{end}}")

	if err != nil {
		panic(err)
	}

	return Template
}

func createNATTemplate() (*template.Template) {

	FuncMap := template.FuncMap{
		"setBold":        setBold,
		"pfx":            getPrefix,
	}

	Template, err := template.New("nat").Funcs(FuncMap).Parse(
		"{{with .NAT}}\n{{pfx 1}}NAT:" +

			"{{with .Config}}{{with .Nat44Global}}" +

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

			//End
			"{{end}}{{end}}" +

			"{{end}}")

	if err != nil {
		panic(err)
	}

	return Template
}

func createDNATTemplate() (*template.Template) {

	FuncMap := template.FuncMap{
		"setBold":        setBold,
		"pfx":            getPrefix,
	}

	Template, err := template.New("dnat").Funcs(FuncMap).Parse(
		"{{with .DNAT}}\n{{pfx 1}}DNAT:" +

			"{{with .Config}}{{with .DNat44}}" +

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

			//End
			"{{end}}{{end}}" +

			"{{end}}")

	if err != nil {
		panic(err)
	}

	return Template
}

// Render data according to templates in text form.
func (ed EtcdDump) textRenderer(showConf bool, templates []*template.Template) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	buffer.WriteTo(os.Stdout)
	for _, key := range ed.getSortedKeys() {
		vd, _ := ed[key]
		vd.ShowConf = showConf

		var wasError error
		for _, templateVal := range templates {
			wasError = templateVal.Execute(buffer, vd)
			if wasError != nil {
				return nil, wasError
			}
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