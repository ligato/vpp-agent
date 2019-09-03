// Copyright (c) 2019 Cisco and/or its affiliates.
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

package supervisor

import (
	"bufio"
	"log"
	"os"
	"sync"

	"github.com/pkg/errors"
)

// SvLogger is a logger object compatible with the process manager. It uses
// writer to print log to stdout or a file
type SvLogger struct {
	mx sync.Mutex
	writer *bufio.Writer

	file *os.File
}

// NewSvLogger prepares new supervisor logger for given process.
func NewSvLogger(logfilePath string) (svLogger *SvLogger, err error) {
	var file *os.File

	writer := bufio.NewWriterSize(os.Stdout, 1)
	if logfilePath != "" {
		if file, err = os.OpenFile(logfilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666); err != nil {
			return nil, errors.Errorf("failed to open log file %s: %v", logfilePath, err)
		}
		writer = bufio.NewWriter(file)
	}

	return &SvLogger{
		writer: writer,
		file:   file,
	}, nil
}

// Write message to the file or stdout
func (l *SvLogger) Write(p []byte) (n int, err error) {
	l.mx.Lock()
	defer l.mx.Unlock()

	return l.writer.Write(p)
}

// Close the file if necessary
func (l *SvLogger) Close() error {
	if err := l.writer.Flush(); err != nil {
		log.Printf("error writing buffer to file: %v", err)
	}
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
