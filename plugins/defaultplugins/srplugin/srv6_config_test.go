// Copyright (c) 2018 Bell Canada, Pantheon Technologies and/or its affiliates.
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

package srplugin_test

import (
	"fmt"
	"testing"

	"git.fd.io/govpp.git/adapter/mock"
	"git.fd.io/govpp.git/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/srv6"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/srplugin"
	"github.com/ligato/vpp-agent/tests/vppcallfake"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

//TODO add more tests: cover remove/modify for localsids/policies/policy segments/steering
//TODO add more tests: cover delayed configuration

var (
	sidA = sid("A::")
	sidB = sid("B::")
	sidC = sid("C::")
	sidD = sid("D::")
)

const (
	errorMessage = "this is test error"
	segmentName1 = "segmentName1"
	segmentName2 = "segmentName2"
	steeringName = "steeringName"
)

// TestAddLocalSID tests all cases where configurator's AddLocalSID is used (except of complicated cases involving multiple configurator methods)
func TestAddLocalSID(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name     string
		Verify   func(srv6.SID, *srv6.LocalSID, error, *vppcallfake.SRv6Calls)
		FailIn   interface{}
		FailWith error
	}{
		{
			Name: "simple addition of local sid",
			Verify: func(sid srv6.SID, data *srv6.LocalSID, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				state := fakeVPPCalls.LocalSIDState()
				recordedData, exists := state[sid.String()]
				Expect(exists).To(BeTrue())
				Expect(recordedData).To(Equal(data))
			},
		},
		{
			Name:     "failure propagation from VPPCall's AddLocalSid",
			FailIn:   vppcallfake.AddLocalSidFuncCall{},
			FailWith: fmt.Errorf(errorMessage),
			Verify: func(sid srv6.SID, data *srv6.LocalSID, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { //wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				configurator, fakeVPPCalls, connection := srv6TestSetup(t)
				defer srv6TestTeardown(connection, configurator)
				sid := sidA
				data := localSID()
				if td.FailIn != nil {
					fakeVPPCalls.FailIn(td.FailIn, td.FailWith)
				}
				err := configurator.AddLocalSID(sid, data)
				td.Verify(sid, data, err, fakeVPPCalls)
			}()
		})
	}
}

// TestDeleteLocalSID tests all cases where configurator's DeleteLocalSID is used (except of complicated cases involving multiple configurator methods)
func TestDeleteLocalSID(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name     string
		Verify   func(error, *vppcallfake.SRv6Calls)
		FailIn   interface{}
		FailWith error
	}{
		{
			Name: "simple deletion of local sid",
			Verify: func(err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				Expect(fakeVPPCalls.LocalSIDState()).To(BeEmpty())
			},
		},
		{
			Name:     "failure propagation from VPPCall's DeleteLocalSid",
			FailIn:   vppcallfake.DeleteLocalSidFuncCall{},
			FailWith: fmt.Errorf(errorMessage),
			Verify: func(err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { //wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				// setup
				configurator, fakeVPPCalls, connection := srv6TestSetup(t)
				defer srv6TestTeardown(connection, configurator)
				sid := sidA
				data := localSID()
				configurator.AddLocalSID(sid, data)
				if td.FailIn != nil {
					fakeVPPCalls.FailIn(td.FailIn, td.FailWith)
				}
				// run tested method and verify
				err := configurator.DeleteLocalSID(sid, data)
				td.Verify(err, fakeVPPCalls)
			}()
		})
	}
}

