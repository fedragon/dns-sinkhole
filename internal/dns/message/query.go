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

	TypeA Type = 1

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
	id        uint16
	flags     uint16
	questions []Question
}

func (q *Query) ID() uint16 {
	return q.id
}

func (q *Query) OpCode() uint8 {
	return uint8((q.flags & opCodeMask) >> 11)
}

func (q *Query) IsRecursionDesired() bool {
	return (q.flags&recursionDesiredMask)>>8 == 1
}

func (q *Query) Questions() []Question {
	return q.questions
}

func (q *Query) Type() Type {
	return q.questions[0].Type
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

	// skip the next 6 bytes (number of answers, authorities, and additional records - all zeroed in queries)
	_, err = read(r, 6)
	if err != nil {
		return nil, err
	}

	questions, err := unmarshalQuestions(r, numQuestions)
	if err != nil {
		return nil, err
	}

	return &Query{
		id:        identification,
		flags:     flags,
		questions: questions,
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
