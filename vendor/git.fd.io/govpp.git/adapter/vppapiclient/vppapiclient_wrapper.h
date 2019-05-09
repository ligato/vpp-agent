// Copyright (c) 2018 Cisco and/or its affiliates.
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

#ifndef included_vppapiclient_wrapper_h
#define included_vppapiclient_wrapper_h

#include <stdlib.h>
#include <stdio.h>
#include <stdint.h>
#include <arpa/inet.h>

#include <vpp-api/client/vppapiclient.h> // VPP has to be installed!

// function go_msg_callback is defined in vppapiclient.go
extern void go_msg_callback(uint16_t msg_id, void* data, size_t size);

typedef struct __attribute__((__packed__)) _req_header {
    uint16_t msg_id;
    uint32_t client_index;
    uint32_t context;
} req_header_t;

typedef struct __attribute__((__packed__)) _reply_header {
    uint16_t msg_id;
} reply_header_t;

static void
govpp_msg_callback(unsigned char *data, int size)
{
    reply_header_t *header = ((reply_header_t *)data);
    go_msg_callback(ntohs(header->msg_id), data, size);
}

static int
govpp_send(uint32_t context, void *data, size_t size)
{
	req_header_t *header = ((req_header_t *)data);
	header->context = htonl(context);
    return vac_write(data, size);
}

static int
govpp_connect(char *shm, int rx_qlen)
{
    return vac_connect("govpp", shm, govpp_msg_callback, rx_qlen);
}

static int
govpp_disconnect()
{
    return vac_disconnect();
}

static uint32_t
govpp_get_msg_index(char *name_and_crc)
{
    return vac_get_msg_index(name_and_crc);
}

#endif /* included_vppapiclient_wrapper_h */
