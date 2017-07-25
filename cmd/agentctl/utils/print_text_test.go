package utils_test


import (
	data "github.com/ligato/vpp-agent/cmd/agentctl/testing"
	"testing"
	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
	"strconv"
	"github.com/onsi/gomega"
	"strings"
)

// Test01VppInterfacePrintText verifies presence of every VPP and an interface from the input data
func Test01VppInterfacePrintText(t *testing.T) {
	etcdDump := utils.NewEtcdDump()
	etcdDump = data.TableData()

	result := etcdDump.PrintDataAsText(false, false)
	gomega.Expect(result).ToNot(gomega.BeNil())
	output := result.String()

	// Check Vpp and interface presence
	for i := 1; i <= 3; i++ {
		vppName := "vpp-" + strconv.Itoa(i)
		gomega.Expect(strings.Contains(output, vppName)).To(gomega.BeTrue())
		for j := 1; j <= 3; j++ {
			interfaceName := vppName + "-interface-" + strconv.Itoa(j)
			gomega.Expect(strings.Contains(output, interfaceName)).To(gomega.BeTrue())
		}
	}
}

// Test02StatusPrintText tests presence of status flags in the output
func Test02StatusPrintText(t *testing.T) {
	etcdDump := utils.NewEtcdDump()
	etcdDump = data.TableData()

	result := etcdDump.PrintDataAsText(false, false)
	gomega.Expect(result).ToNot(gomega.BeNil())
	output := result.String()

	// Tested  flags
	notInCfg := "NOT-IN-CONFIG"
	adminUp := "ADMIN-UP"
	adminDown := "ADMIN-DOWN"
	operUp := "OPER-UP"
	operDown := "OPER-DOWN"

	// Status flag expected in every interface
	gomega.Expect(strings.Contains(output, notInCfg)).To(gomega.BeTrue())
	gomega.Expect(strings.Count(output, notInCfg)).To(gomega.BeEquivalentTo(9))
	// Status flag expected in every active interface
	gomega.Expect(strings.Contains(output, adminUp)).To(gomega.BeTrue())
	gomega.Expect(strings.Count(output, adminUp)).To(gomega.BeEquivalentTo(6))
	// Status flag expected in every inactive interface
	gomega.Expect(strings.Contains(output, adminDown)).To(gomega.BeTrue())
	gomega.Expect(strings.Count(output, adminDown)).To(gomega.BeEquivalentTo(3))
	// Status flag expected in every active interface
	gomega.Expect(strings.Contains(output, operUp)).To(gomega.BeTrue())
	gomega.Expect(strings.Count(output, operUp)).To(gomega.BeEquivalentTo(6))
	// Status flag expected in every inactive interface
	gomega.Expect(strings.Contains(output, operDown)).To(gomega.BeTrue())
	gomega.Expect(strings.Count(output, operDown)).To(gomega.BeEquivalentTo(3))
}

// Test03InterfaceStatsPrintText tests presence of state flags on active interfaces in output
func Test03InterfaceStatsPrintText(t *testing.T) {
	etcdDump := utils.NewEtcdDump()
	etcdDump = data.TableData()

	result := etcdDump.PrintDataAsText(false, false)
	gomega.Expect(result).ToNot(gomega.BeNil())
	output := result.String()

	statsFlags := []string{"Stats", "In:", "Out:", "Misc:"}

	for _, flag := range statsFlags {
		gomega.Expect(strings.Contains(output, flag)).To(gomega.BeTrue())
		// Flags are expected in every active interface
		gomega.Expect(strings.Count(output, flag)).To(gomega.BeEquivalentTo(6))
	}
}