package smbios_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yumaojun03/dmidecode/smbios"
)

var (
	s = &smbios.Structure{
		Header: smbios.Header{
			Type:   9,
			Length: 17,
			Handle: 2304,
		},
		Formatted: []byte{0x1, 0xb1, 0xd, 0x3, 0x4, 0x1, 0x0, 0x4, 0x1, 0xff, 0xff, 0xff, 0xff},
		Strings:   []string{"PCIe Slot 1"},
	}
)

func TestRead(t *testing.T) {
	_, ss, err := smbios.ReadStructures()
	t.Log(ss, err)

}

func TestTypes(t *testing.T) {
	should := assert.New(t)

	should.Equal("System Boot", smbios.SystemBoot.String())
}

func TestGetByte(t *testing.T) {
	should := assert.New(t)

	should.Equal(s.GetByte(0x0b), uint8(0xff))
}

func TestGetBytes(t *testing.T) {
	should := assert.New(t)

	should.Equal(s.GetBytes(0x09, 0x0b), []uint8([]byte{0xff, 0xff}))
}

func TestString(t *testing.T) {
	should := assert.New(t)

	should.Equal(s.GetString(0x0), "PCIe Slot 1")
}

// TestGetString_NegativeOffset tests that negative offsets return <BAD INDEX>
func TestGetString_NegativeOffset(t *testing.T) {
	should := assert.New(t)

	should.Equal(s.GetString(-1), "<BAD INDEX>")
	should.Equal(s.GetString(-100), "<BAD INDEX>")
}

// TestGetString_OffsetOutOfBounds tests that offsets beyond FormattedCount return Unknown
func TestGetString_OffsetOutOfBounds(t *testing.T) {
	should := assert.New(t)

	// s.Formatted has 13 bytes (indices 0-12)
	should.Equal(s.GetString(13), "Unknown")
	should.Equal(s.GetString(100), "Unknown")
}

// TestGetString_StringIndexZero tests that string index 0 returns Unknown
func TestGetString_StringIndexZero(t *testing.T) {
	should := assert.New(t)

	// Create a structure with index 0 at a specific offset
	testStruct := &smbios.Structure{
		Header: smbios.Header{
			Type:   17,
			Length: 10,
			Handle: 0x26,
		},
		Formatted: []byte{0x00, 0x01, 0x02}, // index 0, 1, 2
		Strings:   []string{"String1", "String2"},
	}

	should.Equal(testStruct.GetString(0), "Unknown") // index = 0
}

// TestGetString_StringIndexOutOfBounds tests that invalid string indices return <BAD INDEX>
// This is the actual bug that was causing panics in production
func TestGetString_StringIndexOutOfBounds(t *testing.T) {
	should := assert.New(t)

	// Create a structure that mimics the real-world failure case
	// Formatted data points to string index 3, but only 2 strings exist
	testStruct := &smbios.Structure{
		Header: smbios.Header{
			Type:   17, // Memory Device
			Length: 92,
			Handle: 0x26,
		},
		Formatted: []byte{0x01, 0x02, 0x03, 0x04}, // indices: 1, 2, 3, 4
		Strings:   []string{"CPU0_DIMM_A1", "NODE 0"}, // only 2 strings (indices 1-2 valid)
	}

	// Valid indices (within strings array)
	should.Equal(testStruct.GetString(0), "CPU0_DIMM_A1") // index 1 -> Strings[0]
	should.Equal(testStruct.GetString(1), "NODE 0")       // index 2 -> Strings[1]

	// Invalid indices (beyond strings array) - should return <BAD INDEX> instead of panicking
	should.Equal(testStruct.GetString(2), "<BAD INDEX>") // index 3, but only 2 strings
	should.Equal(testStruct.GetString(3), "<BAD INDEX>") // index 4, but only 2 strings
}

// TestGetString_ValidIndex tests normal operation with valid indices
func TestGetString_ValidIndex(t *testing.T) {
	should := assert.New(t)

	testStruct := &smbios.Structure{
		Header: smbios.Header{
			Type:   1,
			Length: 10,
			Handle: 0x01,
		},
		Formatted: []byte{0x01, 0x02, 0x03},
		Strings:   []string{"Manufacturer", "Product", "Version"},
	}

	should.Equal(testStruct.GetString(0), "Manufacturer")
	should.Equal(testStruct.GetString(1), "Product")
	should.Equal(testStruct.GetString(2), "Version")
}

func TestU16(t *testing.T) {
	should := assert.New(t)

	should.Equal(s.U16(0x05, 0x07), uint16(0x1))
}

func TestU32(t *testing.T) {
	should := assert.New(t)

	should.Equal(s.U32(0x05, 0x09), uint32(0x1040001))
}

func TestU64(t *testing.T) {
	should := assert.New(t)

	should.Equal(s.U64(0x05, 0x0d), uint64(0x0))
}

func TestParsePanice(t *testing.T) {
	should := assert.New(t)

	err := mockParse(s)
	should.Equal(err.Error(), "parse structure (Header: Type: 9, Length: 17, Handle: 2304, Data: [1 177 13 3 4 1 0 4 1 255 255 255 255] Strings: [PCIe Slot 1])  pannic: parse panic test")
}

func mockParse(*smbios.Structure) (err error) {
	defer smbios.ParseRecovery(s, &err)
	panic("parse panic test")
}
