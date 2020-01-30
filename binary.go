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
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("not a struct")
	}

	buffer := bytes.NewBuffer(nil)
	for i := 0; i < val.NumField(); i++ {
		fieldType := typ.Field(i)
		fieldTag := fieldType.Tag
		fieldValue := val.Field(i)
		tagEndian, tagSize, _, err := getTagItems(fieldTag.Get("binary"))
		if err != nil {
			return nil, err
		}
		if tagEndian == nil {
			tagEndian = &tagItemEndian{binary.BigEndian}
		}
		if tagSize == nil {
			tagSize = &tagItemSize{0}
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
	}

	return buffer.Bytes(), nil
}

func Unpack(v interface{}, bytes []byte) error {
	panic("not implemented")
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
		if size == 0 {
			return bytes, nil
		} else if byteSize := len(bytes); size > byteSize {
			padding := make([]byte, size-byteSize, size)
			return append(padding, bytes...), nil
		} else {
			return bytes[:size], nil
		}
	default:
		return nil, errors.New("not supported kind")
	}
}
