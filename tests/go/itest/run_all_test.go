package itest

import (
	"reflect"
	"strings"
	"testing"

	"github.com/ligato/vpp-agent/clientv1/defaultplugins/localclient"
	"github.com/ligato/vpp-agent/tests/go/itest/iftst"
	"github.com/ligato/vpp-agent/tests/go/itest/testutil"
	"github.com/onsi/gomega"
)

// Test runs all TC methods of multiple test suites in a sequence.
func Test(t *testing.T) {
	suite := &suiteMemif{T: t,
		When: testutil.When{
			WhenIface: iftst.WhenIface{
				Log:       testutil.NewLogger("WhenIface", t),
				NewChange: localclient.DataChangeRequest,
				NewResync: localclient.DataResyncRequest,
			}},
		Then: testutil.Then{
			ThenIface: iftst.ThenIface{
				Log:       testutil.NewLogger("ThenIface", t),
				NewChange: localclient.DataChangeRequest,
			}},
	}
	RunTestSuite(suite, t)
}

// RunTestSuite uses reflection to run each method prefixed with "TC".
func RunTestSuite(testSuite interface{}, t *testing.T, teardowns ...func()) {
	vppInstanceCounter := 0 // Each test uses different ETCD subtree.

	suite := reflect.ValueOf(testSuite)
	suiteName := reflect.TypeOf(testSuite).Elem().Name()

	t.Log("Suite:", suiteName)
	t.Run(suiteName, func(t *testing.T) {
		for i := 0; i < suite.NumMethod(); i++ {
			tc := suite.Method(i)
			tcName := suite.Type().Method(i).Name

			if strings.HasPrefix(tcName, "TC") {
				//TODO: currently the repeated modification of env var are ignore by flags
				vppInstanceCounter++

				//TODO inject the Microservice Label
				//os.Setenv(servicelabel.MicroserviceLabelEnvVar, fmt.Sprintf(
				//	"TEST_VPP_%d", vppInstanceCounter))

				t.Log("Case :", tcName)
				ok := t.Run(tcName, func(t *testing.T) {
					gomega.RegisterTestingT(t)
					tc.Call([]reflect.Value{})

					t.Log("Finished TC ", suiteName, " ", tcName)
				})
				for _, teardown := range teardowns {
					teardown()
				}
				if !ok {
					t.Log("FAILED TC ", suiteName, " ", tcName)
					break
				}
			}
		}

	})
}
