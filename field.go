package binary

import (
	"encoding/binary"
	"errors"
	"reflect"
)

type fieldInfo struct {
	name       string
	tagEndian  binary.ByteOrder
	tagSize    uint
	tagSizeof  []string
	typeSize   int
	actualSize uint
	byteSize   uint
	byteStart  uint
	byteEnd    uint
}

func getFieldInfo(v interface{}) (map[string]*fieldInfo, error) {
	typ := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("not a struct")
	}

	numField := typ.NumField()
	fieldInfoMap := make(map[string]*fieldInfo)
	byteCursor := uint(0)
	for i := 0; i < numField; i++ {
		field := typ.Field(i)
		fieldType := typ.Field(i).Type
		fieldVal := val.Field(i)
		fieldName := field.Name
		tagEndian, tagSize, tagSizeof, err := getTagInfo(field.Tag.Get("binary"))
		if err != nil {
			return nil, err
		}

		typeSize := 0
		actualSize := uint(0)
		switch fieldType.Kind() {
		case reflect.Bool,
			reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Float32, reflect.Float64, reflect.Array:
			typeSize = sizeof(fieldType)
			actualSize = uint(typeSize)
		case reflect.Slice:
			typeSize = -1
			actualSize = uint(sizeof(fieldType.Elem()) * fieldVal.Len())
		case reflect.Struct:
			if fieldType.NumField() > 0 {
				return nil, errors.New("embedded none empty struct not supported")
			}
		default:
			return nil, errors.New("not supported kind")
		}
		byteSize := tagSize
		if tagSize == 0 {
			if typeSize > 0 {
				byteSize = uint(typeSize)
			} else {
				byteSize = actualSize
			}
		}
		byteStart := byteCursor
		byteEnd := byteStart + byteSize
		fieldInfoMap[fieldName] = &fieldInfo{
			name:       fieldName,
			tagEndian:  tagEndian,
			tagSize:    tagSize,
			tagSizeof:  tagSizeof,
			typeSize:   typeSize,
			actualSize: uint(actualSize),
			byteSize:   byteSize,
			byteStart:  byteStart,
			byteEnd:    byteEnd,
		}
		byteCursor += byteSize
	}

	return fieldInfoMap, nil
}

func sizeof(t reflect.Type) int {
	switch t.Kind() {
	case reflect.Array:
		if s := sizeof(t.Elem()); s >= 0 {
			return s * t.Len()
		}

	case reflect.Struct:
		sum := 0
		for i, n := 0, t.NumField(); i < n; i++ {
			s := sizeof(t.Field(i).Type)
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum

	case reflect.Bool,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return int(t.Size())
	}

	return -1
}
