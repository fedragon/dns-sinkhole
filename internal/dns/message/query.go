package message

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

type Class uint16
type Type uint16

const (
	ClassInternetAddress Class = 1

	TypeA    Type = 1
	TypeAAAA Type = 28

	queryMask              = 0b1000_0000_0000_0000
	opCodeMask             = 0b0111_1000_0000_0000
	recursionDesiredMask   = 0b0000_0001_0000_0000
	recursionAvailableMask = 0b0000_0000_1000_0000
)

var (
	ErrTooShort = errors.New("message too short")
	byteOrder   = binary.BigEndian
)

type Query struct {
	ID               uint16
	OpCode           uint8
	RecursionDesired bool
	Question         Question
}

func UnmarshalQuery(data []byte) (*Query, error) {
	if len(data) < 12 {
		return nil, ErrTooShort
	}

	r := bufio.NewReader(bytes.NewReader(data))

	id, err := read(r, 2)
	if err != nil {
		return nil, err
	}
	identification := byteOrder.Uint16(id)

	fs, err := read(r, 2)
	if err != nil {
		return nil, err
	}
	flags := byteOrder.Uint16(fs)
	if flags&queryMask>>15 != 0 {
		return nil, errors.New("not a query")
	}

	qs, err := read(r, 2)
	if err != nil {
		return nil, err
	}
	numQuestions := byteOrder.Uint16(qs)
	if numQuestions == 0 {
		return nil, errors.New("no questions")
	}

	if numQuestions > 1 {
		return nil, errors.New("too many questions")
	}

	// skip the next 6 bytes (number of answers, authorities, and additional records - all zeroed in queries)
	_, err = read(r, 6)
	if err != nil {
		return nil, err
	}

	question, err := unmarshalQuestion(r)
	if err != nil {
		return nil, err
	}

	return &Query{
		ID:               identification,
		OpCode:           uint8((flags & opCodeMask) >> 11),
		RecursionDesired: (flags&recursionDesiredMask)>>8 == 1,
		Question:         question,
	}, nil
}

func read(r io.Reader, n int) ([]byte, error) {
	res := make([]byte, n)
	_, err := io.ReadFull(r, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
