package insteon //nolint:testpackage

import (
	"testing"
)

func TestCalculateCRC(t *testing.T) {
	test1 := []byte{0x2f, 0x00, 0x00, 0x02, 0x0f, 0xe7, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	expected := byte(0xd7)
	actual := calculateCRC(test1)

	if expected != actual {
		t.Fatalf("crc checksum fail: %x != %x", expected, actual)
	}
}
