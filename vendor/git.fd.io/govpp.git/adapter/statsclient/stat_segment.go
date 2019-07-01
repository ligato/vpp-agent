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

package statsclient

import (
	"fmt"
	"net"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/ftrvxmtrx/fd"

	"git.fd.io/govpp.git/adapter"
)

var (
	maxWaitInProgress = time.Second * 1
)

type statSegDirectoryEntry struct {
	directoryType statDirectoryType
	// unionData can represent: offset, index or value
	unionData    uint64
	offsetVector uint64
	name         [128]byte
}

type statDirectoryType int32

func (t statDirectoryType) String() string {
	return adapter.StatType(t).String()
}

type statSegment struct {
	sharedHeader []byte
	memorySize   int64

	oldHeader bool
}

func (c *statSegment) connect(sockName string) error {
	addr := &net.UnixAddr{
		Net:  "unixpacket",
		Name: sockName,
	}

	Log.Debugf("connecting to: %v", addr)

	conn, err := net.DialUnix(addr.Net, nil, addr)
	if err != nil {
		Log.Warnf("connecting to socket %s failed: %s", addr, err)
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			Log.Warnf("closing socket failed: %v", err)
		}
	}()

	Log.Debugf("connected to socket")

	files, err := fd.Get(conn, 1, nil)
	if err != nil {
		return fmt.Errorf("getting file descriptor over socket failed: %v", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no files received over socket")
	}
	defer func() {
		for _, f := range files {
			if err := f.Close(); err != nil {
				Log.Warnf("closing file %s failed: %v", f.Name(), err)
			}
		}
	}()

	Log.Debugf("received %d files over socket", len(files))

	f := files[0]

	info, err := f.Stat()
	if err != nil {
		return err
	}

	size := info.Size()

	Log.Debugf("fd: name=%v size=%v", info.Name(), size)

	data, err := syscall.Mmap(int(f.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		Log.Warnf("mapping shared memory failed: %v", err)
		return fmt.Errorf("mapping shared memory failed: %v", err)
	}

	Log.Debugf("successfuly mapped shared memory")

	c.sharedHeader = data
	c.memorySize = size

	header := c.readHeader()
	Log.Debugf("stat segment header: %+v", header)

	// older VPP (19.04) did not have version in stat segment header
	// we try to provide fallback support by skipping it in header
	if header.version > MaxVersion && header.inProgress > 1 && header.epoch == 0 {
		h := c.readHeaderOld()
		Log.Infof("statsclient: falling back to old stat segment version (VPP 19.04): %+v", h)
		c.oldHeader = true
	}

	return nil
}

func (c *statSegment) disconnect() error {
	if err := syscall.Munmap(c.sharedHeader); err != nil {
		Log.Warnf("unmapping shared memory failed: %v", err)
		return fmt.Errorf("unmapping shared memory failed: %v", err)
	}

	Log.Debugf("successfuly unmapped shared memory")

	return nil
}

type sharedHeaderBase struct {
	epoch           int64
	inProgress      int64
	directoryOffset int64
	errorOffset     int64
	statsOffset     int64
}

type statSegSharedHeader struct {
	version uint64
	sharedHeaderBase
}

func (c *statSegment) readHeaderOld() (header statSegSharedHeader) {
	h := (*sharedHeaderBase)(unsafe.Pointer(&c.sharedHeader[0]))
	header.version = 0
	header.epoch = atomic.LoadInt64(&h.epoch)
	header.inProgress = atomic.LoadInt64(&h.inProgress)
	header.directoryOffset = atomic.LoadInt64(&h.directoryOffset)
	header.errorOffset = atomic.LoadInt64(&h.errorOffset)
	header.statsOffset = atomic.LoadInt64(&h.statsOffset)
	return
}

func (c *statSegment) readHeader() (header statSegSharedHeader) {
	h := (*statSegSharedHeader)(unsafe.Pointer(&c.sharedHeader[0]))
	header.version = atomic.LoadUint64(&h.version)
	header.epoch = atomic.LoadInt64(&h.epoch)
	header.inProgress = atomic.LoadInt64(&h.inProgress)
	header.directoryOffset = atomic.LoadInt64(&h.directoryOffset)
	header.errorOffset = atomic.LoadInt64(&h.errorOffset)
	header.statsOffset = atomic.LoadInt64(&h.statsOffset)
	return
}

func (c *statSegment) readVersion() uint64 {
	if c.oldHeader {
		return 0
	}
	header := (*statSegSharedHeader)(unsafe.Pointer(&c.sharedHeader[0]))
	version := atomic.LoadUint64(&header.version)
	return version
}

func (c *statSegment) readEpoch() (int64, bool) {
	if c.oldHeader {
		h := c.readHeaderOld()
		return h.epoch, h.inProgress != 0
	}
	header := (*statSegSharedHeader)(unsafe.Pointer(&c.sharedHeader[0]))
	epoch := atomic.LoadInt64(&header.epoch)
	inprog := atomic.LoadInt64(&header.inProgress)
	return epoch, inprog != 0
}

func (c *statSegment) readOffsets() (dir, err, stat int64) {
	if c.oldHeader {
		h := c.readHeaderOld()
		return h.directoryOffset, h.errorOffset, h.statsOffset
	}
	header := (*statSegSharedHeader)(unsafe.Pointer(&c.sharedHeader[0]))
	dirOffset := atomic.LoadInt64(&header.directoryOffset)
	errOffset := atomic.LoadInt64(&header.errorOffset)
	statOffset := atomic.LoadInt64(&header.statsOffset)
	return dirOffset, errOffset, statOffset
}

type statSegAccess struct {
	epoch int64
}

func (c *statSegment) accessStart() *statSegAccess {
	epoch, inprog := c.readEpoch()
	t := time.Now()
	for inprog {
		if time.Since(t) > maxWaitInProgress {
			return nil
		}
		epoch, inprog = c.readEpoch()
	}
	return &statSegAccess{
		epoch: epoch,
	}
}

func (c *statSegment) accessEnd(acc *statSegAccess) bool {
	epoch, inprog := c.readEpoch()
	if acc.epoch != epoch || inprog {
		return false
	}
	return true
}

type vecHeader struct {
	length     uint64
	vectorData [0]uint8
}

func vectorLen(v unsafe.Pointer) uint64 {
	vec := *(*vecHeader)(unsafe.Pointer(uintptr(v) - unsafe.Sizeof(uintptr(0))))
	return vec.length
}

//go:nosplit
func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}
