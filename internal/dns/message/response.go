package message

import "fmt"

type Response struct {
	id        uint16
	flags     uint16
	questions []Question
	answers   []Record
}

func NewResponse(query *Query, answers []Record) *Response {
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
		answers:   answers,
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
