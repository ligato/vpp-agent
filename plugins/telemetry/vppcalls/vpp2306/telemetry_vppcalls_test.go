//  Copyright (c) 2022 Cisco and/or its affiliates.
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

package vpp2306_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/plugins/telemetry/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/telemetry/vppcalls/vpp2306"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2306/vlib"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
)

func TestGetBuffers(t *testing.T) {
	ctx, handler := testSetup(t)
	defer ctx.TeardownTestCtx()

	const reply = `Pool Name            Index NUMA  Size  Data Size  Total  Avail  Cached   Used  
default-numa-0         0     0   2304     2048    17290  17290     0       0   `
	ctx.MockVpp.MockReply(&vlib.CliInbandReply{
		Reply: reply,
	})

	info, err := handler.GetBuffersInfo(context.TODO())

	Expect(err).ShouldNot(HaveOccurred())
	Expect(info.Items).To(HaveLen(1))
	Expect(info.Items[0]).To(Equal(vppcalls.BuffersItem{
		//ThreadID: 0,
		Name:  "default-numa-0",
		Index: 0,
		Size:  2304,
		Alloc: 0,
		Free:  17290,
		//NumAlloc: 256,
		//NumFree:  19,
	}))
	/*Expect(info.Items[1]).To(Equal(vppcalls.BuffersItem{
		ThreadID: 0,
		Name:     "lacp-ethernet",
		Index:    1,
		Size:     256,
		Alloc:    1130000,
		Free:     27000,
		NumAlloc: 512,
		NumFree:  12,
	}))
	Expect(info.Items[2]).To(Equal(vppcalls.BuffersItem{
		ThreadID: 0,
		Name:     "marker-ethernet",
		Index:    2,
		Size:     256,
		Alloc:    1110000000,
		Free:     0,
		NumAlloc: 0,
		NumFree:  0,
	}))*/
}

