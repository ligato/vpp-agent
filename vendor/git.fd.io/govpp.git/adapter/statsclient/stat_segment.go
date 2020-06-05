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
	"syscall"
	"unsafe"

	"github.com/ftrvxmtrx/fd"

	"git.fd.io/govpp.git/adapter"
)

var (
	ErrStatDataLenIncorrect = fmt.Errorf("stat data length incorrect")
)

const (
	minVersion = 0
	maxVersion = 1
)

func checkVersion(ver uint64) error {
	if ver < minVersion {
		return fmt.Errorf("stat segment version is too old: %v (minimal version: %v)", ver, minVersion)
	} else if ver > maxVersion {
		return fmt.Errorf("stat segment version is not supported: %v (minimal version: %v)", ver, maxVersion)
	}
	return nil
}

type statSegment struct {
	sharedHeader []byte
	memorySize   int64

	// legacyVersion represents stat segment version 0
	// and is used as fallback for VPP 19.04
	legacyVersion bool
}

func (c *statSegment) getHeader() (header sharedHeader) {
	if c.legacyVersion {
		return loadSharedHeaderLegacy(c.sharedHeader)
	}
	return loadSharedHeader(c.sharedHeader)
}

func (c *statSegment) getEpoch() (int64, bool) {
	h := c.getHeader()
	return h.epoch, h.inProgress != 0
}

func (c *statSegment) getOffsets() (dir, err, stat int64) {
	h := c.getHeader()
	return h.directoryOffset, h.errorOffset, h.statsOffset
}

