package message

import (
	"fmt"
)

type Response struct {
	id        uint16
	flags     uint16
	questions []Question
	Answers   []Record
}

func NewResponse(query *Query, answer Record) *Response {
	var flags uint16
	flags |= 1 << 15 // QueryResponse: 1 for Response
	if query.RecursionDesired {
		flags |= 1 << 8 // RecursionDesired: 1
	}
	flags |= 1 << 7 // RecursionAvailable: 1

	res := &Response{
		id:        query.ID,
		flags:     flags,
		questions: []Question{query.Question},
		Answers:   []Record{answer},
	}

	return res
}

func (r *Response) ID() uint16 {
	return r.id
}

func (r *Response) IsRecursionDesired() bool {
	return (r.flags&recursionDesiredMask)>>8 == 1
}

func (r *Response) IsRecursionAvailable() bool {
	return (r.flags&recursionAvailableMask)>>7 == 1
}

func MarshalResponse(r *Response) ([]byte, error) {
	var data []byte
	data = byteOrder.AppendUint16(data, r.id)
	data = byteOrder.AppendUint16(data, r.flags)
	data = byteOrder.AppendUint16(data, uint16(len(r.questions)))
	data = byteOrder.AppendUint16(data, uint16(len(r.Answers)))
	data = byteOrder.AppendUint16(data, 0) // number of authoritative records
	data = byteOrder.AppendUint16(data, 0) // number of additional records

	for _, q := range r.questions {
		buf, err := marshalQuestion(q)
		if err != nil {
			return nil, err
		}
		data = append(data, buf...)
	}

	for _, r := range r.Answers {
		buf, err := marshalRecord(r)
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
