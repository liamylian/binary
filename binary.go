package binary

import (
	"bytes"
	"encoding/binary"
	"errors"
	"reflect"
)

func Pack(v interface{}) ([]byte, error) {
	type fieldByteIndex struct {
		start  uint64
		end    uint64
		endian binary.ByteOrder
	}

	typ := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("not a struct")
	}

	sizeofTagItems := make(map[string]*tagItemSizeOf)
	fieldBytes := make(map[string]fieldByteIndex)
	buffer := bytes.NewBuffer(nil)
	for i := 0; i < val.NumField(); i++ {
		fieldType := typ.Field(i)
		fieldTag := fieldType.Tag
		filedName := fieldType.Name
		fieldValue := val.Field(i)
		tagEndian, tagSize, tagSizeof, err := getTagItems(fieldTag.Get("binary"))
		if err != nil {
			return nil, err
		}
		if tagEndian == nil {
			tagEndian = &tagItemEndian{binary.BigEndian}
		}
		if tagSize == nil {
			tagSize = &tagItemSize{0}
		}
		if tagSizeof != nil {
			sizeofTagItems[filedName] = tagSizeof
		}

		var bytes []byte
		if fieldValue.Kind() == reflect.Struct {
			if fieldValue.NumField() > 0 {
				bytes, err = Pack(fieldValue.Interface())
			} else if tagSize.Bytes > 0 {
				bytes = make([]byte, tagSize.Bytes)
			}
		} else {
			bytes, err = pack(fieldValue, tagEndian.Endian, tagSize.Bytes)
		}
		if err != nil {
			return nil, err
		}
		if _, err := buffer.Write(bytes); err != nil {
			return nil, err
		}

		byteEnd := buffer.Len()
		byteStart := byteEnd - len(bytes)
		fieldBytes[filedName] = fieldByteIndex{uint64(byteStart), uint64(byteEnd), tagEndian.Endian}
	}
	resultBytes := buffer.Bytes()

	for fieldName, sizeofTagItem := range sizeofTagItems {
		sumSize := uint64(0)
		sumSizeLen := uint64(8)
		for _, fieldName := range sizeofTagItem.Fields {
			byteIndex := fieldBytes[fieldName]
			sumSize += byteIndex.end - byteIndex.start
		}
		byteIndex := fieldBytes[fieldName]
		byteLen := byteIndex.end - byteIndex.start
		sumSizeBytes := make([]byte, sumSizeLen)
		byteIndex.endian.PutUint64(sumSizeBytes, sumSize)
		copy(resultBytes[byteIndex.start:byteIndex.end], sumSizeBytes[sumSizeLen-byteLen:])
	}

	return resultBytes, nil
}

func pack(v reflect.Value, endian binary.ByteOrder, size int) ([]byte, error) {
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
		byteSize := len(bytes)
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
