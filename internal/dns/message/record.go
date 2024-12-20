package message

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strings"
)

type Record struct {
	DomainName string
	Type       Type
	Class      Class
	TTL        uint32
	Length     uint16
	Data       []byte
}

func unmarshalRecord(r *bufio.Reader) (Record, error) {
	var parts []string

	for {
		label, err := r.ReadByte()
		if err != nil {
			return Record{}, err
		}

		if label == 0 {
			break
		}

		buf := make([]byte, label)
		_, err = io.ReadFull(r, buf)
		if err != nil {
			return Record{}, err
		}

		parts = append(parts, string(buf))
	}

	type_, err := read(r, 2)
	if err != nil {
		return Record{}, err
	}

	class, err := read(r, 2)
	if err != nil {
		return Record{}, err
	}

	ttl, err := read(r, 4)
	if err != nil {
		return Record{}, err
	}

	length, err := read(r, 2)
	if err != nil {
		return Record{}, err
	}

	data, err := read(r, int(byteOrder.Uint16(length)))
	if err != nil {
		return Record{}, err
	}

	return Record{
		DomainName: strings.Join(parts, "."),
		Type:       Type(byteOrder.Uint16(type_)),
		Class:      Class(byteOrder.Uint16(class)),
		TTL:        byteOrder.Uint32(ttl),
		Length:     byteOrder.Uint16(length),
		Data:       data,
	}, nil
}

func marshalRecord(r Record) ([]byte, error) {
	var data []byte
	parts := strings.Split(r.DomainName, ".")
	for _, part := range parts {
		length := len(part)
		if length > math.MaxUint8 {
			return nil, fmt.Errorf("substring length cannot be cast to uint8: %v", length)
		}
		data = append(data, uint8(length))
		data = append(data, []byte(part)...)
	}
	data = append(data, uint8(0))

	data = byteOrder.AppendUint16(data, uint16(r.Type))
	data = byteOrder.AppendUint16(data, uint16(r.Class))
	data = byteOrder.AppendUint32(data, r.TTL)
	data = byteOrder.AppendUint16(data, r.Length)

	return append(data, r.Data...), nil
}
