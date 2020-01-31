package binary

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBinary(t *testing.T) {
	type packet struct {
		Protocol uint16
		Version  uint8
		Length   uint16 `binary:"big,sizeof=Cmd+Data+Padding+CRC"`
		Cmd      uint8
		Padding  struct{} `binary:"size=1B"`
		Data     []byte
		CRC      uint16
	}

	pkt1 := packet{
		Protocol: 1,
		Version:  2,
		Length:   0, // 7
		Cmd:      8,
		Data:     []byte{9, 10, 11},
		CRC:      12,
	}

	bytes, err := Pack(pkt1)
	assert.Nil(t, err)
	fmt.Printf("%X\n", bytes)

	pkt2 := packet{}
	err = Unpack(&pkt2, bytes)
	assert.Nil(t, err)
	// assert.Equal(t, pkt1, pkt2)
}
