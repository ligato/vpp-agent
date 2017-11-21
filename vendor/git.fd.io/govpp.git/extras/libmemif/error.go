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

package libmemif

/*
#cgo LDFLAGS: -lmemif

#include <unistd.h>
#include <libmemif.h>
*/
import "C"

// List of errors thrown by go-libmemif.
// Error handling code should compare returned error by value against these variables.
var (
	ErrSyscall       = newMemifError(1)
	ErrAccess        = newMemifError(2)
	ErrNoFile        = newMemifError(3)
	ErrFileLimit     = newMemifError(4)
	ErrProcFileLimit = newMemifError(5)
	ErrAlready       = newMemifError(6)
	ErrAgain         = newMemifError(7)
	ErrBadFd         = newMemifError(8)
	ErrNoMem         = newMemifError(9)
	ErrInvalArgs     = newMemifError(10)
	ErrNoConn        = newMemifError(11)
	ErrConn          = newMemifError(12)
	ErrClbFDUpdate   = newMemifError(13)
	ErrFileNotSock   = newMemifError(14)
	ErrNoShmFD       = newMemifError(15)
	ErrCookie        = newMemifError(16)

	// Not thrown, instead properly handled inside the golang adapter:
	ErrNoBufRing    = newMemifError(17)
	ErrNoBuf        = newMemifError(18)
	ErrNoBufDetails = newMemifError(19)

	ErrIntWrite     = newMemifError(20)
	ErrMalformedMsg = newMemifError(21)
	ErrQueueID      = newMemifError(22)
	ErrProto        = newMemifError(23)
	ErrIfID         = newMemifError(24)
	ErrAcceptSlave  = newMemifError(25)
	ErrAlreadyConn  = newMemifError(26)
	ErrMode         = newMemifError(27)
	ErrSecret       = newMemifError(28)
	ErrNoSecret     = newMemifError(29)
	ErrMaxRegion    = newMemifError(30)
	ErrMaxRing      = newMemifError(31)
	ErrNotIntFD     = newMemifError(32)
	ErrDisconnect   = newMemifError(33)
	ErrDisconnected = newMemifError(34)
	ErrUnknownMsg   = newMemifError(35)
	ErrPollCanceled = newMemifError(36)

	// Errors added by the adapter:
	ErrNotInit     = newMemifError(100, "libmemif is not initialized")
	ErrAlreadyInit = newMemifError(101, "libmemif is already initialized")
	ErrUnsupported = newMemifError(102, "the feature is not supported by C-libmemif")

	// Received unrecognized error code from C-libmemif.
	ErrUnknown = newMemifError(-1, "unknown error")
)

// MemifError implements and extends the error interface with the method Code(),
// which returns the integer error code as returned by C-libmemif.
type MemifError struct {
	code        int
	description string
}

// Error prints error description.
func (e *MemifError) Error() string {
	return e.description
}

// Code returns the integer error code as returned by C-libmemif.
func (e *MemifError) Code() int {
	return e.code
}

// A registry of libmemif errors. Used to convert C-libmemif error code into
// the associated MemifError.
var errorRegistry = map[int]*MemifError{}

// newMemifError builds and registers a new MemifError.
func newMemifError(code int, desc ...string) *MemifError {
	var err *MemifError
	if len(desc) > 0 {
		err = &MemifError{code: code, description: "libmemif: " + desc[0]}
	} else {
		err = &MemifError{code: code, description: "libmemif: " + C.GoString(C.memif_strerror(C.int(code)))}
	}
	errorRegistry[code] = err
	return err
}

// getMemifError returns the MemifError associated with the given C-libmemif
// error code.
func getMemifError(code int) error {
	if code == 0 {
		return nil /* success */
	}
	err, known := errorRegistry[code]
	if !known {
		return ErrUnknown
	}
	return err
}