func TestGetRuntime(t *testing.T) {
	tests := []struct {
		name        string
		reply       string
		threadCount int
		itemCount   int
		itemIdx     int
		item        vppcalls.RuntimeItem
	}{
		{
			name: "19.08",
			reply: `Time 84714.7, average vectors/node 0.00, last 128 main loops 0.00 per node 0.00
  vector rates in 0.0000e0, out 0.0000e0, drop 0.0000e0, punt 0.0000e0
             Name                 State         Calls          Vectors        Suspends         Clocks       Vectors/Call  
acl-plugin-fa-cleaner-process  event wait                6               5               1          1.10e4            0.00
api-rx-from-ring                 active                  0               0            7870          8.63e5            0.00
avf-process                    event wait                0               0               1          4.53e3            0.00
bfd-process                    event wait                0               0               1          7.01e3            0.00
bond-process                   event wait                0               0               1          2.95e3            0.00
cdp-process                     any wait                 0               0               1          5.46e3            0.00
dhcp-client-process             any wait                 0               0             847          6.63e3            0.00
dhcp6-client-cp-process         any wait                 0               0               1          1.52e3            0.00
dhcp6-pd-client-cp-process      any wait                 0               0               1          1.71e3            0.00
dhcp6-pd-reply-publisher-proce event wait                0               0               1          9.73e2            0.00
dhcp6-reply-publisher-process  event wait                0               0               1          9.12e2            0.00
dns-resolver-process            any wait                 0               0              85          8.98e3            0.00
fib-walk                        any wait                 0               0           42247          1.08e4            0.00
flow-report-process             any wait                 0               0               1          1.33e3            0.00
flowprobe-timer-process         any wait                 0               0               1          5.18e3            0.00
gbp-scanner                    event wait                0               0               1          5.17e3            0.00
igmp-timer-process             event wait                0               0               1          6.53e3            0.00
ikev2-manager-process           any wait                 0               0           84353          7.84e3            0.00
ioam-export-process             any wait                 0               0               1          1.64e3            0.00
ip-neighbor-scan-process        any wait                 0               0            1412          9.65e3            0.00
ip-route-resolver-process       any wait                 0               0             847          6.12e3            0.00
ip4-reassembly-expire-walk      any wait                 0               0            8464          6.92e3            0.00
ip6-icmp-neighbor-discovery-ev  any wait                 0               0           84353          8.58e3            0.00
ip6-reassembly-expire-walk      any wait                 0               0            8464          6.67e3            0.00
l2fib-mac-age-scanner-process  event wait                0               0               1          1.98e3            0.00
lacp-process                   event wait                0               0               1          1.58e5            0.00
lisp-retry-service              any wait                 0               0           42247          1.08e4            0.00
lldp-process                   event wait                0               0               1          8.76e4            0.00
memif-process                  event wait                0               0               1          9.34e3            0.00
nat-det-expire-walk               done                   1               0               0          2.92e3            0.00
nat-ha-process                 event wait                0               0               1          4.12e3            0.00
nat64-expire-walk              event wait                0               0               1          2.41e3            0.00
nsh-md2-ioam-export-process     any wait                 0               0               1          1.10e4            0.00
perfmon-periodic-process       event wait                0               0               1          3.61e7            0.00
rd-cp-process                   any wait                 0               0               1          1.55e3            0.00
send-dhcp6-client-message-proc  any wait                 0               0               1          2.22e3            0.00
send-dhcp6-pd-client-message-p  any wait                 0               0               1          1.43e3            0.00
send-rs-process                 any wait                 0               0               1          1.49e3            0.00
startup-config-process            done                   1               0               1          5.68e3            0.00
statseg-collector-process       time wait                0               0            8464          2.79e5            0.00
udp-ping-process                any wait                 0               0               1          6.96e3            0.00
unix-cli-127.0.0.1:mdns           done                   2               0               4          2.14e9            0.00
unix-epoll-input                 polling          20325059               0               0          1.13e7            0.00
vhost-user-process              any wait                 0               0               1          3.73e3            0.00
vhost-user-send-interrupt-proc  any wait                 0               0               1          1.28e3            0.00
vpe-link-state-process         event wait                0               0               1          9.63e2            0.00
vpe-oam-process                 any wait                 0               0           41419          9.59e3            0.00
vxlan-gpe-ioam-export-process   any wait                 0               0               1          1.60e3            0.00
wildcard-ip4-arp-publisher-pro event wait                0               0               1          1.44e3            0.00
`,
			threadCount: 1,
			itemCount:   49,
			item: vppcalls.RuntimeItem{
				Name:           "acl-plugin-fa-cleaner-process",
				State:          "event wait",
				Calls:          6,
				Vectors:        5,
				Suspends:       1,
				Clocks:         1.10e4,
				VectorsPerCall: 0,
			},
		},
		{
			name: "one thread",
			reply: `Time 3151.2, average vectors/node 1.00, last 128 main loops 0.00 per node 0.00
  vector rates in 2.8561e-3, out 0.0000e0, drop 4.4428e-3, punt 0.0000e0
             Name                 State         Calls          Vectors        Suspends         Clocks       Vectors/Call     Perf Ticks   
acl-plugin-fa-cleaner-process  event wait                0               0               1          5.14e3            0.00
af-packet-input               interrupt wa               9               9               0          1.55e5            1.00
api-rx-from-ring                any wait                 0               0            4735          4.72e6            0.00
avf-process                    event wait                0               0               1          4.52e3            0.00
bfd-process                    event wait                0               0               1          6.59e3            0.00
bond-process                   event wait                0               0               1          2.07e3            0.00
cdp-process                     any wait                 0               0               1          4.43e3            0.00
dhcp-client-process             any wait                 0               0              32          8.73e3            0.00
dhcp6-client-cp-process         any wait                 0               0               1          1.94e3            0.00
dhcp6-pd-client-cp-process      any wait                 0               0               1          1.73e3            0.00
dhcp6-pd-reply-publisher-proce event wait                0               0               1          1.01e3            0.00
dhcp6-reply-publisher-process  event wait                0               0               1          8.75e2            0.00
dns-resolver-process            any wait                 0               0               4          2.11e4            0.00
error-drop                       active                 14              14               0          1.29e5            1.00
ethernet-input                   active                  9               9               0          6.41e5            1.00
fib-walk                        any wait                 0               0            1571          2.12e4            0.00
flow-report-process             any wait                 0               0               1          1.13e3            0.00
flowprobe-timer-process         any wait                 0               0               1          5.27e3            0.00
gbp-scanner                    event wait                0               0               1          5.36e3            0.00
igmp-timer-process             event wait                0               0               1          5.24e4            0.00
ikev2-manager-process           any wait                 0               0            3132          1.32e4            0.00
ioam-export-process             any wait                 0               0               1          1.18e3            0.00
ip-neighbor-scan-process        any wait                 0               0              53          1.49e4            0.00
ip-route-resolver-process       any wait                 0               0              32          5.80e3            0.00
ip4-drop                         active                  5               5               0          3.13e3            1.00
ip4-local                        active                  5               5               0          1.00e4            1.00
ip4-lookup                       active                  5               5               0          1.08e6            1.00
ip4-reassembly-expire-walk      any wait                 0               0             315          1.27e4            0.00
ip6-icmp-neighbor-discovery-ev  any wait                 0               0            3132          1.12e4            0.00
ip6-input                        active                  9               9               0          3.41e3            1.00
ip6-not-enabled                  active                  9               9               0          1.47e3            1.00
ip6-reassembly-expire-walk      any wait                 0               0             315          8.52e3            0.00
l2fib-mac-age-scanner-process  event wait                0               0               1          1.18e3            0.00
lacp-process                   event wait                0               0               1          1.84e5            0.00
lisp-retry-service              any wait                 0               0            1571          1.49e4            0.00
lldp-process                   event wait                0               0               1          5.81e5            0.00
memif-process                   any wait                 0               0            1168          1.11e5            0.00
nat-det-expire-walk               done                   1               0               0          2.50e3            0.00
nat64-expire-walk              event wait                0               0               1          1.34e4            0.00
nsh-md2-ioam-export-process     any wait                 0               0               1          7.89e3            0.00
perfmon-periodic-process       event wait                0               0               1          1.18e8            0.00
rd-cp-process                   any wait                 0               0               1          1.52e3            0.00
send-dhcp6-client-message-proc  any wait                 0               0               1          1.56e3            0.00
send-dhcp6-pd-client-message-p  any wait                 0               0               1          1.53e3            0.00
send-rs-process                 any wait                 0               0               1          1.69e3            0.00
startup-config-process            done                   1               0               1          6.13e3            0.00
statseg-collector-process       time wait                0               0             315          3.77e5            0.00
udp-ping-process                any wait                 0               0               1          1.62e4            0.00
unix-cli-127.0.0.1:39670       event wait                0               0             103          2.26e7            0.00
unix-cli-127.0.0.1:40652         active                  1               0               3          4.64e9            0.00
unix-epoll-input                 polling           1698354               0               0          5.00e6            0.00
vhost-user-process              any wait                 0               0               1          5.29e3            0.00
vhost-user-send-interrupt-proc  any wait                 0               0               1          1.88e3            0.00
vpe-link-state-process         event wait                0               0              15          2.33e4            0.00
vpe-oam-process                 any wait                 0               0            1540          1.21e4            0.00
vxlan-gpe-ioam-export-process   any wait                 0               0               1          1.38e3            0.00
wildcard-ip4-arp-publisher-pro event wait                0               0               1          2.24e3            0.00
`,
			threadCount: 1,
			itemCount:   57,
			itemIdx:     1,
			item: vppcalls.RuntimeItem{
				Name:           "af-packet-input",
				State:          "interrupt wa",
				Calls:          9,
				Vectors:        9,
				Suspends:       0,
				Clocks:         1.55e5,
				VectorsPerCall: 1,
			},
		},
		{
			name: "three threads",
			reply: `Thread 0 vpp_main (lcore 0)
Time 21.5, average vectors/node 0.00, last 128 main loops 0.00 per node 0.00
  vector rates in 0.0000e0, out 5.0000e-2, drop 0.0000e0, punt 0.0000e0
             Name                 State         Calls          Vectors        Suspends         Clocks       Vectors/Call        
acl-plugin-fa-cleaner-process  event wait                6               5               1          3.12e4            0.00
api-rx-from-ring                any wait                 0               0              31          8.61e6            0.00
avf-process                    event wait                0               0               1          7.79e3            0.00
bfd-process                    event wait                0               0               1          6.80e3            0.00
cdp-process                     any wait                 0               0               1          1.78e8            0.00
dhcp-client-process             any wait                 0               0               1          2.59e3            0.00
dns-resolver-process            any wait                 0               0               1          3.35e3            0.00
fib-walk                        any wait                 0               0              11          1.08e4            0.00
flow-report-process             any wait                 0               0               1          1.64e3            0.00
flowprobe-timer-process         any wait                 0               0               1          1.16e4            0.00
igmp-timer-process             event wait                0               0               1          1.81e4            0.00
ikev2-manager-process           any wait                 0               0              22          5.47e3            0.00
ioam-export-process             any wait                 0               0               1          3.26e3            0.00
ip-route-resolver-process       any wait                 0               0               1          1.69e3            0.00
ip4-reassembly-expire-walk      any wait                 0               0               3          4.27e3            0.00
ip6-icmp-neighbor-discovery-ev  any wait                 0               0              22          4.48e3            0.00
ip6-reassembly-expire-walk      any wait                 0               0               3          6.88e3            0.00
l2fib-mac-age-scanner-process  event wait                0               0               1          3.94e3            0.00
lacp-process                   event wait                0               0               1          1.35e8            0.00
lisp-retry-service              any wait                 0               0              11          9.68e3            0.00
lldp-process                   event wait                0               0               1          1.49e8            0.00
memif-process                  event wait                0               0               1          2.67e4            0.00
nat-det-expire-walk               done                   1               0               0          5.42e3            0.00
nat64-expire-walk              event wait                0               0               1          5.87e4            0.00
rd-cp-process                   any wait                 0               0          614363          3.93e2            0.00
send-rs-process                 any wait                 0               0               1          3.22e3            0.00
startup-config-process            done                   1               0               1          1.33e4            0.00
udp-ping-process                any wait                 0               0               1          3.69e4            0.00
unix-cli-127.0.0.1:38448         active                  0               0              23          6.72e7            0.00
unix-epoll-input                 polling           8550283               0               0          3.77e3            0.00
vhost-user-process              any wait                 0               0               1          2.48e3            0.00
vhost-user-send-interrupt-proc  any wait                 0               0               1          1.43e3            0.00
vpe-link-state-process         event wait                0               0               1          1.58e3            0.00
vpe-oam-process                 any wait                 0               0              11          9.20e3            0.00
vxlan-gpe-ioam-export-process   any wait                 0               0               1          1.59e4            0.00
wildcard-ip4-arp-publisher-pro event wait                0               0               1          1.03e4            0.00
---------------
Thread 1 vpp_wk_0 (lcore 1)
Time 21.5, average vectors/node 0.00, last 128 main loops 0.00 per node 0.00
  vector rates in 0.0000e0, out 0.0000e0, drop 0.0000e0, punt 0.0000e0
             Name                 State         Calls          Vectors        Suspends         Clocks       Vectors/Call     
unix-epoll-input                 polling          15251181               0               0          3.67e3            0.00
---------------
Thread 2 vpp_wk_1 (lcore 2)
Time 21.5, average vectors/node 0.00, last 128 main loops 0.00 per node 0.00
  vector rates in 0.0000e0, out 0.0000e0, drop 0.0000e0, punt 0.0000e0
             Name                 State         Calls          Vectors        Suspends         Clocks       Vectors/Call     
unix-epoll-input                 polling          20563870               0               0          3.56e3            0.00
`,
			threadCount: 3,
			itemCount:   36,
			item: vppcalls.RuntimeItem{
				Name:           "acl-plugin-fa-cleaner-process",
				State:          "event wait",
				Calls:          6,
				Vectors:        5,
				Suspends:       1,
				Clocks:         3.12e4,
				VectorsPerCall: 0,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, handler := testSetup(t)
			defer ctx.TeardownTestCtx()

			ctx.MockVpp.MockReply(&vlib.CliInbandReply{Reply: test.reply})

			info, err := handler.GetRuntimeInfo(context.TODO())

			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(info.Threads)).To(Equal(test.threadCount))
			Expect(info.Threads[0].Items).To(HaveLen(test.itemCount))
			Expect(info.Threads[0].Items[test.itemIdx]).To(Equal(test.item))
		})
	}
}

