package itest

import (
	"github.com/ligato/vpp-agent/clientv1/defaultplugins/localclient"
	"github.com/ligato/vpp-agent/tests/go/itest/iftst"
	"github.com/ligato/vpp-agent/tests/go/itest/testutil"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"testing"
)

// Test runs all TC methods of multiple test suites in a sequence.
func Test(t *testing.T) {
	doneChan := make(chan struct{}, 1)

	go func() {
		RunTestSuite(&suiteMemif{T: t,
			When: testutil.When{
				WhenIface: iftst.WhenIface{
					Log:       testutil.NewLogger("WhenIface", t),
					NewChange: localclient.DataChangeRequest},
			},
			Then: testutil.Then{
				ThenIface: iftst.ThenIface{
					Log:       testutil.NewLogger("ThenIface", t),
					NewChange: localclient.DataChangeRequest},
				/*TODO OperState
				k := intf.InterfaceKey(data.Name)
				found, _, err = etcdmux.NewRootBroker().GetValue(servicelabel.GetAgentPrefix()+k, ifState)*/
			},
		}, t)
		//RunTestSuite(&suiteBD{T: t}, t)
		//RunTestSuite(&suiteRoute{T: t}, t)

		doneChan <- struct{}{}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	select {
	case <-doneChan:
		t.Log("Tests finished")
	case <-sigChan:
		t.Log("Interrupt received, returning.")
		t.Fatal("Interrupted by user")
		t.SkipNow()
		os.Exit(1) //TODO avoid this workaround
	}
}

// RunTestSuite uses reflection to run each method prefixed with "TC".
func RunTestSuite(testSuite interface{}, t *testing.T, teardowns ...func()) {
	vppInstanceCounter := 0 // Each test uses different ETCD subtree.

	suite := reflect.ValueOf(testSuite)

	suiteName := reflect.TypeOf(testSuite).Elem().Name()
	t.Log("suiteName '", suiteName, "'")
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

				t.Log("tcName ", tcName)
				ok := t.Run(tcName, func(t *testing.T) {
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
