package binary

import (
	"bytes"
	"encoding/binary"
	"errors"
	"reflect"
)

func Pack(v interface{}) ([]byte, error) {
	typElem := reflect.TypeOf(v)
	valElem := reflect.ValueOf(v)
	if valElem.Kind() == reflect.Ptr {
		typElem = typElem.Elem()
		valElem = valElem.Elem()
	}

	if typElem.Kind() != reflect.Struct {
		return nil, errors.New("not a struct or pointer to struct")
	}

	fieldInfoMap, err := getFieldInfo(v)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(nil)
	numField := typElem.NumField()
	for i := 0; i < numField; i++ {
		field := typElem.Field(i)
		fieldValue := valElem.Field(i)
		fieldInfo := fieldInfoMap[field.Name]

		var bytes []byte
		if fieldValue.Kind() == reflect.Struct {
			bytes = make([]byte, fieldInfo.byteSize)
		} else {
			bytes, err = pack(fieldValue, fieldInfo.tagEndian, fieldInfo.byteSize)
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
		if len(fieldInfo.tagSizeof) == 0 {
			continue
		}

		sumSize := uint64(0)
		sumSizeLen := uint(8)
		for _, otherFieldName := range fieldInfo.tagSizeof {
			otherFieldInfo := fieldInfoMap[otherFieldName]
			sumSize += uint64(otherFieldInfo.byteSize)
		}
		sumSizeBytes := make([]byte, sumSizeLen)
		fieldInfo.tagEndian.PutUint64(sumSizeBytes, sumSize)
		copy(resultBytes[fieldInfo.byteStart:], sumSizeBytes[sumSizeLen-fieldInfo.byteSize:])
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
			return append(padding, bytes...), nil // todo endian problem
		} else {
			return bytes[byteSize-size:], nil
		}
	default:
		return nil, errors.New("not supported kind")
	}
}

func Unpack(v interface{}, bytes []byte) error {
	val := reflect.ValueOf(v)
	if !(val.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Struct) {
		return errors.New("not a pointer to struct")
	}
	valElem := val.Elem()

	fieldInfoMap, err := getFieldInfo(v)
	if err != nil {
		return err
	}

	// check size
	totalBytes := 0
	for _, fieldInfo := range fieldInfoMap {
		if fieldInfo.byteSize < 0 {
			return errors.New("variant bytes before sizeof")
		}
		totalBytes += int(fieldInfo.byteSize)
		if len(fieldInfo.tagSizeof) == 0 {
			continue
		}

		fieldValueMap := make(map[string]int)
		for _, otherFieldName := range fieldInfo.tagSizeof {
			otherFieldInfo := fieldInfoMap[otherFieldName]
			if otherFieldInfo.typeSize >= 0 {
				fieldValueMap[otherFieldInfo.name] = otherFieldInfo.typeSize
			} else {
				fieldValueMap[otherFieldInfo.name] = -1 // need solve
			}
		}

		field := valElem.FieldByName(fieldInfo.name)
		length := uint64(0)
		switch field.Kind() {
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			valBytes := bytes[fieldInfo.byteStart:fieldInfo.byteEnd]
			if fieldInfo.byteSize < 8 {
				valBytes = append(make([]byte, 8-fieldInfo.byteSize), valBytes...) // todo endian problem
			}
			length = fieldInfo.tagEndian.Uint64(valBytes)
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		}
		solvedField, value, err := solveSum(fieldValueMap, int(length))
		if err != nil {
			return err
		}
		if solvedField != "" {
			fieldInfoMap[solvedField].typeSize = value
			fieldInfoMap[solvedField].byteSize = uint(value)
			totalBytes += value
		}
		break
	}
	if totalBytes != len(bytes) {
		return errors.New("mismatch bytes")
	}

	return nil
}