func TestGetMemory(t *testing.T) {
	tests := []struct {
		name        string
		reply       string
		threadCount int
		threadIdx   int
		thread      vppcalls.MemoryThread
	}{
		{
			name: "single",
			reply: `Thread 0 vpp_main
  base 0x7f74752f2000, size 1g, locked, unmap-on-destroy, name 'main heap'
    page stats: page-size 4K, total 262144, mapped 14970, not-mapped 247174
      numa 0: 14970 pages, 58.48m bytes
    total: 1023.99M, used: 55.46M, free: 968.54M, trimmable: 968.53M
      free chunks 310 free fastbin blks 0
      max total allocated 1023.99M
`,
			threadCount: 1,
			threadIdx:   0,
			thread: vppcalls.MemoryThread{
				ID:              0,
				Name:            "vpp_main",
				Size:            1e9,
				Pages:           262144,
				PageSize:        4000,
				Used:            55.46e6,
				Total:           1023.99e6,
				Free:            968.54e6,
				Trimmable:       968.53e6,
				FreeChunks:      310,
				FreeFastbinBlks: 0,
				MaxTotalAlloc:   1023.99e6,
			},
		},
		{
			name: "unknown",
			reply: `Thread 0 vpp_main
  base 0x7ff4bf55f000, size 1g, locked, unmap-on-destroy, name 'main heap'
    page stats: page-size 4K, total 262144, mapped 14945, not-mapped 247174, unknown 25
      numa 0: 14945 pages, 58.38m bytes
    total: 1023.99M, used: 55.46M, free: 968.54M, trimmable: 968.53M
      free chunks 303 free fastbin blks 0
      max total allocated 1023.99M
`,
			threadCount: 1,
			threadIdx:   0,
			thread: vppcalls.MemoryThread{
				ID:              0,
				Name:            "vpp_main",
				Size:            1e9,
				Pages:           262144,
				PageSize:        4000,
				Used:            55.46e6,
				Total:           1023.99e6,
				Free:            968.54e6,
				Trimmable:       968.53e6,
				FreeChunks:      303,
				FreeFastbinBlks: 0,
				MaxTotalAlloc:   1023.99e6,
			},
		},
		{
			name: "3 workers",
			reply: `Thread 0 vpp_main
  base 0x7f0f14823000, size 1g, locked, unmap-on-destroy, name 'main heap'
    page stats: page-size 4K, total 262144, mapped 19483, not-mapped 242661
      numa 0: 19483 pages, 76.11m bytes
    total: 1023.99M, used: 72.26M, free: 951.74M, trimmable: 950.90M
      free chunks 298 free fastbin blks 0
      max total allocated 1023.99M

Thread 1 vpp_wk_0
  base 0x7f0f14823000, size 1g, locked, unmap-on-destroy, name 'main heap'
    page stats: page-size 4K, total 262144, mapped 19483, not-mapped 242661
      numa 0: 19483 pages, 76.11m bytes
    total: 1023.99M, used: 72.26M, free: 951.74M, trimmable: 950.90M
      free chunks 299 free fastbin blks 0
      max total allocated 1023.99M

Thread 2 vpp_wk_1
  base 0x7f0f14823000, size 1g, locked, unmap-on-destroy, name 'main heap'
    page stats: page-size 4K, total 262144, mapped 19483, not-mapped 242661
      numa 0: 19483 pages, 76.11m bytes
    total: 1023.99M, used: 72.26M, free: 951.74M, trimmable: 950.90M
      free chunks 299 free fastbin blks 0
      max total allocated 1023.99M

Thread 3 vpp_wk_2
  base 0x7f0f14823000, size 1g, locked, unmap-on-destroy, name 'main heap'
    page stats: page-size 4K, total 262144, mapped 19483, not-mapped 242661
      numa 0: 19483 pages, 76.11m bytes
    total: 1023.99M, used: 72.26M, free: 951.74M, trimmable: 950.90M
      free chunks 299 free fastbin blks 0
      max total allocated 1023.99M
`,
			threadCount: 4,
			threadIdx:   1,
			thread: vppcalls.MemoryThread{
				ID:              1,
				Name:            "vpp_wk_0",
				Size:            1.e9,
				Pages:           262144,
				PageSize:        4000,
				Used:            72.26e6,
				Total:           1023.99e6,
				Free:            951.74e6,
				Trimmable:       950.90e6,
				FreeChunks:      299,
				FreeFastbinBlks: 0,
				MaxTotalAlloc:   1023.99e6,
			},
		},
		// "19.08 update" test case tests for "page information not available" error.
		// It contains reply from VPP version 20.09. The format of replies changed
		// since VPP version 21.01, so the test case should be updated accordingly.
		//
		// {
		// 	name: "19.08 update",
		// //			reply: `Thread 0 vpp_main
		// //  virtual memory start 0x7fc363c20000, size 1048640k, 262160 pages, page size 4k
		// //    page information not available (errno 1)
		// //  total: 1.00G, used: 56.78M, free: 967.29M, trimmable: 966.64M
		// //    free chunks 337 free fastbin blks 0
		// //    max total allocated 1.00G
		// //`,
		// 	threadCount: 1,
		// 	threadIdx:   0,
		// 	thread: vppcalls.MemoryThread{
		// 		ID:              0,
		// 		Name:            "vpp_main",
		// 		Size:            1048.64e6,
		// 		Pages:           262160,
		// 		PageSize:        4000,
		// 		Used:            56.78e6,
		// 		Total:           1e9,
		// 		Free:            967.29e6,
		// 		Trimmable:       966.64e6,
		// 		FreeChunks:      337,
		// 		FreeFastbinBlks: 0,
		// 		MaxTotalAlloc:   1e9,
		// 	},
		// },
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, handler := testSetup(t)
			defer ctx.TeardownTestCtx()

			ctx.MockVpp.MockReply(&vlib.CliInbandReply{Reply: test.reply})

			info, err := handler.GetMemory(context.TODO())

			Expect(err).ShouldNot(HaveOccurred())
			Expect(info.Threads).To(HaveLen(test.threadCount))
			Expect(info.Threads[test.threadIdx]).To(Equal(test.thread))
		})
	}
}

