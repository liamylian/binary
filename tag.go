package binary

import (
	"encoding/binary"
	"errors"
	"strconv"
	"strings"
)

const (
	tagItemKeyBigEndian    = "big"
	tagItemKeyLittleEndian = "little"
	tagItemKeyPadding      = "padding"
	tagItemKeySizeof       = "sizeof"
)

type tagItemEndian struct {
	Endian binary.ByteOrder
}

type tagItemPadding struct {
	Bytes int
}

func newTagItemPadding(value string) (*tagItemPadding, error) {
	value = strings.TrimRight(value, "B")
	value = strings.Replace(value, tagItemKeyPadding+"=", "", 1)
	bytes, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	return &tagItemPadding{bytes}, nil
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

func getTagItems(tagValue string) (endian *tagItemEndian, padding *tagItemPadding, sizeof *tagItemSizeOf, err error) {
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
		} else if strings.HasPrefix(itemValue, tagItemKeyPadding) {
			tagItem, err := newTagItemPadding(itemValue)
			if err != nil {
				return nil, nil, nil, err
			}
			padding = tagItem
		} else if strings.HasPrefix(itemValue, tagItemKeySizeof) {
			tagItem, err := newTagItemSizeOf(itemValue)
			if err != nil {
				return nil, nil, nil, err
			}
			sizeof = tagItem
		} else {
			return nil, nil, nil, errors.New("invalid tag itemValue")
		}
	}

	return
}
