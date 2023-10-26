package message

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strings"
)

type Question struct {
	Name  string
	Type  Type
	Class Class
}

func unmarshalQuestion(r *bufio.Reader) (Question, error) {
	var parts []string

	for {
		label, err := r.ReadByte()
		if err != nil {
			return Question{}, err
		}

		if label == 0 {
			break
		}

		buf := make([]byte, label)
		_, err = io.ReadFull(r, buf)
		if err != nil {
			return Question{}, err
		}

		parts = append(parts, string(buf))
	}

	type_, err := read(r, 2)
	if err != nil {
		return Question{}, err
	}

	class, err := read(r, 2)
	if err != nil {
		return Question{}, err
	}

	return Question{
		Name:  strings.Join(parts, "."),
		Type:  Type(byteOrder.Uint16(type_)),
		Class: Class(byteOrder.Uint16(class)),
	}, nil
}

func (q Question) marshal() ([]byte, error) {
	var data []byte
	parts := strings.Split(q.Name, ".")
	for _, part := range parts {
		length := len(part)
		if length > math.MaxUint8 {
			return nil, fmt.Errorf("substring length cannot be cast to uint8: %v", length)
		}
		data = append(data, uint8(length))
		data = append(data, []byte(part)...)
	}
	data = append(data, uint8(0))

	data = byteOrder.AppendUint16(data, uint16(q.Type))
	data = byteOrder.AppendUint16(data, uint16(q.Class))

	return data, nil
}