func TestGetNodeCounters(t *testing.T) {
	ctx, handler := testSetup(t)
	defer ctx.TeardownTestCtx()

	const reply = `   Count                    Node                  Reason
        32            ipsec-output-ip4            IPSec policy protect
        32               esp-encrypt              ESP pkts received
        64             ipsec-input-ip4            IPSEC pkts received
        32             ip4-icmp-input             unknown type
        32             ip4-icmp-input             echo replies sent
        14             ethernet-input             l3 mac mismatch
         1                arp-input               ARP replies sent
         4                ip4-input               ip4 spoofed local-address packet drops
         2             memif1/1-output            interface is down
         1                cdp-input               good cdp packets (processed)
`
	ctx.MockVpp.MockReply(&vlib.CliInbandReply{
		Reply: reply,
	})

	info, err := handler.GetNodeCounters(context.TODO())

	Expect(err).ShouldNot(HaveOccurred())
	Expect(info.Counters).To(HaveLen(10))
	Expect(info.Counters[0]).To(Equal(vppcalls.NodeCounter{
		Value: 32,
		Node:  "ipsec-output-ip4",
		Name:  "IPSec policy protect",
	}))
	Expect(info.Counters[6]).To(Equal(vppcalls.NodeCounter{
		Value: 1,
		Node:  "arp-input",
		Name:  "ARP replies sent",
	}))
	Expect(info.Counters[7]).To(Equal(vppcalls.NodeCounter{
		Value: 4,
		Node:  "ip4-input",
		Name:  "ip4 spoofed local-address packet drops",
	}))
	Expect(info.Counters[8]).To(Equal(vppcalls.NodeCounter{
		Value: 2,
		Node:  "memif1/1-output",
		Name:  "interface is down",
	}))
	Expect(info.Counters[9]).To(Equal(vppcalls.NodeCounter{
		Value: 1,
		Node:  "cdp-input",
		Name:  "good cdp packets (processed)",
	}))
}

func testSetup(t *testing.T) (*vppmock.TestCtx, vppcalls.TelemetryVppAPI) {
	ctx := vppmock.SetupTestCtx(t)
	handler := vpp2306.NewTelemetryVppHandler(ctx.MockVPPClient)
	return ctx, handler
}
