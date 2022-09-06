//  Copyright (c) 2021 Cisco and/or its affiliates.
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

// The VPP Proxy example demonstrates how to use GoVPP proxy to access
// VPP binapi and stats API remotely via HTTP server.
package main

import (
	"context"
	"encoding/gob"
	"flag"
	"io"
	"log"

	"go.fd.io/govpp/api"
	"go.fd.io/govpp/proxy"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106"
	interfaces "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/vpe"
)

// VPP version used in the example.
const vppVersion = vpp2106.Version

var (
	address = flag.String("addr", ":9191", "agent address")
)

func main() {
	flag.Parse()

	client, err := proxy.Connect(*address)
	if err != nil {
		log.Fatalln("connecting to proxy failed:", err)
	}

	proxyStats(client)
	proxyBinapi(client)
}

func proxyStats(client *proxy.Client) {
	statsProvider, err := client.NewStatsClient()
	if err != nil {
		log.Fatalln(err)
	}

	var sysStats api.SystemStats
	if err := statsProvider.GetSystemStats(&sysStats); err != nil {
		log.Fatalln("getting stats failed:", err)
	}
	log.Printf("SystemStats: %+v", sysStats)
}

func proxyBinapi(client *proxy.Client) {
	binapiChannel, err := client.NewBinapiClient()
	if err != nil {
		log.Fatalln(err)
	}

	// All binapi messages must be registered to gob
	for _, msg := range binapi.Versions[vppVersion].AllMessages() {
		gob.Register(msg)
	}

	// Check compatibility with remote VPP version
	var msgs []api.Message
	msgs = append(msgs, interfaces.AllMessages()...)
	msgs = append(msgs, vpe.AllMessages()...)
	if err := binapiChannel.CheckCompatiblity(msgs...); err != nil {
		log.Fatalf("compatibility check (VPP %v) failed: %v", vppVersion, err)
	}
	log.Printf("compatibility OK! (VPP %v)", vppVersion)

	var (
		vpeSvc       = vpe.NewServiceClient(binapiChannel)
		interfaceSvc = interfaces.NewServiceClient(binapiChannel)
	)

	// Show VPP version
	version, err := vpeSvc.ShowVersion(context.Background(), &vpe.ShowVersion{})
	if err != nil {
		log.Fatalln("ShowVersion failed:", err)
	}
	log.Printf("Version: %+v", version)

	// List interfaces
	stream, err := interfaceSvc.SwInterfaceDump(context.Background(), &interfaces.SwInterfaceDump{})
	if err != nil {
		log.Fatalln("SwInterfaceDump failed:", err)
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
