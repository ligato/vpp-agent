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

// Binary simple-client is an example VPP management application that exercises the
// govpp API on real-world use-cases.
package main

// Generates Go bindings for all VPP APIs located in the json directory.
//go:generate binapi-generator --input-dir=bin_api --output-dir=bin_api

import (
	"fmt"
	"net"
	"os"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/examples/bin_api/acl"
	"git.fd.io/govpp.git/examples/bin_api/interfaces"
	"git.fd.io/govpp.git/examples/bin_api/tap"
)

func main() {
	fmt.Println("Starting simple VPP client...")

	// connect to VPP
	conn, err := govpp.Connect()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer conn.Disconnect()

	// create an API channel that will be used in the examples
	ch, err := conn.NewAPIChannel()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer ch.Close()

	// check whether the VPP supports our version of some messages
	compatibilityCheck(ch)

	// individual examples
	aclVersion(ch)
	aclConfig(ch)
	aclDump(ch)

	tapConnect(ch)

	interfaceDump(ch)
	interfaceNotifications(ch)
}

// compatibilityCheck shows how an management application can check whether generated API messages are
// compatible with the version of VPP which the library is connected to.
func compatibilityCheck(ch *api.Channel) {
	err := ch.CheckMessageCompatibility(
		&interfaces.SwInterfaceDump{},
		&interfaces.SwInterfaceDetails{},
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// aclVersion is the simplest API example - one empty request message and one reply message.
func aclVersion(ch *api.Channel) {
	req := &acl.ACLPluginGetVersion{}
	reply := &acl.ACLPluginGetVersionReply{}

	err := ch.SendRequest(req).ReceiveReply(reply)

	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("%+v\n", reply)
	}
}

// aclConfig is another simple API example - in this case, the request contains structured data.
func aclConfig(ch *api.Channel) {
	req := &acl.ACLAddReplace{
		ACLIndex: ^uint32(0),
		Tag:      []byte("access list 1"),
		R: []acl.ACLRule{
			{
				IsPermit:       1,
				SrcIPAddr:      net.ParseIP("10.0.0.0").To4(),
				SrcIPPrefixLen: 8,
				DstIPAddr:      net.ParseIP("192.168.1.0").To4(),
				DstIPPrefixLen: 24,
				Proto:          6,
			},
			{
				IsPermit:       1,
				SrcIPAddr:      net.ParseIP("8.8.8.8").To4(),
				SrcIPPrefixLen: 32,
				DstIPAddr:      net.ParseIP("172.16.0.0").To4(),
				DstIPPrefixLen: 16,
				Proto:          6,
			},
		},
	}
	reply := &acl.ACLAddReplaceReply{}

	err := ch.SendRequest(req).ReceiveReply(reply)

	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("%+v\n", reply)
	}
}

// aclDump shows an example where SendRequest and ReceiveReply are not chained together.
func aclDump(ch *api.Channel) {
	req := &acl.ACLDump{}
	reply := &acl.ACLDetails{}

	reqCtx := ch.SendRequest(req)
	err := reqCtx.ReceiveReply(reply)

	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("%+v\n", reply)
	}
}

// tapConnect example shows how the Go channels in the API channel can be accessed directly instead
// of using SendRequest and ReceiveReply wrappers.
func tapConnect(ch *api.Channel) {
	req := &tap.TapConnect{
		TapName:      []byte("testtap"),
		UseRandomMac: 1,
	}

	// send the request to the request go channel
	ch.ReqChan <- &api.VppRequest{Message: req}

	// receive a reply from the reply go channel
	vppReply := <-ch.ReplyChan
	if vppReply.Error != nil {
		fmt.Println("Error:", vppReply.Error)
		return
	}

	// decode the message
	reply := &tap.TapConnectReply{}
	err := ch.MsgDecoder.DecodeMsg(vppReply.Data, reply)

	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("%+v\n", reply)
	}
}

// interfaceDump shows an example of multipart request (multiple replies are expected).
func interfaceDump(ch *api.Channel) {
	req := &interfaces.SwInterfaceDump{}
	reqCtx := ch.SendMultiRequest(req)

	for {
		msg := &interfaces.SwInterfaceDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break // break out of the loop
		}
		if err != nil {
			fmt.Println("Error:", err)
		}
		fmt.Printf("%+v\n", msg)
	}
}

// interfaceNotifications shows the usage of notification API. Note that for notifications,
// you are supposed to create your own Go channel with your preferred buffer size. If the channel's
// buffer is full, the notifications will not be delivered into it.
func interfaceNotifications(ch *api.Channel) {
	// subscribe for specific notification message
	notifChan := make(chan api.Message, 100)
	subs, _ := ch.SubscribeNotification(notifChan, interfaces.NewSwInterfaceSetFlags)

	// enable interface events in VPP
	ch.SendRequest(&interfaces.WantInterfaceEvents{
		Pid:           uint32(os.Getpid()),
		EnableDisable: 1,
	}).ReceiveReply(&interfaces.WantInterfaceEventsReply{})

	// generate some events in VPP
	ch.SendRequest(&interfaces.SwInterfaceSetFlags{
		SwIfIndex:   0,
		AdminUpDown: 0,
	}).ReceiveReply(&interfaces.SwInterfaceSetFlagsReply{})
	ch.SendRequest(&interfaces.SwInterfaceSetFlags{
		SwIfIndex:   0,
		AdminUpDown: 1,
	}).ReceiveReply(&interfaces.SwInterfaceSetFlagsReply{})

	// receive one notification
	notif := (<-notifChan).(*interfaces.SwInterfaceSetFlags)
	fmt.Printf("%+v\n", notif)

	// unsubscribe from delivery of the notifications
	ch.UnsubscribeNotification(subs)
}
