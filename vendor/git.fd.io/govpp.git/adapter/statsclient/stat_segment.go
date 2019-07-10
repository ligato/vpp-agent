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

type statDirectoryType int32

func (t statDirectoryType) String() string {
	return adapter.StatType(t).String()
}

type statSegDirectoryEntry struct {
	directoryType statDirectoryType
	// unionData can represent: offset, index or value
	unionData    uint64
	offsetVector uint64
	name         [128]byte
}

type statSegment struct {
	sharedHeader []byte
	memorySize   int64

	// oldHeader defines version 0 for stat segment
	// and is used for VPP 19.04
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
		Log.Warnf("statsclient: falling back to old stat segment version (VPP 19.04): %+v", h)
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

func (c *statSegment) copyData(dirEntry *statSegDirectoryEntry) adapter.Stat {
	switch typ := adapter.StatType(dirEntry.directoryType); typ {
	case adapter.ScalarIndex:
		return adapter.ScalarStat(dirEntry.unionData)

	case adapter.ErrorIndex:
		_, errOffset, _ := c.readOffsets()
		offsetVector := unsafe.Pointer(&c.sharedHeader[errOffset])

		var errData adapter.Counter
		if c.oldHeader {
			// error were not vector (per-worker) in VPP 19.04
			offset := uintptr(dirEntry.unionData) * unsafe.Sizeof(uint64(0))
			val := *(*adapter.Counter)(add(offsetVector, offset))
			errData = val
		} else {
			vecLen := vectorLen(offsetVector)
			for i := uint64(0); i < vecLen; i++ {
				cb := *(*uint64)(add(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
				offset := uintptr(cb) + uintptr(dirEntry.unionData)*unsafe.Sizeof(adapter.Counter(0))
				val := *(*adapter.Counter)(add(unsafe.Pointer(&c.sharedHeader[0]), offset))
				errData += val
			}
		}
		return adapter.ErrorStat(errData)

	case adapter.SimpleCounterVector:
		if dirEntry.unionData == 0 {
			Log.Debugf("\toffset is not valid")
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			Log.Debugf("\toffset out of range")
			break
		}

		simpleCounter := unsafe.Pointer(&c.sharedHeader[dirEntry.unionData]) // offset
		vecLen := vectorLen(simpleCounter)
		offsetVector := add(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := make([][]adapter.Counter, vecLen)
		for i := uint64(0); i < vecLen; i++ {
			cb := *(*uint64)(add(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			counterVec := unsafe.Pointer(&c.sharedHeader[uintptr(cb)])
			vecLen2 := vectorLen(counterVec)
			for j := uint64(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.Counter(0))
				val := *(*adapter.Counter)(add(counterVec, offset))
				data[i] = append(data[i], val)
			}
		}
		return adapter.SimpleCounterStat(data)

	case adapter.CombinedCounterVector:
		if dirEntry.unionData == 0 {
			Log.Debugf("\toffset is not valid")
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			Log.Debugf("\toffset out of range")
			break
		}

		combinedCounter := unsafe.Pointer(&c.sharedHeader[dirEntry.unionData]) // offset
		vecLen := vectorLen(combinedCounter)
		offsetVector := add(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := make([][]adapter.CombinedCounter, vecLen)
		for i := uint64(0); i < vecLen; i++ {
			cb := *(*uint64)(add(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			counterVec := unsafe.Pointer(&c.sharedHeader[uintptr(cb)])
			vecLen2 := vectorLen(counterVec)
			for j := uint64(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.CombinedCounter{})
				val := *(*adapter.CombinedCounter)(add(counterVec, offset))
				data[i] = append(data[i], val)
			}
		}
		return adapter.CombinedCounterStat(data)

	case adapter.NameVector:
		if dirEntry.unionData == 0 {
			Log.Debugf("\toffset is not valid")
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			Log.Debugf("\toffset out of range")
			break
		}

		nameVector := unsafe.Pointer(&c.sharedHeader[dirEntry.unionData]) // offset
		vecLen := vectorLen(nameVector)
		offsetVector := add(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := make([]adapter.Name, vecLen)
		for i := uint64(0); i < vecLen; i++ {
			cb := *(*uint64)(add(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			if cb == 0 {
				Log.Debugf("\tname vector cb out of range")
				continue
			}
			nameVec := unsafe.Pointer(&c.sharedHeader[cb])
			vecLen2 := vectorLen(nameVec)

			var nameStr []byte
			for j := uint64(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(byte(0))
				val := *(*byte)(add(nameVec, offset))
				if val > 0 {
					nameStr = append(nameStr, val)
				}
			}
			data[i] = adapter.Name(nameStr)
		}
		return adapter.NameStat(data)

	default:
		Log.Warnf("Unknown type %d for stat entry: %q", dirEntry.directoryType, dirEntry.name)
	}

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
