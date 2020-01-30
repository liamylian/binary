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
	typElem := typ
	valElem := val
	if val.Kind() == reflect.Ptr {
		typElem = typ.Elem()
		valElem = val.Elem()
	}

	if typElem.Kind() != reflect.Struct {
		return nil, errors.New("not a struct or pointer to struct")
	}

	buffer := bytes.NewBuffer(nil)
	for i := 0; i < valElem.NumField(); i++ {
		fieldType := typ.Field(i)
		fieldTag := fieldType.Tag
		fieldValue := valElem.Field(i)
		endian, padding, _, err := getTagItems(fieldTag.Get("binary"))
		if err != nil {
			return nil, err
		}
		if endian == nil {
			endian = &tagItemEndian{binary.BigEndian}
		}

		switch fieldValue.Kind() {
		case reflect.Bool,
			reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Float32, reflect.Float64,
			reflect.Array, reflect.Slice:
			if padding != nil {
				return nil, errors.New("only empty struct supports padding")
			}
			if err := binary.Write(buffer, endian.Endian, fieldValue.Interface()); err != nil {
				return nil, err
			}
		case reflect.Struct:
			if fieldValue.NumField() == 0 {
				if padding != nil {
					bytes := make([]byte, padding.Bytes)
					if err := binary.Write(buffer, endian.Endian, bytes); err != nil {
						return nil, err
					}
				}
				continue
			}
			if padding != nil {
				return nil, errors.New("only empty struct supports padding")
			}
			fallthrough
		default:
			bytes, err := Pack(fieldValue.Interface())
			if err != nil {
				return nil, err
			}
			if _, err := buffer.Write(bytes); err != nil {
				return nil, err
			}
		}
	}

	return buffer.Bytes(), nil
}

func Unpack(v interface{}, bytes []byte) error {
	panic("not implemented")
}