// TestModifyLocalSID tests all cases where configurator's ModifyLocalSID is used (except of complicated cases involving multiple configurator methods)
func TestModifyLocalSID(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name     string
		Verify   func(srv6.SID, *srv6.LocalSID, *srv6.LocalSID, error, *vppcallfake.SRv6Calls)
		FailIn   interface{}
		FailWith error
	}{
		{
			Name: "simple modify of local sid",
			Verify: func(sid srv6.SID, data *srv6.LocalSID, prevData *srv6.LocalSID, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				state := fakeVPPCalls.LocalSIDState()
				recordedData, exists := state[sid.String()]
				Expect(exists).To(BeTrue())
				Expect(recordedData).To(Equal(data))
			},
		},
		{
			Name:     "failure propagation from VPPCall's AddLocalSid",
			FailIn:   vppcallfake.AddLocalSidFuncCall{},
			FailWith: fmt.Errorf(errorMessage),
			Verify: func(sid srv6.SID, data *srv6.LocalSID, prevData *srv6.LocalSID, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
		{
			Name:     "failure propagation from VPPCall's DeleteLocalSid",
			FailIn:   vppcallfake.DeleteLocalSidFuncCall{},
			FailWith: fmt.Errorf(errorMessage),
			Verify: func(sid srv6.SID, data *srv6.LocalSID, prevData *srv6.LocalSID, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { //wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				// setup and teardown
				configurator, fakeVPPCalls, connection := srv6TestSetup(t)
				defer srv6TestTeardown(connection, configurator)
				// data
				sid := sidA
				prevData := &srv6.LocalSID{
					FibTableID: 0,
					BaseEndFunction: &srv6.LocalSID_End{
						Psp: true,
					},
				}
				data := &srv6.LocalSID{
					FibTableID: 1,
					BaseEndFunction: &srv6.LocalSID_End{
						Psp: false,
					},
				}
				// state and failure setup
				configurator.AddLocalSID(sid, prevData)
				if td.FailIn != nil {
					fakeVPPCalls.FailIn(td.FailIn, td.FailWith)
				}
				// run tested method and verify
				err := configurator.ModifyLocalSID(sid, data, prevData)
				td.Verify(sid, data, prevData, err, fakeVPPCalls)
			}()
		})
	}
}

// TestAddPolicy tests all cases where configurator's AddPolicy and AddPolicySegment is used (except of complicated cases involving other configurator methods)
func TestAddPolicy(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name                              string
		VerifyAfterAddPolicy              func(srv6.SID, *srv6.Policy, *srv6.PolicySegment, *srv6.PolicySegment, error, *vppcallfake.SRv6Calls)
		VerifyAfterFirstAddPolicySegment  func(srv6.SID, *srv6.Policy, *srv6.PolicySegment, *srv6.PolicySegment, error, *vppcallfake.SRv6Calls)
		VerifyAfterSecondAddPolicySegment func(srv6.SID, *srv6.Policy, *srv6.PolicySegment, *srv6.PolicySegment, error, *vppcallfake.SRv6Calls)
		FailIn                            interface{}
		FailWith                          error
		SetPolicySegmentsFirst            bool
	}{
		{
			Name: "add policy and add 2 segment", // handling of first segment is special -> adding 2 segments
			VerifyAfterAddPolicy: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				Expect(fakeVPPCalls.PoliciesState()).To(BeEmpty())
			},
			VerifyAfterFirstAddPolicySegment: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				verifyOnePolicyWithSegments(fakeVPPCalls, bsid, policy, segment)
			},
			VerifyAfterSecondAddPolicySegment: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				verifyOnePolicyWithSegments(fakeVPPCalls, bsid, policy, segment, segment2)
			},
		},
		{
			Name: "add 2 segments to nonexisting policy and add policy", // handling of first segment is special -> adding 2 segments
			SetPolicySegmentsFirst: true,
			VerifyAfterFirstAddPolicySegment: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				Expect(fakeVPPCalls.PoliciesState()).To(HaveLen(0))
			},
			VerifyAfterSecondAddPolicySegment: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				Expect(fakeVPPCalls.PoliciesState()).To(HaveLen(0))
			},
			VerifyAfterAddPolicy: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				verifyOnePolicyWithSegments(fakeVPPCalls, bsid, policy, segment, segment2)
			},
		},
		{
			Name:     "failure propagation from VPPCall's AddPolicy",
			FailIn:   vppcallfake.AddPolicyFuncCall{},
			FailWith: fmt.Errorf(errorMessage),
			VerifyAfterFirstAddPolicySegment: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
		{
			Name:                   "failure propagation from VPPCall's AddPolicySegment",
			FailIn:                 vppcallfake.AddPolicySegmentFuncCall{},
			FailWith:               fmt.Errorf(errorMessage),
			SetPolicySegmentsFirst: true,
			VerifyAfterAddPolicy: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { //wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				// setup and teardown
				configurator, fakeVPPCalls, connection := srv6TestSetup(t)
				defer srv6TestTeardown(connection, configurator)
				// Data
				policy := policy()
				segment := policySegment(1, sidB, sidC, sidD)
				segment2 := policySegment(1, sidA, sidB, sidC)
				// failure setup
				if td.FailIn != nil {
					fakeVPPCalls.FailIn(td.FailIn, td.FailWith)
				}
				// run tested methods and verification after each of them
				if td.SetPolicySegmentsFirst {
					err := configurator.AddPolicySegment(sidA, segmentName1, segment)
					if td.VerifyAfterFirstAddPolicySegment != nil {
						td.VerifyAfterFirstAddPolicySegment(sidA, policy, segment, segment2, err, fakeVPPCalls)
					}
					err = configurator.AddPolicySegment(sidA, segmentName2, segment2)
					if td.VerifyAfterSecondAddPolicySegment != nil {
						td.VerifyAfterSecondAddPolicySegment(sidA, policy, segment, segment2, err, fakeVPPCalls)
					}
					err = configurator.AddPolicy(sidA, policy)
					if td.VerifyAfterAddPolicy != nil {
						td.VerifyAfterAddPolicy(sidA, policy, segment, segment2, err, fakeVPPCalls)
					}
				} else {
					err := configurator.AddPolicy(sidA, policy)
					if td.VerifyAfterAddPolicy != nil {
						td.VerifyAfterAddPolicy(sidA, policy, segment, segment2, err, fakeVPPCalls)
					}
					err = configurator.AddPolicySegment(sidA, segmentName1, segment)
					if td.VerifyAfterFirstAddPolicySegment != nil {
						td.VerifyAfterFirstAddPolicySegment(sidA, policy, segment, segment2, err, fakeVPPCalls)
					}
					err = configurator.AddPolicySegment(sidA, segmentName2, segment2)
					if td.VerifyAfterSecondAddPolicySegment != nil {
						td.VerifyAfterSecondAddPolicySegment(sidA, policy, segment, segment2, err, fakeVPPCalls)
					}
				}
			}()
		})
	}
}

