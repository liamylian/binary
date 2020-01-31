package binary

import (
	"encoding/binary"
	"errors"
)

func byte2Int(endian binary.ByteOrder, bytes []byte) (int, error) {
	val, err := byte2Uint(endian, bytes)
	return int(val), err
}

func byte2Uint(endian binary.ByteOrder, bytes []byte) (uint, error) {
	switch len(bytes) {
	case 1:
		num := uint8(bytes[0])
		return uint(num), nil
	case 2:
		num := endian.Uint16(bytes)
		return uint(num), nil
	case 4:
		num := endian.Uint32(bytes)
		return uint(num), nil
	case 8:
		num := endian.Uint64(bytes)
		return uint(num), nil
	default:
		return 0, errors.New("byte2Uint: bad byte size")
	}
}
