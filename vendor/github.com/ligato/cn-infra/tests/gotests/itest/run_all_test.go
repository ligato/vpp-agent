package itest

import (
	"os"
	"os/signal"
	"reflect"
	"strings"
	"testing"
)

// Test runs all TC methods of multiple test suites in sequence
func Test(t *testing.T) {
	doneChan := make(chan struct{}, 1)

	go func() {
		RunTestSuite(&suiteFlavorLocal{T: t}, t)
		RunTestSuite(&suiteFlavorRPC{T: t}, t)
		RunTestSuite(&suiteFlavorAllConnectors{T: t}, t)

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

// RunTestSuite use reflection to run each method prefixed with "TC"
func RunTestSuite(testSuite interface{}, t *testing.T, teardowns ...func()) {
	vppInstanceCounter := 0 //each test uses different ETCD subtree

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
