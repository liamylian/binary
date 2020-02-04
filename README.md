# binary

## Usage

```go
packet := struct {
    Protocol uint16 `binary:"big"`
    Length   uint16 `binary:"sizeof=Cmd+Data+Padding+CRC"` 
    Cmd      uint8
    Data     []byte
    Padding  [2]byte
    CRC      uint16
} {}

bytes, err := Pack(packet)
err = Unpack(&packet, bytes)
```

*tags:*

- `big`, big endian
- `little`, little endian
- `sizeof`, auto calculate size of fields, and put result into field (field must be integer).

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
- array
- slice
