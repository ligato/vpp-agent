//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// +build proxy

package vpp_test

import (
	"context"
	"encoding/gob"
	"io"
	"log"
	"testing"

	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/proxy"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/interfaces"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/vpe"
)

func init() {
	// this is required for proxy client encoder to work properly
	for _, msg := range vpp1908.Messages.AllMessages() {
		gob.Register(msg)
	}
}

// This test demonstrates how to use proxy to access VPP binapi and stats
// remotely via vpp-agent HTTP server. Run this test with:
//
// 	go test -v -tags proxy ./plugins/govppmux/vppcalls
//
func TestProxyClient(t *testing.T) {
	// connect to proxy server
	client, err := proxy.Connect(":9191")
	if err != nil {
		log.Fatalln("connecting to proxy failed:", err)
	}

	// proxy stats
	statsProvider, err := client.NewStatsClient()
	if err != nil {
		log.Fatalln(err)
	}

	var sysStats api.SystemStats
	if err := statsProvider.GetSystemStats(&sysStats); err != nil {
		log.Fatalln("getting stats failed:", err)
	}
	log.Printf("SystemStats: %+v", sysStats)

	// proxy binapi
	binapiChannel, err := client.NewBinapiClient()
	if err != nil {
		log.Fatalln(err)
	}

	var msgs []api.Message
	msgs = append(msgs, interfaces.AllMessages()...)
	msgs = append(msgs, vpe.AllMessages()...)
	if err := binapiChannel.CheckCompatiblity(msgs...); err != nil {
		panic(err)
	}
	log.Println("compatibility OK!")

	// - using binapi message directly
	req := &vpe.CliInband{Cmd: "show version"}
	reply := new(vpe.CliInbandReply)
	if err := binapiChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		log.Fatalln("binapi request failed:", err)
	}
	log.Printf("VPP version: %+v", reply.Reply)

	// - or using generated rpc service
	svc := interfaces.NewServiceClient(binapiChannel)

	stream, err := svc.DumpSwInterface(context.Background(), &interfaces.SwInterfaceDump{})
	if err != nil {
		log.Fatalln("binapi request failed:", err)
	}
	log.Printf("dumping interfaces")
	for {
		iface, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("- interface %d: %v", iface.SwIfIndex, iface.InterfaceName)
	}
}
