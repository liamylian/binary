package binary

import (
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	tagItemKeyBigEndian    = "big"
	tagItemKeyLittleEndian = "little"
	tagItemKeySizeof       = "sizeof"

	reSizeof = `^sizeof=((\w+\+)*\w+)$`
)

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

func getTagInfo(value string) (endian binary.ByteOrder, sizeof []string, err error) {
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
