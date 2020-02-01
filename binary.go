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
			bytes, err = pack(fieldValue, fieldInfo.tagEndian)
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

func pack(v reflect.Value, endian binary.ByteOrder) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	if err := binary.Write(buffer, endian, v.Interface()); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func Unpack(v interface{}, bytes []byte) error {
	typ := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if !(val.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Struct) {
		return errors.New("not a pointer to struct")
	}
	valElem := val.Elem()
	typElem := typ.Elem()
	fieldInfoMap, err := getFieldInfo(v)
	if err != nil {
		return err
	}

	// solve size
	numField := typElem.NumField()
	for i := 0; i < numField; i++ {
		fieldInfo := fieldInfoMap[typElem.Field(i).Name]
		if fieldInfo.byteSize < 0 {
			return errors.New("variant bytes before sizeof")
		}
		if len(fieldInfo.tagSizeof) == 0 {
			continue
		}

		fieldValueMap := make(map[string]int)
		for _, otherFieldName := range fieldInfo.tagSizeof {
			otherFieldInfo := fieldInfoMap[otherFieldName]
			if otherFieldInfo.byteSizeNeedResolve {
				fieldValueMap[otherFieldInfo.name] = -1 // need solve
			} else {
				fieldValueMap[otherFieldInfo.name] = int(otherFieldInfo.byteSize)
			}
		}

		field := valElem.FieldByName(fieldInfo.name)
		length := uint64(0)
		switch field.Kind() {
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			valBytes := bytes[fieldInfo.byteStart:fieldInfo.byteEnd]
			val, err := byte2Uint(fieldInfo.tagEndian, valBytes)
			if err != nil {
				return err
			}
			length = uint64(val)
		}
		solvedField, value, err := solveSum(fieldValueMap, int(length))
		if err != nil {
			return err
		}
		if solvedField != "" {
			fieldInfoMap[solvedField].byteSizeNeedResolve = false
			fieldInfoMap[solvedField].byteSize = uint(value)
		}
		break
	}

	// check size
	totalBytes := 0
	byteCursor := 0
	for i := 0; i < numField; i++ {
		fieldInfo := fieldInfoMap[typElem.Field(i).Name]
		fieldInfo.byteStart = uint(byteCursor)
		fieldInfo.byteEnd = fieldInfo.byteStart + fieldInfo.byteSize
		totalBytes += int(fieldInfo.byteSize)
		byteCursor += int(fieldInfo.byteSize)
	}
	if totalBytes != len(bytes) {
		return errors.New("mismatch bytes")
	}

	// unpack
	for i := 0; i < numField; i++ {
		fieldInfo := fieldInfoMap[typElem.Field(i).Name]
		field := valElem.Field(i)
		if field.Kind() == reflect.Struct {
			continue
		}
		if err := unpack(field, bytes[fieldInfo.byteStart:fieldInfo.byteEnd], fieldInfo.tagEndian); err != nil {
			return err
		}
	}

	return nil
}

func unpack(v reflect.Value, data []byte, endian binary.ByteOrder) error {
	buffer := bytes.NewBuffer(data)
	if v.Kind() == reflect.Slice {
		eleSize := v.Type().Elem().Size()
		sliceLen := len(data) / int(eleSize)
		slice := reflect.MakeSlice(v.Type(), sliceLen, sliceLen)
		v.Addr().Elem().Set(slice)
	}

	err := binary.Read(buffer, endian, v.Addr().Interface())
	return err
}