func (c *statSegment) connect(sockName string) error {
	if c.sharedHeader != nil {
		return fmt.Errorf("already connected")
	}

	addr := net.UnixAddr{
		Net:  "unixpacket",
		Name: sockName,
	}
	Log.Debugf("connecting to: %v", addr)

	conn, err := net.DialUnix(addr.Net, nil, &addr)
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

	file := files[0]
	defer func() {
		if err := file.Close(); err != nil {
			Log.Warnf("closing file failed: %v", err)
		}
	}()

	info, err := file.Stat()
	if err != nil {
		return err
	}
	size := info.Size()

	data, err := syscall.Mmap(int(file.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		Log.Debugf("mapping shared memory failed: %v", err)
		return fmt.Errorf("mapping shared memory failed: %v", err)
	}

	Log.Debugf("successfuly mmapped shared memory segment (size: %v) %v", size, len(data))

	c.sharedHeader = data
	c.memorySize = size

	hdr := loadSharedHeader(c.sharedHeader)
	Log.Debugf("stat segment header: %+v", hdr)

	if hdr.legacyVersion() {
		c.legacyVersion = true
		hdr = loadSharedHeaderLegacy(c.sharedHeader)
		Log.Debugf("falling back to legacy version (VPP <=19.04) of stat segment (header: %+v)", hdr)
	}

	if err := checkVersion(hdr.version); err != nil {
		return err
	}

	return nil
}

func (c *statSegment) disconnect() error {
	if c.sharedHeader == nil {
		return nil
	}

	if err := syscall.Munmap(c.sharedHeader); err != nil {
		Log.Debugf("unmapping shared memory failed: %v", err)
		return fmt.Errorf("unmapping shared memory failed: %v", err)
	}
	c.sharedHeader = nil

	Log.Debugf("successfuly unmapped shared memory")
	return nil
}

type statDirectoryType int32

const (
	statDirIllegal               = 0
	statDirScalarIndex           = 1
	statDirCounterVectorSimple   = 2
	statDirCounterVectorCombined = 3
	statDirErrorIndex            = 4
	statDirNameVector            = 5
	statDirEmpty                 = 6
)

func (t statDirectoryType) String() string {
	return adapter.StatType(t).String()
}

type statSegDirectoryEntry struct {
	directoryType statDirectoryType
	// unionData can represent:
	// - offset
	// - index
	// - value
	unionData    uint64
	offsetVector uint64
	name         [128]byte
}

func (c *statSegment) getStatDirVector() unsafe.Pointer {
	dirOffset, _, _ := c.getOffsets()
	return unsafe.Pointer(&c.sharedHeader[dirOffset])
}

func (c *statSegment) getStatDirIndex(p unsafe.Pointer, index uint32) *statSegDirectoryEntry {
	return (*statSegDirectoryEntry)(unsafe.Pointer(uintptr(p) + uintptr(index)*unsafe.Sizeof(statSegDirectoryEntry{})))
}

func (c *statSegment) copyEntryData(dirEntry *statSegDirectoryEntry) adapter.Stat {
	dirType := adapter.StatType(dirEntry.directoryType)

	switch dirType {
	case statDirScalarIndex:
		return adapter.ScalarStat(dirEntry.unionData)

	case statDirErrorIndex:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		_, errOffset, _ := c.getOffsets()
		offsetVector := unsafe.Pointer(&c.sharedHeader[errOffset])

		var errData adapter.Counter
		if c.legacyVersion {
			// error were not vector (per-worker) in VPP 19.04
			offset := uintptr(dirEntry.unionData) * unsafe.Sizeof(uint64(0))
			val := *(*adapter.Counter)(statSegPointer(offsetVector, offset))
			errData = val
		} else {
			vecLen := uint32(vectorLen(offsetVector))

			for i := uint32(0); i < vecLen; i++ {
				cb := *(*uint64)(statSegPointer(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
				offset := uintptr(cb) + uintptr(dirEntry.unionData)*unsafe.Sizeof(adapter.Counter(0))
				debugf("error index, cb: %d, offset: %d", cb, offset)
				val := *(*adapter.Counter)(statSegPointer(unsafe.Pointer(&c.sharedHeader[0]), offset))
				errData += val
			}
		}
		return adapter.ErrorStat(errData)

	case statDirCounterVectorSimple:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		vecLen := uint32(vectorLen(unsafe.Pointer(&c.sharedHeader[dirEntry.unionData])))
		offsetVector := statSegPointer(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := make([][]adapter.Counter, vecLen)
		for i := uint32(0); i < vecLen; i++ {
			cb := *(*uint64)(statSegPointer(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			counterVec := unsafe.Pointer(&c.sharedHeader[uintptr(cb)])
			vecLen2 := uint32(vectorLen(counterVec))
			data[i] = make([]adapter.Counter, vecLen2)
			for j := uint32(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.Counter(0))
				val := *(*adapter.Counter)(statSegPointer(counterVec, offset))
				data[i][j] = val
			}
		}
		return adapter.SimpleCounterStat(data)

	case statDirCounterVectorCombined:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		vecLen := uint32(vectorLen(unsafe.Pointer(&c.sharedHeader[dirEntry.unionData])))
		offsetVector := statSegPointer(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := make([][]adapter.CombinedCounter, vecLen)
		for i := uint32(0); i < vecLen; i++ {
			cb := *(*uint64)(statSegPointer(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			counterVec := unsafe.Pointer(&c.sharedHeader[uintptr(cb)])
			vecLen2 := uint32(vectorLen(counterVec))
			data[i] = make([]adapter.CombinedCounter, vecLen2)
			for j := uint32(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.CombinedCounter{})
				val := *(*adapter.CombinedCounter)(statSegPointer(counterVec, offset))
				data[i][j] = val
			}
		}
		return adapter.CombinedCounterStat(data)

	case statDirNameVector:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		vecLen := uint32(vectorLen(unsafe.Pointer(&c.sharedHeader[dirEntry.unionData])))
		offsetVector := statSegPointer(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := make([]adapter.Name, vecLen)
		for i := uint32(0); i < vecLen; i++ {
			cb := *(*uint64)(statSegPointer(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			if cb == 0 {
				debugf("name vector out of range for %s (%v)", dirEntry.name, i)
				continue
			}
			nameVec := unsafe.Pointer(&c.sharedHeader[cb])
			vecLen2 := uint32(vectorLen(nameVec))

			nameStr := make([]byte, 0, vecLen2)
			for j := uint32(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(byte(0))
				val := *(*byte)(statSegPointer(nameVec, offset))
				if val > 0 {
					nameStr = append(nameStr, val)
				}
			}
			data[i] = adapter.Name(nameStr)
		}
		return adapter.NameStat(data)

	case statDirEmpty:
		// no-op

	default:
		// TODO: monitor occurrences with metrics
		debugf("Unknown type %d for stat entry: %q", dirEntry.directoryType, dirEntry.name)
	}
	return nil
}

func (c *statSegment) updateEntryData(dirEntry *statSegDirectoryEntry, stat *adapter.Stat) error {
	switch (*stat).(type) {
	case adapter.ScalarStat:
		*stat = adapter.ScalarStat(dirEntry.unionData)

	case adapter.ErrorStat:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		_, errOffset, _ := c.getOffsets()
		offsetVector := unsafe.Pointer(&c.sharedHeader[errOffset])

		var errData adapter.Counter
		if c.legacyVersion {
			// error were not vector (per-worker) in VPP 19.04
			offset := uintptr(dirEntry.unionData) * unsafe.Sizeof(uint64(0))
			val := *(*adapter.Counter)(statSegPointer(offsetVector, offset))
			errData = val
		} else {
			vecLen := uint32(vectorLen(unsafe.Pointer(&c.sharedHeader[errOffset])))

			for i := uint32(0); i < vecLen; i++ {
				cb := *(*uint64)(statSegPointer(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
				offset := uintptr(cb) + uintptr(dirEntry.unionData)*unsafe.Sizeof(adapter.Counter(0))
				val := *(*adapter.Counter)(statSegPointer(unsafe.Pointer(&c.sharedHeader[0]), offset))
				errData += val
			}
		}
		*stat = adapter.ErrorStat(errData)

	case adapter.SimpleCounterStat:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		vecLen := uint32(vectorLen(unsafe.Pointer(&c.sharedHeader[dirEntry.unionData])))
		offsetVector := statSegPointer(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := (*stat).(adapter.SimpleCounterStat)
		if uint32(len(data)) != vecLen {
			return ErrStatDataLenIncorrect
		}
		for i := uint32(0); i < vecLen; i++ {
			cb := *(*uint64)(statSegPointer(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			counterVec := unsafe.Pointer(&c.sharedHeader[uintptr(cb)])
			vecLen2 := uint32(vectorLen(counterVec))
			simpData := data[i]
			if uint32(len(simpData)) != vecLen2 {
				return ErrStatDataLenIncorrect
			}
			for j := uint32(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.Counter(0))
				val := *(*adapter.Counter)(statSegPointer(counterVec, offset))
				simpData[j] = val
			}
		}

	case adapter.CombinedCounterStat:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		vecLen := uint32(vectorLen(unsafe.Pointer(&c.sharedHeader[dirEntry.unionData])))
		offsetVector := statSegPointer(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := (*stat).(adapter.CombinedCounterStat)
		if uint32(len(data)) != vecLen {
			return ErrStatDataLenIncorrect
		}
		for i := uint32(0); i < vecLen; i++ {
			cb := *(*uint64)(statSegPointer(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			counterVec := unsafe.Pointer(&c.sharedHeader[uintptr(cb)])
			vecLen2 := uint32(vectorLen(counterVec))
			combData := data[i]
			if uint32(len(combData)) != vecLen2 {
				return ErrStatDataLenIncorrect
			}
			for j := uint32(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.CombinedCounter{})
				val := *(*adapter.CombinedCounter)(statSegPointer(counterVec, offset))
				combData[j] = val
			}
		}

	case adapter.NameStat:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		vecLen := uint32(vectorLen(unsafe.Pointer(&c.sharedHeader[dirEntry.unionData])))
		offsetVector := statSegPointer(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := (*stat).(adapter.NameStat)
		if uint32(len(data)) != vecLen {
			return ErrStatDataLenIncorrect
		}
		for i := uint32(0); i < vecLen; i++ {
			cb := *(*uint64)(statSegPointer(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			if cb == 0 {
				continue
			}
			nameVec := unsafe.Pointer(&c.sharedHeader[cb])
			vecLen2 := uint32(vectorLen(nameVec))

			nameData := data[i]
			if uint32(len(nameData))+1 != vecLen2 {
				return ErrStatDataLenIncorrect
			}
			for j := uint32(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(byte(0))
				val := *(*byte)(statSegPointer(nameVec, offset))
				if val == 0 {
					break
				}
				nameData[j] = val
			}
		}

	default:
		if Debug {
			Log.Debugf("Unrecognized stat type %T for stat entry: %v", stat, dirEntry.name)
		}
	}
	return nil
}
