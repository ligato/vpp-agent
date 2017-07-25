package utils_test

import (
	"fmt"
	"github.com/onsi/gomega"
	data "github.com/ligato/vpp-agent/cmd/agentctl/testing"
	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
	"strconv"
	"strings"
	"testing"
)

// Test01VppInterfacesPrintJsonData tests VPPs and interfaces presence in the output + the presence of a statistics data
// (the header in every interface and the data flags in active interfaces)
func Test01VppInterfacesPrintJsonData(t *testing.T) {
	etcdDump := utils.NewEtcdDump()
	etcdDump = data.TableData()

	result, _ := etcdDump.PrintDataAsJSON(nil)
	gomega.Expect(result).ToNot(gomega.BeNil())

	output := result.String()

	fmt.Print(output)

	// Check Vpp and interface presence
	for i := 1; i <= 3; i++ {
		vppName := "vpp-" + strconv.Itoa(i)
		gomega.Expect(strings.Contains(output, vppName)).To(gomega.BeTrue())
		// Interface Root
		gomega.Expect(strings.Contains(output, "interface")).To(gomega.BeTrue())
		for j := 1; j <= 3; j++ {
			gomega.Expect(strings.Contains(output, "Test-Interface")).To(gomega.BeTrue())
		}
	}

	// Test statistics presence (including empty)
	gomega.Expect(strings.Contains(output, "statistics")).To(gomega.BeTrue())
	gomega.Expect(strings.Count(output, "statistics")).To(gomega.BeEquivalentTo(9)) // Interface count

	// Test statistics data
	dataFlags := []string{"in_packets", "out_packets", "in_miss_packets"}
	for _, flag := range dataFlags {
		gomega.Expect(strings.Contains(output, flag)).To(gomega.BeTrue())
		// Interfaces with statistics data
		gomega.Expect(strings.Count(output, flag)).To(gomega.BeEquivalentTo(6))
	}
}

// Test02PrintJsonMetadata tests presence of a metadata in the output in case the 'showEtcd' switch is set to true. The metadata should
// be present on every interface
func Test02PrintJsonMetadata(t *testing.T) {
	etcdDump := utils.NewEtcdDump()
	etcdDump = data.TableData()

	result, _ := etcdDump.PrintDataAsJSON(nil)
	gomega.Expect(result).ToNot(gomega.BeNil())
	output := result.String()

	gomega.Expect(strings.Contains(output, "Keys")).To(gomega.BeTrue())
	count := strings.Count(output, "Keys")
	gomega.Expect(count).To(gomega.BeEquivalentTo(3))
}
