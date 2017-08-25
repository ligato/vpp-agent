// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crc16

import "testing"

type testCase struct {
	Message []byte
	CRC     uint16
}

func TestARC(t *testing.T) {
	tests := []testCase{
		{[]byte("123456789"), 0xBB3D}}
	table := MakeTable(IBM)
	for _, testcase := range tests {
		result := ^Update(0xFFFF, table, testcase.Message)
		if testcase.CRC != result {
			t.Fatalf("ARC CRC-16 value is incorrect, expected %x, received %x.", testcase.CRC, result)
		}
	}
}

func TestModbus(t *testing.T) {
	tests := []testCase{
		{[]byte{0xEA, 0x03, 0x00, 0x00, 0x00, 0x64}, 0x3A53},
		{[]byte{0x4B, 0x03, 0x00, 0x2C, 0x00, 0x37}, 0xBFCB},
		{[]byte("123456789"), 0x4B37},
		{[]byte{0x0D, 0x01, 0x00, 0x62, 0x00, 0x33}, 0x0DDD}}
	for _, testcase := range tests {
		result := ^ChecksumIBM(testcase.Message)
		if testcase.CRC != result {
			t.Fatalf("Modbus CRC-16 value is incorrect, expected %d, received %d.", testcase.CRC, result)
		}
	}
}

func BenchmarkChecksumIBM(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ChecksumIBM([]byte{0xEA, 0x03, 0x00, 0x00, 0x00, 0x64})
	}
}

func BenchmarkMakeTable(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MakeTable(IBM)
	}
}

func TestCCITTFalse(t *testing.T) {
	data := []byte("testdata")
	target := uint16(0xDC7C)

	actual := ChecksumCCITTFalse(data)
	if actual != target {
		t.Fatalf("CCITT checksum did not return the correct value, expected %x, received %x", target, actual)
	}
}