// TestDeletePolicy tests all cases where configurator's DeletePolicy and DeletePolicySegment is used (except of complicated cases involving other configurator methods)
func TestDeletePolicy(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name                                 string
		VerifyAfterRemovePolicy              func(srv6.SID, *srv6.Policy, *srv6.PolicySegment, *srv6.PolicySegment, error, *vppcallfake.SRv6Calls)
		VerifyAfterFirstRemovePolicySegment  func(srv6.SID, *srv6.Policy, *srv6.PolicySegment, *srv6.PolicySegment, error, *vppcallfake.SRv6Calls)
		VerifyAfterSecondRemovePolicySegment func(srv6.SID, *srv6.Policy, *srv6.PolicySegment, *srv6.PolicySegment, error, *vppcallfake.SRv6Calls)
		FailIn                               interface{}
		FailWith                             error
		RemovePoliceSegment                  bool
	}{
		{
			Name: "remove policy (without removing segments)",
			VerifyAfterRemovePolicy: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				Expect(fakeVPPCalls.PoliciesState()).To(BeEmpty())
			},
		},
		{
			Name:                "remove segments and remove policy",
			RemovePoliceSegment: true,
			VerifyAfterFirstRemovePolicySegment: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				Expect(fakeVPPCalls.PoliciesState()).ToNot(BeEmpty())
			},
			VerifyAfterSecondRemovePolicySegment: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				Expect(fakeVPPCalls.PoliciesState()).ToNot(BeEmpty())
			},
			VerifyAfterRemovePolicy: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				Expect(fakeVPPCalls.PoliciesState()).To(BeEmpty())
			},
		},
		{
			Name:     "failure propagation from VPPCall's DeletePolicy",
			FailIn:   vppcallfake.DeletePolicyFuncCall{},
			FailWith: fmt.Errorf(errorMessage),
			VerifyAfterRemovePolicy: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
		{
			Name:                "failure propagation from VPPCall's DeletePolicySegment",
			FailIn:              vppcallfake.DeletePolicySegmentFuncCall{},
			FailWith:            fmt.Errorf(errorMessage),
			RemovePoliceSegment: true,
			VerifyAfterFirstRemovePolicySegment: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { //wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				// setup and teardown
				configurator, fakeVPPCalls, connection := srv6TestSetup(t)
				defer srv6TestTeardown(connection, configurator)
				// Data
				policy := policy()
				segment := policySegment(1, sidB, sidC, sidD)
				segment2 := policySegment(1, sidA, sidB, sidC)
				configurator.AddPolicy(sidA, policy)
				configurator.AddPolicySegment(sidA, segmentName1, segment)
				configurator.AddPolicySegment(sidA, segmentName2, segment2) // handling of first segment is special -> adding 2 segments
				// failure setup
				if td.FailIn != nil {
					fakeVPPCalls.FailIn(td.FailIn, td.FailWith)
				}
				// run tested methods and verification after each of them
				if td.RemovePoliceSegment {
					err := configurator.RemovePolicySegment(sidA, segmentName1, segment)
					if td.VerifyAfterFirstRemovePolicySegment != nil {
						td.VerifyAfterFirstRemovePolicySegment(sidA, policy, segment, segment2, err, fakeVPPCalls)
					}
					err = configurator.RemovePolicySegment(sidA, segmentName2, segment2)
					if td.VerifyAfterSecondRemovePolicySegment != nil {
						td.VerifyAfterSecondRemovePolicySegment(sidA, policy, segment, segment2, err, fakeVPPCalls)
					}
				}
				err := configurator.RemovePolicy(sidA, policy)
				if td.VerifyAfterRemovePolicy != nil {
					td.VerifyAfterRemovePolicy(sidA, policy, segment, segment2, err, fakeVPPCalls)
				}
			}()
		})
	}
}

