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
	tagItemKeySize         = "size"
	tagItemKeySizeof       = "sizeof"
)

type tagItemEndian struct {
	Endian binary.ByteOrder
}

type tagItemSize struct {
	Bytes int
}

func newTagItemSize(value string) (*tagItemSize, error) {
	re := regexp.MustCompile(`size=(\d+)(B|W|DW|QW)`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 3 {
		return nil, errors.New("binary: invalid size value")
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, err
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

	return &tagItemSize{bytes}, nil
}

type tagItemSizeOf struct {
	Fields []string
}

func newTagItemSizeOf(value string) (*tagItemSizeOf, error) {
	tagItem := &tagItemSizeOf{}
	fields := strings.Split(value, "+")
	for _, field := range fields {
		field := strings.TrimSpace(field)
		if field != "" {
			tagItem.Fields = append(tagItem.Fields, field)
		}
	}
	if len(tagItem.Fields) == 0 {
		return nil, errors.New("tag item sizeof has no fields")
	}

	return tagItem, nil
}

func getTagItems(tagValue string) (endian *tagItemEndian, size *tagItemSize, sizeof *tagItemSizeOf, err error) {
	itemValues := strings.Split(tagValue, ",")
	for _, itemValue := range itemValues {
		itemValue = strings.TrimSpace(itemValue)
		if itemValue == "" {
			continue
		}

		if itemValue == tagItemKeyBigEndian {
			endian = &tagItemEndian{binary.BigEndian}
		} else if itemValue == tagItemKeyLittleEndian {
			endian = &tagItemEndian{binary.LittleEndian}
		} else if strings.HasPrefix(itemValue, tagItemKeySize+"=") {
			tagItem, err := newTagItemSize(itemValue)
			if err != nil {
				return nil, nil, nil, err
			}
			size = tagItem
		} else if strings.HasPrefix(itemValue, tagItemKeySizeof+"=") {
			tagItem, err := newTagItemSizeOf(itemValue)
			if err != nil {
				return nil, nil, nil, err
			}
			sizeof = tagItem
		} else {
			return nil, nil, nil, fmt.Errorf("invalid tag item: %s", itemValue)
		}
	}

	return
}
