package binary

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBinary(t *testing.T) {
	type packet struct {
		Protocol uint16
		Version  uint8
		Length   uint16 `binary:"big,sizeof=Cmd+Data+Padding+CRC"`
		Cmd      uint8
		Padding  [2]byte
		Data     []byte
		CRC      uint16
	}

	pkt1 := packet{
		Protocol: 1,
		Version:  2,
		Length:   8,
		Cmd:      8,
		Data:     []byte{9, 10, 11},
		CRC:      12,
	}

	bytes, err := Pack(pkt1)
	assert.Nil(t, err)

	pkt2 := packet{}
	err = Unpack(&pkt2, bytes)
	assert.Nil(t, err)
	assert.Equal(t, pkt1, pkt2)
}