// TestModifyPolicy tests all cases where configurator's ModifyPolicy is used (except of complicated cases involving other configurator methods)
func TestModifyPolicy(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name     string
		Verify   func(srv6.SID, *srv6.Policy, *srv6.Policy, *srv6.PolicySegment, error, *vppcallfake.SRv6Calls)
		FailIn   interface{}
		FailWith error
	}{
		{
			Name: "policy attributes modification",
			Verify: func(bsid srv6.SID, policy *srv6.Policy, prevPolicy *srv6.Policy, segment *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				verifyOnePolicyWithSegments(fakeVPPCalls, bsid, policy, segment)
			},
		},
		{
			Name:     "failure propagation from VPPCall's AddPolicy",
			FailIn:   vppcallfake.AddPolicyFuncCall{},
			FailWith: fmt.Errorf(errorMessage),
			Verify: func(bsid srv6.SID, policy *srv6.Policy, prevPolicy *srv6.Policy, segment *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
		{
			Name:     "failure propagation from VPPCall's DeletePolicy",
			FailIn:   vppcallfake.DeletePolicyFuncCall{},
			FailWith: fmt.Errorf(errorMessage),
			Verify: func(bsid srv6.SID, policy *srv6.Policy, prevPolicy *srv6.Policy, segment *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { //wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				// setup and teardown
				configurator, fakeVPPCalls, connection := srv6TestSetup(t)
				defer srv6TestTeardown(connection, configurator)
				// Data
				prevPolicy := &srv6.Policy{
					FibTableID:       0,
					SprayBehaviour:   true,
					SrhEncapsulation: true,
				}
				policy := &srv6.Policy{
					FibTableID:       1,
					SprayBehaviour:   false,
					SrhEncapsulation: false,
				}
				segment := policySegment(1, sidB, sidC, sidD)
				configurator.AddPolicy(sidA, prevPolicy)
				configurator.AddPolicySegment(sidA, segmentName1, segment)
				// failure setup
				if td.FailIn != nil {
					fakeVPPCalls.FailIn(td.FailIn, td.FailWith)
				}
				// run tested methods and verification after each of them
				err := configurator.ModifyPolicy(sidA, policy, prevPolicy)
				if td.Verify != nil {
					td.Verify(sidA, policy, prevPolicy, segment, err, fakeVPPCalls)
				}
			}()
		})
	}
}

// TestModifyPolicySegment tests all cases where configurator's ModifyPolicySegment is used (except of complicated cases involving other configurator methods)
func TestModifyPolicySegment(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name           string
		Verify         func(srv6.SID, *srv6.Policy, *srv6.PolicySegment, *srv6.PolicySegment, *srv6.PolicySegment, error, *vppcallfake.SRv6Calls)
		FailIn         interface{}
		FailWith       error
		OnlyOneSegment bool
	}{
		{
			Name: "policy segment modification (non-last segment)", // last segment is handled differently
			Verify: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, prevSegment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				verifyOnePolicyWithSegments(fakeVPPCalls, bsid, policy, segment2, segment)
			},
		},
		{
			Name:           "policy segment modification (last segment)", // last segment is handled differently
			OnlyOneSegment: true,
			Verify: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, prevSegment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				verifyOnePolicyWithSegments(fakeVPPCalls, bsid, policy, segment)
			},
		},
		{
			Name:           "failure propagation from VPPCall's AddPolicy",
			OnlyOneSegment: true,
			FailIn:         vppcallfake.AddPolicyFuncCall{},
			FailWith:       fmt.Errorf(errorMessage),
			Verify: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, prevSegment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
		{
			Name:           "failure propagation from VPPCall's DeletePolicy",
			OnlyOneSegment: true,
			FailIn:         vppcallfake.DeletePolicyFuncCall{},
			FailWith:       fmt.Errorf(errorMessage),
			Verify: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, prevSegment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
		{
			Name:     "failure propagation from VPPCall's DeletePolicySegment",
			FailIn:   vppcallfake.DeletePolicySegmentFuncCall{},
			FailWith: fmt.Errorf(errorMessage),
			Verify: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, prevSegment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
		{
			Name:     "failure propagation from VPPCall's AddPolicySegment",
			FailIn:   vppcallfake.AddPolicySegmentFuncCall{},
			FailWith: fmt.Errorf(errorMessage),
			Verify: func(bsid srv6.SID, policy *srv6.Policy, segment *srv6.PolicySegment, prevSegment *srv6.PolicySegment, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { //wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				// setup and teardown
				configurator, fakeVPPCalls, connection := srv6TestSetup(t)
				defer srv6TestTeardown(connection, configurator)
				// Data
				policy := policy()
				prevSegment := policySegment(0, sidA, sidB, sidC)
				segment := policySegment(1, sidB, sidC, sidD)
				segment2 := policySegment(2, sidC, sidD, sidA)
				configurator.AddPolicy(sidA, policy)
				configurator.AddPolicySegment(sidA, segmentName1, prevSegment)
				if !td.OnlyOneSegment {
					configurator.AddPolicySegment(sidA, segmentName2, segment2)
				}
				// failure setup
				if td.FailIn != nil {
					fakeVPPCalls.FailIn(td.FailIn, td.FailWith)
				}
				// run tested methods and verification after each of them
				err := configurator.ModifyPolicySegment(sidA, segmentName1, segment, prevSegment)
				if td.Verify != nil {
					td.Verify(sidA, policy, segment, prevSegment, segment2, err, fakeVPPCalls)
				}
			}()
		})
	}
}

// TestFillingAlreadyCreatedSegmentEmptyPolicy tests cases where policy is created, but cleaned off segments and
// new segment is added. This test is testing special case around last segment in policy.
func TestFillingAlreadyCreatedSegmentEmptyPolicy(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name     string
		Verify   func(srv6.SID, *srv6.Policy, *srv6.PolicySegment, error, *vppcallfake.SRv6Calls)
		FailIn   interface{}
		FailWith error
	}{
		{
			Name: "all segments removal and adding new onw", // last segment is handled differently
			Verify: func(bsid srv6.SID, policy *srv6.Policy, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				verifyOnePolicyWithSegments(fakeVPPCalls, bsid, policy, segment2)
			},
		},
		{
			Name:     "failure propagation from VPPCall's DeletePolicy",
			FailIn:   vppcallfake.DeletePolicyFuncCall{},
			FailWith: fmt.Errorf(errorMessage),
			Verify: func(bsid srv6.SID, policy *srv6.Policy, segment2 *srv6.PolicySegment, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
	}
	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { //wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				// setup and teardown
				configurator, fakeVPPCalls, connection := srv6TestSetup(t)
				defer srv6TestTeardown(connection, configurator)
				// Data
				policy := policy()
				segment := policySegment(0, sidA, sidB, sidC)
				segment2 := policySegment(1, sidB, sidC, sidD)
				// case building
				Expect(configurator.AddPolicy(sidA, policy)).To(BeNil())
				Expect(configurator.AddPolicySegment(sidA, segmentName1, segment)).To(BeNil())
				Expect(configurator.RemovePolicySegment(sidA, segmentName1, segment)).To(BeNil())
				// failure setup
				if td.FailIn != nil {
					fakeVPPCalls.FailIn(td.FailIn, td.FailWith)
				}
				// run tested methods and verification after each of them
				err := configurator.AddPolicySegment(sidA, segmentName2, segment2)
				td.Verify(sidA, policy, segment2, err, fakeVPPCalls)
			}()
		})
	}
}

