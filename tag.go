package binary

import (
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	tagItemKeyBigEndian    = "big"
	tagItemKeyLittleEndian = "little"
	tagItemKeySize         = "byteSize"
	tagItemKeySizeof       = "sizeof"

	reSize   = `^byteSize=(\d+)(B|W|DW|QW)$`
	reSizeof = `^sizeof=((\w+\+)*\w+)$`
)

func getTagSize(value string) (uint, error) {
	re := regexp.MustCompile(reSize)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 3 {
		return 0, errors.New("binary: invalid field tag byteSize value")
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}
	if num < 0 {
		return 0, errors.New("binary: tag byteSize value, must equal or greater than 0")
	}

	var bytes int
	switch matches[2] {
	case "B":
		bytes = num
	case "W":
		bytes = 2 * num
	case "DW":
		bytes = 4 * num
	case "QW":
		bytes = 8 * num
	}
	return uint(bytes), nil
}

func getTagSizeOf(value string) (fieldNames []string, err error) {
	re := regexp.MustCompile(reSizeof)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 3 {
		err = errors.New("binary: invalid field tag sizeof value")
		return
	}

	names := strings.Split(matches[1], "+")
	for _, fieldName := range names {
		field := strings.TrimSpace(fieldName)
		if field != "" {
			fieldNames = append(fieldNames, field)
		}
	}

	return
}

func getTagInfo(value string) (endian binary.ByteOrder, size uint, sizeof []string, err error) {
	itemValues := strings.Split(value, ",")
	for _, itemValue := range itemValues {
		itemValue = strings.TrimSpace(itemValue)
		if itemValue == "" {
			continue
		}

		if itemValue == tagItemKeyBigEndian {
			endian = binary.BigEndian
		} else if itemValue == tagItemKeyLittleEndian {
			endian = binary.LittleEndian
		} else if strings.HasPrefix(itemValue, tagItemKeySize+"=") {
			size, err = getTagSize(itemValue)
			if err != nil {
				return
			}
		} else if strings.HasPrefix(itemValue, tagItemKeySizeof+"=") {
			sizeof, err = getTagSizeOf(itemValue)
			if err != nil {
				return
			}
		} else {
			err = fmt.Errorf("invalid binary tag item: %s", itemValue)
			return
		}
	}
	if endian == nil {
		endian = binary.BigEndian
	}

	return
}
