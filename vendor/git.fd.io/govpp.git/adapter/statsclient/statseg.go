package statsclient

import (
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	MaxWaitInProgress    = time.Millisecond * 100
	CheckDelayInProgress = time.Microsecond * 10
)

type sharedHeaderBase struct {
	epoch           int64
	inProgress      int64
	directoryOffset int64
	errorOffset     int64
	statsOffset     int64
}

type sharedHeaderV0 struct {
	sharedHeaderBase
}

type sharedHeader struct {
	version uint64
	sharedHeaderBase
}

func (h *sharedHeader) legacyVersion() bool {
	// older VPP (<=19.04) did not have version in stat segment header
	// we try to provide fallback support by skipping it in header
	if h.version > maxVersion && h.inProgress > 1 && h.epoch == 0 {
		return true
	}
	return false
}

func loadSharedHeader(b []byte) (header sharedHeader) {
	h := (*sharedHeader)(unsafe.Pointer(&b[0]))
	header.version = atomic.LoadUint64(&h.version)
	header.epoch = atomic.LoadInt64(&h.epoch)
	header.inProgress = atomic.LoadInt64(&h.inProgress)
	header.directoryOffset = atomic.LoadInt64(&h.directoryOffset)
	header.errorOffset = atomic.LoadInt64(&h.errorOffset)
	header.statsOffset = atomic.LoadInt64(&h.statsOffset)
	return
}

func loadSharedHeaderLegacy(b []byte) (header sharedHeader) {
	h := (*sharedHeaderV0)(unsafe.Pointer(&b[0]))
	header.version = 0
	header.epoch = atomic.LoadInt64(&h.epoch)
	header.inProgress = atomic.LoadInt64(&h.inProgress)
	header.directoryOffset = atomic.LoadInt64(&h.directoryOffset)
	header.errorOffset = atomic.LoadInt64(&h.errorOffset)
	header.statsOffset = atomic.LoadInt64(&h.statsOffset)
	return
}

type statSegAccess struct {
	epoch int64
}

func (c *statSegment) accessStart() statSegAccess {
	t := time.Now()

	epoch, inprog := c.getEpoch()
	for inprog {
		if time.Since(t) > MaxWaitInProgress {
			return statSegAccess{}
		} else {
			time.Sleep(CheckDelayInProgress)
		}
		epoch, inprog = c.getEpoch()
	}
	return statSegAccess{
		epoch: epoch,
	}
}

func (c *statSegment) accessEnd(acc *statSegAccess) bool {
	epoch, inprog := c.getEpoch()
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
	vec := *(*vecHeader)(unsafe.Pointer(uintptr(v) - unsafe.Sizeof(uint64(0))))
	return vec.length
}

//go:nosplit
func statSegPointer(p unsafe.Pointer, offset uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + offset)
}