// TestAddSteering tests all cases where configurator's AddSteering is used
func TestAddSteering(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name                   string
		VerifyAfterAddPolicy   func(*srv6.Steering, *vppcallfake.SRv6Calls)
		VerifyAfterAddSteering func(*srv6.Steering, error, *vppcallfake.SRv6Calls)
		FailIn                 interface{}
		FailWith               error
		ReferencePolicyByIndex bool
		CreatePolicyAfter      bool
		CustomSteeringData     *srv6.Steering
	}{
		{
			Name: "addition of steering (with already existing BSID-referenced policy)",
			VerifyAfterAddPolicy: func(steering *srv6.Steering, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(fakeVPPCalls.SteeringState()).To(BeEmpty())
			},
			VerifyAfterAddSteering: func(steering *srv6.Steering, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				state := fakeVPPCalls.SteeringState()
				_, exists := state[*steering]
				Expect(exists).To(BeTrue())
			},
		},
		{
			Name:              "addition of steering (with BSID-referenced policy added later)",
			CreatePolicyAfter: true,
			VerifyAfterAddSteering: func(steering *srv6.Steering, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				Expect(fakeVPPCalls.SteeringState()).To(BeEmpty())
			},
			VerifyAfterAddPolicy: func(steering *srv6.Steering, fakeVPPCalls *vppcallfake.SRv6Calls) {
				state := fakeVPPCalls.SteeringState()
				_, exists := state[*steering]
				Expect(exists).To(BeTrue())
			},
		},
		{
			Name: "addition of steering (with already existing Index-referenced policy)",
			ReferencePolicyByIndex: true,
			VerifyAfterAddPolicy: func(steering *srv6.Steering, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(fakeVPPCalls.SteeringState()).To(BeEmpty())
			},
			VerifyAfterAddSteering: func(steering *srv6.Steering, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				state := fakeVPPCalls.SteeringState()
				_, exists := state[*steering]
				Expect(exists).To(BeTrue())
			},
		},
		{
			Name: "addition of steering (with Index-referenced policy added later)",
			ReferencePolicyByIndex: true,
			CreatePolicyAfter:      true,
			VerifyAfterAddSteering: func(steering *srv6.Steering, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				Expect(fakeVPPCalls.SteeringState()).To(BeEmpty())
			},
			VerifyAfterAddPolicy: func(steering *srv6.Steering, fakeVPPCalls *vppcallfake.SRv6Calls) {
				state := fakeVPPCalls.SteeringState()
				_, exists := state[*steering]
				Expect(exists).To(BeTrue())
			},
		},
		{
			Name:               "invalid BSID as policy reference",
			CustomSteeringData: steeringWithPolicyBsidRef("XYZ"), // valid binding sid = valid IPv6
			VerifyAfterAddSteering: func(steering *srv6.Steering, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
			},
		},
		{
			Name:     "failure propagation from VPPCall's AddSteering",
			FailIn:   vppcallfake.AddSteeringFuncCall{},
			FailWith: fmt.Errorf(errorMessage),
			VerifyAfterAddSteering: func(steering *srv6.Steering, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { //wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				configurator, fakeVPPCalls, connection := srv6TestSetup(t)
				defer srv6TestTeardown(connection, configurator)
				// data
				policy := policy()
				segment := policySegment(1, sidB, sidC, sidD)
				steering := steeringWithPolicyBsidRef(sidA.String())
				if td.ReferencePolicyByIndex {
					steering = steeringWithPolicyIndexRef(0)
				}
				if td.CustomSteeringData != nil {
					steering = td.CustomSteeringData
				}
				// failure setup
				if td.FailIn != nil {
					fakeVPPCalls.FailIn(td.FailIn, td.FailWith)
				}
				// case building
				if td.CreatePolicyAfter {
					err := configurator.AddSteering(steeringName, steering)
					if td.VerifyAfterAddSteering != nil {
						td.VerifyAfterAddSteering(steering, err, fakeVPPCalls)
					}
					configurator.AddPolicy(sidA, policy)
					configurator.AddPolicySegment(sidA, segmentName1, segment)
					if td.VerifyAfterAddPolicy != nil {
						td.VerifyAfterAddPolicy(steering, fakeVPPCalls)
					}
				} else {
					configurator.AddPolicy(sidA, policy)
					configurator.AddPolicySegment(sidA, segmentName1, segment)
					if td.VerifyAfterAddPolicy != nil {
						td.VerifyAfterAddPolicy(steering, fakeVPPCalls)
					}
					err := configurator.AddSteering(steeringName, steering)
					if td.VerifyAfterAddSteering != nil {
						td.VerifyAfterAddSteering(steering, err, fakeVPPCalls)
					}
				}
			}()
		})
	}
}

