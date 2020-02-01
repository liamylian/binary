package binary

import (
	"encoding/binary"
	"errors"
	"reflect"
)

type fieldInfo struct {
	name                string
	tagEndian           binary.ByteOrder
	tagSizeof           []string
	tagSizeofVal        uint
	tagPadding          uint
	byteSizeNeedResolve bool
	byteSize            uint
	byteStart           uint
	byteEnd             uint
}

func getFieldInfo(v interface{}) (map[string]*fieldInfo, error) {
	typElem := reflect.TypeOf(v)
	valElem := reflect.ValueOf(v)
	if typElem.Kind() == reflect.Ptr {
		typElem = typElem.Elem()
		valElem = valElem.Elem()
	}
	if typElem.Kind() != reflect.Struct {
		return nil, errors.New("not a struct or pointer to struct")
	}

	numField := typElem.NumField()
	fieldInfoMap := make(map[string]*fieldInfo)
	byteCursor := uint(0)
	for i := 0; i < numField; i++ {
		field := typElem.Field(i)
		fieldType := typElem.Field(i).Type
		fieldVal := valElem.Field(i)
		fieldName := field.Name
		tagInfoEndian, tagInfoPadding, tagInfoSizeof, err := getTagInfo(field.Tag.Get("binary"))
		if err != nil {
			return nil, err
		}

		var tagSizeof []string // ignore tag `sizeof`
		tagSizeofVal := 0
		tagEndian := tagInfoEndian
		tagPadding := uint(0) // ignore tag `padding`
		byteSize := uint(0)
		byteSizeNeedResolve := false
		switch fieldType.Kind() {
		case reflect.Bool, reflect.Float32, reflect.Float64, reflect.Array:
			byteSize = uint(sizeof(fieldType))
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64: // support tag `sizeof`
			byteSize = uint(sizeof(fieldType))
			tagSizeof = tagInfoSizeof
			tagSizeofVal = int(fieldVal.Uint())
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // support tag `sizeof`
			byteSize = uint(sizeof(fieldType))
			if tagSizeofVal <= 0 {
				return nil, errors.New("tag sizeof value < 0")
			}
			tagSizeof = tagInfoSizeof
			tagSizeofVal = int(fieldVal.Int())
		case reflect.Slice:
			byteSize = uint(sizeof(fieldType.Elem()) * fieldVal.Len())
			byteSizeNeedResolve = true
		case reflect.Struct: // support tag `padding`
			if fieldType.NumField() > 0 {
				return nil, errors.New("embedded none empty struct not supported")
			}
			byteSize = tagInfoPadding
		default:
			return nil, errors.New("not supported kind")
		}
		byteStart := byteCursor
		byteEnd := byteStart + byteSize
		fieldInfoMap[fieldName] = &fieldInfo{
			name:                fieldName,
			tagEndian:           tagEndian,
			tagSizeof:           tagSizeof,
			tagSizeofVal:        uint(tagSizeofVal),
			tagPadding:          tagPadding,
			byteSizeNeedResolve: byteSizeNeedResolve,
			byteSize:            byteSize,
			byteStart:           byteStart,
			byteEnd:             byteEnd,
		}
		byteCursor += byteSize
	}

	return fieldInfoMap, nil
}

// return -1 if unknown
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
