package binary

import (
	"bytes"
	"encoding/binary"
	"errors"
	"reflect"
)

func Pack(v interface{}) ([]byte, error) {
	typ := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		return nil, errors.New("not a struct")
	}
	fieldInfoMap, err := getFieldInfo(v)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(nil)
	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)
		fieldInfo := fieldInfoMap[field.Name]
		var bytes []byte

		if fieldValue.Kind() == reflect.Struct {
			bytes = make([]byte, fieldInfo.tagSize)
		} else {
			bytes, err = pack(fieldValue, fieldInfo.tagEndian, fieldInfo.tagSize)
			if err != nil {
				return nil, err
			}
		}
		if _, err := buffer.Write(bytes); err != nil {
			return nil, err
		}
	}
	resultBytes := buffer.Bytes()

	for _, fieldInfo := range fieldInfoMap {
		sumSize := uint64(0)
		sumSizeLen := uint(8)
		for _, otherFieldName := range fieldInfo.tagSizeof {
			otherFieldInfo := fieldInfoMap[otherFieldName]
			sumSize += uint64(otherFieldInfo.tagSize)
		}
		sumSizeBytes := make([]byte, sumSizeLen)
		fieldInfo.tagEndian.PutUint64(sumSizeBytes, sumSize)
		copy(resultBytes[fieldInfo.byteStart:fieldInfo.byteEnd], sumSizeBytes[sumSizeLen-fieldInfo.tagSize:])
	}

	return resultBytes, nil
}

func pack(v reflect.Value, endian binary.ByteOrder, size uint) ([]byte, error) {
	switch v.Kind() {
	case reflect.Bool,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64,
		reflect.Array, reflect.Slice:
		buffer := bytes.NewBuffer(nil)
		if err := binary.Write(buffer, endian, v.Interface()); err != nil {
			return nil, err
		}
		bytes := buffer.Bytes()
		byteSize := uint(len(bytes))
		if size == 0 {
			return bytes, nil
		} else if size > byteSize {
			padding := make([]byte, size-byteSize, size)
			return append(padding, bytes...), nil
		} else {
			return bytes[byteSize-size:], nil
		}
	default:
		return nil, errors.New("not supported kind")
	}
}

func Unpack(v interface{}, bytes []byte) error {
	panic("not implemented")
}