// TestRemoveSteering tests all cases where configurator's RemoveSteering is used (except of complicated cases involving multiple configurator methods)
func TestRemoveSteering(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name     string
		Verify   func(error, *vppcallfake.SRv6Calls)
		FailIn   interface{}
		FailWith error
	}{
		{
			Name: "simple steering removal",
			Verify: func(err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				Expect(fakeVPPCalls.SteeringState()).To(BeEmpty())
			},
		},
		{
			Name:     "failure propagation from VPPCall's RemoveSteering",
			FailIn:   vppcallfake.RemoveSteeringFuncCall{},
			FailWith: fmt.Errorf(errorMessage),
			Verify: func(err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { //wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				// setup
				configurator, fakeVPPCalls, connection := srv6TestSetup(t)
				defer srv6TestTeardown(connection, configurator)
				// data
				policy := policy()
				segment := policySegment(1, sidB, sidC, sidD)
				steering := steeringWithPolicyBsidRef(sidA.String())
				// case building
				configurator.AddPolicy(sidA, policy)
				configurator.AddPolicySegment(sidA, segmentName1, segment)
				configurator.AddSteering(steeringName, steering)
				// failure setup
				if td.FailIn != nil {
					fakeVPPCalls.FailIn(td.FailIn, td.FailWith)
				}
				// run tested method and verify
				err := configurator.RemoveSteering(steeringName, steering)
				td.Verify(err, fakeVPPCalls)
			}()
		})
	}
}

