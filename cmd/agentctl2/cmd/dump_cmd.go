package cmd

import (
	"fmt"
	"os"

	"github.com/ligato/vpp-agent/cmd/agentctl2/restapi"
	"github.com/ligato/vpp-agent/plugins/restapi/resturl"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands.
var dumpCmd = &cobra.Command{
	Use:     "dump",
	Aliases: []string{"d"},
	Short:   "Dump command",
	Long: `
	Dump command
`,
	Args: cobra.MinimumNArgs(2),
}

var dumpLinuxInterface = &cobra.Command{
	Use:   "LinuxInterface",
	Short: "Dump Linux Interface",
	Long: `
	Dump Linux Interface
`,
	Args: cobra.MaximumNArgs(0),
	Run:  linuxInterfaceDumpFunction,
}

var dumpLinuxRoutes = &cobra.Command{
	Use:   "LinuxRoutes",
	Short: "Dump Linux Routes",
	Long: `
	Dump Linux Routes
`,
	Args: cobra.MaximumNArgs(0),
	Run:  linuxRoutesDumpFunction,
}

var dumpLinuxArps = &cobra.Command{
	Use:   "LinuxArps",
	Short: "Dump Linux Arps",
	Long: `
	Dump Linux Arps
`,
	Args: cobra.MaximumNArgs(0),
	Run:  linuxArpsDumpFunction,
}

var dumpACLIP = &cobra.Command{
	Use:   "ACLIP",
	Short: "Dump ACL IP prefix",
	Long: `
	Dump ACL IP prefix
`,
	Args: cobra.MaximumNArgs(0),
	Run:  aclIPDumpFunction,
}

var dumpACLMACIP = &cobra.Command{
	Use:   "ACLMACIP",
	Short: "Dump ACL MAC IP prefix",
	Long: `
	Dump ACL MAC IP prefix
`,
	Args: cobra.MaximumNArgs(0),
	Run:  aclMACIPDumpFunction,
}

var dumpInterface = &cobra.Command{
	Use:   "Interface",
	Short: "Dump Interface",
	Long: `
	Dump Interface
`,
	Args: cobra.MaximumNArgs(0),
	Run:  interfaceDumpFunction,
}

var dumpLoopback = &cobra.Command{
	Use:   "Loopback",
	Short: "Dump Loopback",
	Long: `
	Dump Loopback
`,
	Args: cobra.MaximumNArgs(0),
	Run:  loopbackDumpFunction,
}

var dumpEthernet = &cobra.Command{
	Use:   "Ethernet",
	Short: "Dump Ethernet",
	Long: `
	Dump Ethernet
`,
	Args: cobra.MaximumNArgs(0),
	Run:  ethernetDumpFunction,
}

var dumpMemif = &cobra.Command{
	Use:   "Memif",
	Short: "Dump memif interface",
	Long: `
	Dump memif interface
`,
	Args: cobra.MaximumNArgs(0),
	Run:  memifDumpFunction,
}

var dumpTap = &cobra.Command{
	Use:   "Tap",
	Short: "Dump Tap interface",
	Long: `
	Dump Tap interface
`,
	Args: cobra.MaximumNArgs(0),
	Run:  tapDumpFunction,
}

var dumpAfPacket = &cobra.Command{
	Use:   "AfPacket",
	Short: "Dump af-packet interface",
	Long: `
	Dump af-packet interface
`,
	Args: cobra.MaximumNArgs(0),
	Run:  afPacketDumpFunction,
}

var dumpVxLan = &cobra.Command{
	Use:   "VxLan",
	Short: "Dump vxlan interface",
	Long: `
	Dump vxlan interface
`,
	Args: cobra.MaximumNArgs(0),
	Run:  vxlanDumpFunction,
}

var dumpNatGlobal = &cobra.Command{
	Use:   "NatGlobal",
	Short: "Dump global NAT config",
	Long: `
	Dump global NAT config
`,
	Args: cobra.MaximumNArgs(0),
	Run:  natGlobalDumpFunction,
}

var dumpNatDNat = &cobra.Command{
	Use:   "NatDNat",
	Short: "Dump DNAT configurations",
	Long: `
	Dump DNAT configurations
`,
	Args: cobra.MaximumNArgs(0),
	Run:  dnatDumpFunction,
}

var dumpBd = &cobra.Command{
	Use:   "Bd",
	Short: "Dump Bridge domain",
	Long: `
	Dump Bridge domain
`,
	Args: cobra.MaximumNArgs(0),
	Run:  bridgeDomainDumpFunction,
}

var dumpFib = &cobra.Command{
	Use:   "Fib",
	Short: "Dump Fib",
	Long: `
	Dump Fib
`,
	Args: cobra.MaximumNArgs(0),
	Run:  fibDumpFunction,
}

var dumpXc = &cobra.Command{
	Use:   "Xc",
	Short: "Dump cross-connect",
	Long: `
	Dump cross-connect
`,
	Args: cobra.MaximumNArgs(0),
	Run:  xcDumpFunction,
}

var dumpRoutes = &cobra.Command{
	Use:   "Routes",
	Short: "Dump static route",
	Long: `
	Dump static route
`,
	Args: cobra.MaximumNArgs(0),
	Run:  routeDumpFunction,
}

var dumpArps = &cobra.Command{
	Use:   "Arps",
	Short: "Dump ARPs",
	Long: `
	Dump ARPs
`,
	Args: cobra.MaximumNArgs(0),
	Run:  arpDumpFunction,
}

var dumpPArpIfs = &cobra.Command{
	Use:   "PArpIfs",
	Short: "Dump PArpIfs",
	Long: `
	Dump PArpIfs
`,
	Args: cobra.MaximumNArgs(0),
	Run:  parpifsDumpFunction,
}

var dumpPArpRngs = &cobra.Command{
	Use:   "PArpRngs",
	Short: "Dump proxy ARP ranges",
	Long: `
	Dump ARP ranges
`,
	Args: cobra.MaximumNArgs(0),
	Run:  pArpRngsDumpFunction,
}

var dumpCommand = &cobra.Command{
	Use:   "Command",
	Short: "Dump Command",
	Long: `
	Dump Command
`,
	Args: cobra.MaximumNArgs(0),
	Run:  commandDumpFunction,
}

var dumpTelemetry = &cobra.Command{
	Use:   "Telemetry",
	Short: "Dump telemetry",
	Long: `
	Dump telemetry
`,
	Args: cobra.MaximumNArgs(0),
	Run:  telemetryDumpFunction,
}

var dumpTMemory = &cobra.Command{
	Use:   "TMemory",
	Short: "Dump telemetry memory",
	Long: `
	Dump telemetry memory
`,
	Args: cobra.MaximumNArgs(0),
	Run:  telemetryMemoryDumpFunction,
}

var dumpTRuntime = &cobra.Command{
	Use:   "TRuntime",
	Short: "Dump telemetry runtime",
	Long: `
	Dump telemetry runtime
`,
	Args: cobra.MaximumNArgs(0),
	Run:  telemetryRuntimeDebugFunction,
}

var dumpTNodeCount = &cobra.Command{
	Use:   "TNodeCount",
	Short: "Dump telemetry node count",
	Long: `
	Dump telemetry node count
`,
	Args: cobra.MaximumNArgs(0),
	Run:  telemetryNodeCountDumpFunction,
}

var dumpTracer = &cobra.Command{
	Use:   "Tracer",
	Short: "Trace binary API calls",
	Long: `
	Trace binary API calls
`,
	Args: cobra.MaximumNArgs(0),
	Run:  traceDumpFunction,
}

var dumpIndex = &cobra.Command{
	Use:   "Index",
	Short: "Dump full index page",
	Long: `
	Dump full index page
`,
	Args: cobra.MaximumNArgs(0),
	Run:  indexDumpFunction,
}

func init() {
	RootCmd.AddCommand(dumpCmd)
	dumpCmd.AddCommand(dumpLinuxInterface)
	dumpCmd.AddCommand(dumpLinuxRoutes)
	dumpCmd.AddCommand(dumpLinuxArps)
	dumpCmd.AddCommand(dumpACLIP)
	dumpCmd.AddCommand(dumpACLMACIP)
	dumpCmd.AddCommand(dumpInterface)
	dumpCmd.AddCommand(dumpLoopback)
	dumpCmd.AddCommand(dumpEthernet)
	dumpCmd.AddCommand(dumpMemif)
	dumpCmd.AddCommand(dumpTap)
	dumpCmd.AddCommand(dumpAfPacket)
	dumpCmd.AddCommand(dumpVxLan)
	dumpCmd.AddCommand(dumpNatGlobal)
	dumpCmd.AddCommand(dumpNatDNat)
	dumpCmd.AddCommand(dumpBd)
	dumpCmd.AddCommand(dumpFib)
	dumpCmd.AddCommand(dumpXc)
	dumpCmd.AddCommand(dumpRoutes)
	dumpCmd.AddCommand(dumpArps)
	dumpCmd.AddCommand(dumpPArpIfs)
	dumpCmd.AddCommand(dumpPArpRngs)
	dumpCmd.AddCommand(dumpCommand)
	dumpCmd.AddCommand(dumpTelemetry)
	dumpCmd.AddCommand(dumpTMemory)
	dumpCmd.AddCommand(dumpTRuntime)
	dumpCmd.AddCommand(dumpTNodeCount)
	dumpCmd.AddCommand(dumpTracer)
	dumpCmd.AddCommand(dumpIndex)
}

func linuxInterfaceDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.LinuxInterface)

	fmt.Fprint(os.Stdout, msg)
}

func linuxRoutesDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.LinuxRoutes)

	fmt.Fprint(os.Stdout, msg)
}

func linuxArpsDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.LinuxArps)

	fmt.Fprint(os.Stdout, msg)
}

func aclIPDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.ACLIP)

	fmt.Fprint(os.Stdout, msg)
}

func aclMACIPDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.ACLMACIP)

	fmt.Fprint(os.Stdout, msg)
}

func interfaceDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.Interface)

	fmt.Fprint(os.Stdout, msg)
}

func loopbackDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.Loopback)

	fmt.Fprint(os.Stdout, msg)
}

func ethernetDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.Ethernet)

	fmt.Fprint(os.Stdout, msg)
}

func memifDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.Memif)

	fmt.Fprint(os.Stdout, msg)
}

func tapDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.Tap)

	fmt.Fprint(os.Stdout, msg)
}

func afPacketDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.AfPacket)

	fmt.Fprint(os.Stdout, msg)
}

func vxlanDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.VxLan)

	fmt.Fprint(os.Stdout, msg)
}

func natGlobalDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.NatGlobal)

	fmt.Fprint(os.Stdout, msg)
}

func dnatDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.NatDNat)

	fmt.Fprint(os.Stdout, msg)
}

func bridgeDomainDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.Bd)

	fmt.Fprint(os.Stdout, msg)
}

func fibDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.Fib)

	fmt.Fprint(os.Stdout, msg)
}

func xcDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.Xc)

	fmt.Fprint(os.Stdout, msg)
}

func routeDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.Routes)

	fmt.Fprint(os.Stdout, msg)
}

func arpDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.Arps)

	fmt.Fprint(os.Stdout, msg)
}

func parpifsDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.PArpIfs)

	fmt.Fprint(os.Stdout, msg)
}

func pArpRngsDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.PArpRngs)

	fmt.Fprint(os.Stdout, msg)
}

func commandDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.Command)

	fmt.Fprint(os.Stdout, msg)
}

func telemetryDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.Telemetry)

	fmt.Fprint(os.Stdout, msg)
}

func telemetryMemoryDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.TMemory)

	fmt.Fprint(os.Stdout, msg)
}
func telemetryRuntimeDebugFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.TRuntime)

	fmt.Fprint(os.Stdout, msg)
}
func telemetryNodeCountDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.TNodeCount)

	fmt.Fprint(os.Stdout, msg)
}
func traceDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.Tracer)

	fmt.Fprint(os.Stdout, msg)
}
func indexDumpFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, resturl.Index)

	fmt.Fprint(os.Stdout, msg)
}
