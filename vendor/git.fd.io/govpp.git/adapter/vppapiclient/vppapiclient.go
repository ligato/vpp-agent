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

// +build !windows,!darwin

package vppapiclient

/*
#cgo CFLAGS: -DPNG_DEBUG=1
#cgo LDFLAGS: -lvppapiclient

#include "vppapiclient_wrapper.h"
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"
	"unsafe"

	"github.com/fsnotify/fsnotify"

	"git.fd.io/govpp.git/adapter"
)

var (
	// MaxWaitReady defines maximum duration before waiting for shared memory
	// segment times out
	MaxWaitReady = time.Second * 15
)

const (
	// shmDir is a directory where shared memory is supposed to be created.
	shmDir = "/dev/shm/"
	// vppShmFile is a default name of the file in the shmDir.
	vppShmFile = "vpe-api"
)

// global VPP binary API client, library vppapiclient only supports
// single connection at a time
var globalVppClient *vppClient

// stubVppClient is the default implementation of the VppAPI.
type vppClient struct {
	shmPrefix      string
	msgCallback    adapter.MsgCallback
	inputQueueSize uint16
}

// NewVppClient returns a new VPP binary API client.
func NewVppClient(shmPrefix string) adapter.VppAPI {
	return NewVppClientWithInputQueueSize(shmPrefix, 32)
}

// NewVppClientWithInputQueueSize returns a new VPP binary API client with a custom input queue size.
func NewVppClientWithInputQueueSize(shmPrefix string, inputQueueSize uint16) adapter.VppAPI {
	return &vppClient{
		shmPrefix:      shmPrefix,
		inputQueueSize: inputQueueSize,
	}
}

// Connect connects the process to VPP.
func (a *vppClient) Connect() error {
	if globalVppClient != nil {
		return fmt.Errorf("already connected to binary API, disconnect first")
	}

	rxQlen := C.int(a.inputQueueSize)
	var rc C.int
	if a.shmPrefix == "" {
		rc = C.govpp_connect(nil, rxQlen)
	} else {
		shm := C.CString(a.shmPrefix)
		rc = C.govpp_connect(shm, rxQlen)
	}
	if rc != 0 {
		return fmt.Errorf("connecting to VPP binary API failed (rc=%v)", rc)
	}

	globalVppClient = a
	return nil
}

// Disconnect disconnects the process from VPP.
func (a *vppClient) Disconnect() error {
	globalVppClient = nil

	rc := C.govpp_disconnect()
	if rc != 0 {
		return fmt.Errorf("disconnecting from VPP binary API failed (rc=%v)", rc)
	}

	return nil
}

// GetMsgID returns a runtime message ID for the given message name and CRC.
func (a *vppClient) GetMsgID(msgName string, msgCrc string) (uint16, error) {
	nameAndCrc := C.CString(msgName + "_" + msgCrc)
	defer C.free(unsafe.Pointer(nameAndCrc))

	msgID := uint16(C.govpp_get_msg_index(nameAndCrc))
	if msgID == ^uint16(0) {
		// VPP does not know this message
		return msgID, fmt.Errorf("unknown message: %v (crc: %v)", msgName, msgCrc)
	}

	return msgID, nil
}

// SendMsg sends a binary-encoded message to VPP.
func (a *vppClient) SendMsg(context uint32, data []byte) error {
	rc := C.govpp_send(C.uint32_t(context), unsafe.Pointer(&data[0]), C.size_t(len(data)))
	if rc != 0 {
		return fmt.Errorf("unable to send the message (rc=%v)", rc)
	}
	return nil
}

// SetMsgCallback sets a callback function that will be called by the adapter
// whenever a message comes from VPP.
func (a *vppClient) SetMsgCallback(cb adapter.MsgCallback) {
	a.msgCallback = cb
}

// WaitReady blocks until shared memory for sending
// binary api calls is present on the file system.
func (a *vppClient) WaitReady() error {
	// join the path to the shared memory segment
	var path string
	if a.shmPrefix == "" {
		path = filepath.Join(shmDir, vppShmFile)
	} else {
		path = filepath.Join(shmDir, a.shmPrefix+"-"+vppShmFile)
	}

	// check if file at the path already exists
	if _, err := os.Stat(path); err == nil {
		// file exists, we are ready
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	// file does not exist, start watching folder
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// start watching directory
	if err := watcher.Add(shmDir); err != nil {
		return err
	}

	for {
		select {
		case <-time.After(MaxWaitReady):
			return fmt.Errorf("waiting for shared memory segment timed out (%s)", MaxWaitReady)
		case e := <-watcher.Errors:
			return e
		case ev := <-watcher.Events:
			if ev.Name == path {
				if (ev.Op & fsnotify.Create) == fsnotify.Create {
					// file was created, we are ready
					return nil
				}
			}
		}
	}
}

//export go_msg_callback
func go_msg_callback(msgID C.uint16_t, data unsafe.Pointer, size C.size_t) {
	// convert unsafe.Pointer to byte slice
	sliceHeader := &reflect.SliceHeader{Data: uintptr(data), Len: int(size), Cap: int(size)}
	byteSlice := *(*[]byte)(unsafe.Pointer(sliceHeader))

	globalVppClient.msgCallback(uint16(msgID), byteSlice)
}