// TestModifySteering tests all cases where configurator's ModifySteering is used (except of complicated cases involving multiple configurator methods)
func TestModifySteering(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name     string
		Verify   func(*srv6.Steering, error, *vppcallfake.SRv6Calls)
		FailIn   interface{}
		FailWith error
	}{
		{
			Name: "simple modification of steering",
			Verify: func(steering *srv6.Steering, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).To(BeNil())
				state := fakeVPPCalls.SteeringState()
				_, exists := state[*steering]
				Expect(exists).To(BeTrue())
			},
		},
		{
			Name:     "failure propagation from VPPCall's AddSteering",
			FailIn:   vppcallfake.AddSteeringFuncCall{},
			FailWith: fmt.Errorf(errorMessage),
			Verify: func(steering *srv6.Steering, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
		{
			Name:     "failure propagation from VPPCall's RemoveSteering",
			FailIn:   vppcallfake.RemoveSteeringFuncCall{},
			FailWith: fmt.Errorf(errorMessage),
			Verify: func(steering *srv6.Steering, err error, fakeVPPCalls *vppcallfake.SRv6Calls) {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring(errorMessage))
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { //wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				// setup and teardown
				configurator, fakeVPPCalls, connection := srv6TestSetup(t)
				defer srv6TestTeardown(connection, configurator)
				// data
				bsid := sidA
				policy := policy()
				segment := policySegment(1, sidB, sidC, sidD)
				prevData := &srv6.Steering{
					PolicyBSID: bsid.String(),
					L3Traffic: &srv6.Steering_L3Traffic{
						FibTableID:    0,
						PrefixAddress: "A::",
					},
				}
				data := &srv6.Steering{
					PolicyBSID: bsid.String(),
					L3Traffic: &srv6.Steering_L3Traffic{
						FibTableID:    1,
						PrefixAddress: "B::",
					},
				}
				// case building
				configurator.AddPolicy(bsid, policy)
				configurator.AddPolicySegment(bsid, segmentName1, segment)
				configurator.AddSteering(steeringName, prevData)
				// failure setup
				if td.FailIn != nil {
					fakeVPPCalls.FailIn(td.FailIn, td.FailWith)
				}
				// run tested method and verify
				err := configurator.ModifySteering(steeringName, data, prevData)
				td.Verify(data, err, fakeVPPCalls)
			}()
		})
	}
}

