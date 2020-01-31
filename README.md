# binary

## Usage

```go
packet := struct {
    Protocol uint16 `binary:"big"` // endian, `big` or `little`
    Length   uint16 `binary:"sizeof=Cmd+Data+Padding+CRC"` // sizeof (only support integer, others will be ignored), auto calculate size of fields
    Cmd      uint8
    Data     []byte
    Padding  struct{} `binary:"padding=1B"` // padding (only support empty struct, others will be ignored), specify the size of padding
    CRC      uint16
} {}

bytes, err := Pack(packet)
err = Unpack(&packet, bytes)
```

*tags:*

- big
- little
- sizeof
- size

*supported types:*

- bool
- byte
- uint
- uint8
- uint16
- uint32
- uint64
- int
- int8
- int16
- int32
- int64
- empty struct
- array or slice of above
