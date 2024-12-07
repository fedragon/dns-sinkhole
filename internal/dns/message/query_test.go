package message

import (
	"bufio"
	"bytes"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuery_Identification(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		expected uint16
	}{
		{
			name:     "extracts id from message",
			data:     []byte{0x00, 0xFF, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: 255,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m, err := UnmarshalQuery(c.data)
			assert.NoError(t, err)
			assert.Equal(t, c.expected, m.ID)
		})
	}
}

func TestQuery_OpCode(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		expected uint8
	}{
		{
			name:     "extracts min opcode from message",
			data:     []byte{0x00, 0xFF, 0x08, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: 1,
		},
		{
			name:     "extracts max opcode from message",
			data:     []byte{0x00, 0x00, 0x78, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: 15,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m, err := UnmarshalQuery(c.data)
			assert.NoError(t, err)
			assert.Equal(t, c.expected, m.OpCode)
		})
	}
}

func TestQuery_IsRecursionDesired(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "returns true for recursion desired",
			data:     []byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: true,
		},
		{
			name:     "returns false otherwise",
			data:     []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m, err := UnmarshalQuery(c.data)
			assert.NoError(t, err)
			assert.Equal(t, c.expected, m.RecursionDesired)
		})
	}
}

func TestQuery_parseQuestion(t *testing.T) {
	var question []byte
	question = append(question, 6)
	question = append(question, []byte("gemini")...)
	question = append(question, 3)
	question = append(question, []byte("tuc")...)
	question = append(question, 4)
	question = append(question, []byte("noao")...)
	question = append(question, 3)
	question = append(question, []byte("edu")...)
	question = append(question, 0)
	question = append(question, uint8(0), uint8(1))
	question = append(question, uint8(0), uint8(1))

	r := bufio.NewReader(bytes.NewReader(question))

	q, err := unmarshalQuestion(r)
	assert.NoError(t, err)
	assert.Equal(t, "gemini.tuc.noao.edu", q.Name)
	assert.Equal(t, TypeA, q.Type)
	assert.Equal(t, ClassInternetAddress, q.Class)
}

func TestQuestion_MarshalRoundtrip(t *testing.T) {
	q1 := Question{
		Name:  "www.federico.is",
		Type:  TypeA,
		Class: ClassInternetAddress,
	}

	data, err := marshalQuestion(q1)
	assert.NoError(t, err)
	q2, err := unmarshalQuestion(bufio.NewReader(bytes.NewReader(data)))
	assert.NoError(t, err)
	assert.Equal(t, q1, q2)
}

func TestRecord_A_MarshalRoundtrip(t *testing.T) {
	ip := netip.MustParseAddr("127.0.0.1").As4()
	r1 := Record{
		DomainName: "www.federico.is",
		Type:       TypeA,
		Class:      ClassInternetAddress,
		TTL:        60,
		Length:     uint16(len(ip)),
		Data:       ip[:],
	}

	data, err := marshalRecord(r1)
	assert.NoError(t, err)
	q2, err := unmarshalRecord(bufio.NewReader(bytes.NewReader(data)))
	assert.NoError(t, err)
	assert.Equal(t, r1, q2)
}

func TestRecord_AAAA_MarshalRoundtrip(t *testing.T) {
	ip := netip.MustParseAddr("::1").As16()
	r1 := Record{
		DomainName: "www.federico.is",
		Type:       TypeAAAA,
		Class:      ClassInternetAddress,
		TTL:        60,
		Length:     uint16(len(ip)),
		Data:       ip[:],
	}

	data, err := marshalRecord(r1)
	assert.NoError(t, err)
	q2, err := unmarshalRecord(bufio.NewReader(bytes.NewReader(data)))
	assert.NoError(t, err)
	assert.Equal(t, r1, q2)
}