/* Srv6 Test Setup */

func srv6TestSetup(t *testing.T) (*srplugin.SRv6Configurator, *vppcallfake.SRv6Calls, *core.Connection) {
	RegisterTestingT(t)
	// connection
	ctx := &vppcallmock.TestCtx{
		MockVpp: &mock.VppAdapter{},
	}
	connection, err := core.Connect(ctx.MockVpp)
	Expect(err).ShouldNot(HaveOccurred())
	// Logger
	log := logging.ForPlugin("test-log", logrus.NewLogRegistry())
	log.SetLevel(logging.DebugLevel)
	// Interface index from default plugins
	swIndex := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(log, "test-srv6",
		"sw_if_indexes", ifaceidx.IndexMetadata))
	// Configurator
	fakeVPPCalls := vppcallfake.NewSRv6Calls()
	stopwatch := measure.NewStopwatch("SRConfigurator-Test", log)
	configurator := &srplugin.SRv6Configurator{
		Log:         log,
		GoVppmux:    connection,
		SwIfIndexes: swIndex,
		VppCalls:    fakeVPPCalls,
		Stopwatch:   stopwatch,
	}
	err = configurator.Init()
	Expect(err).To(BeNil())

	return configurator, fakeVPPCalls, connection
}

/* Srv6 Test Teardown */

func srv6TestTeardown(connection *core.Connection, plugin *srplugin.SRv6Configurator) {
	connection.Disconnect()
	err := plugin.Close()
	Expect(err).To(BeNil())
}

func verifyOnePolicyWithSegments(fakeVPPCalls *vppcallfake.SRv6Calls, bsid srv6.SID, policy *srv6.Policy, segments ...*srv6.PolicySegment) {
	policiesState := fakeVPPCalls.PoliciesState()
	Expect(policiesState).To(HaveLen(1))
	policyState, exists := policiesState[bsid.String()]
	Expect(exists).To(BeTrue())
	Expect(policyState.Policy()).To(Equal(policy))
	Expect(policyState.Segments()).To(HaveLen(len(segments)))
	intersection := 0
	for _, actualSegment := range policyState.Segments() {
		for _, expectedSegment := range segments {
			if actualSegment == expectedSegment {
				intersection++
			}
		}
	}
	Expect(intersection).To(BeEquivalentTo(len(segments)), "policy have exactly the same segments as expected")
}

func sid(str string) srv6.SID {
	bsid, err := srplugin.ParseIPv6(str)
	if err != nil {
		panic(fmt.Sprintf("can't parse \"%v\" into SRv6 BSID (IPv6 address)", str))
	}
	return bsid
}

func localSID() *srv6.LocalSID {
	return &srv6.LocalSID{
		FibTableID: 0,
		BaseEndFunction: &srv6.LocalSID_End{
			Psp: true,
		},
	}
}

func policy() *srv6.Policy {
	return &srv6.Policy{
		FibTableID:       0,
		SprayBehaviour:   true,
		SrhEncapsulation: true,
	}
}

func policySegment(weight uint32, sids ...srv6.SID) *srv6.PolicySegment {
	segments := make([]string, len(sids))
	for i, sid := range sids {
		segments[i] = sid.String()
	}

	return &srv6.PolicySegment{
		Weight:   weight,
		Segments: segments,
	}
}

func steeringWithPolicyBsidRef(bsid string) *srv6.Steering {
	return steeringRef(bsid, 0)
}

func steeringWithPolicyIndexRef(index uint32) *srv6.Steering {
	return steeringRef("", index)
}

func steeringRef(bsid string, index uint32) *srv6.Steering {
	return &srv6.Steering{
		PolicyBSID:  bsid,
		PolicyIndex: index,
		L3Traffic: &srv6.Steering_L3Traffic{
			FibTableID:    0,
			PrefixAddress: "A::",
		},
	}
}
