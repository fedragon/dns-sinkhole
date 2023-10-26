package message

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type Class uint16
type Type uint16

const (
	ClassInternetAddress Class = 1

	TypeA Type = 1

	queryMask            = 0b1000_0000_0000_0000
	opCodeMask           = 0b0111_1000_0000_0000
	recursionDesiredMask = 0b0000_0001_0000_0000
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

func ParseQuery(data []byte) (*Query, error) {
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

	questions, err := parseQuestions(r, numQuestions)
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

func parseQuestions(r *bufio.Reader, n uint16) ([]Question, error) {
	var questions []Question

	for i := 0; i < int(n); i++ {
		q, err := unmarshalQuestion(r)
		if err != nil {
			return nil, err
		}

		questions = append(questions, q)
	}

	return questions, nil
}

type Response struct {
	id        uint16
	flags     uint16
	questions []Question
	answers   []Record
}

func NewResponse(query *Query, answer Record) *Response {
	var flags uint16
	flags |= 1 << 15 // QueryResponse: 1 for Response
	if query.IsRecursionDesired() {
		flags |= 1 << 8 // RecursionDesired: 1
	}
	flags |= 1 << 7 // RecursionAvailable: 1

	res := &Response{
		id:        query.id,
		flags:     flags,
		questions: query.Questions(),
		answers:   []Record{answer},
	}

	return res
}

func (m *Response) Marshal() ([]byte, error) {
	var data []byte
	data = byteOrder.AppendUint16(data, m.id)
	data = byteOrder.AppendUint16(data, m.flags)
	data = byteOrder.AppendUint16(data, uint16(len(m.questions)))
	data = byteOrder.AppendUint16(data, uint16(len(m.answers)))
	data = byteOrder.AppendUint16(data, 0) // number of authoritative records
	data = byteOrder.AppendUint16(data, 0) // number of additional records

	for _, q := range m.questions {
		buf, err := q.marshal()
		if err != nil {
			return nil, err
		}
		data = append(data, buf...)
	}

	for _, r := range m.answers {
		buf, err := r.marshal()
		if err != nil {
			return nil, err
		}
		data = append(data, buf...)
	}

	if len(data) > 512 {
		return nil, fmt.Errorf("response does not fit in 512 bytes. length: %v", len(data))
	}

	return data, nil
}
