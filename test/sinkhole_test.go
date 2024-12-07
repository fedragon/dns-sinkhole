package test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fedragon/sinkhole/internal/dns"
	"github.com/fedragon/sinkhole/internal/dns/message"
)

func TestSinkhole_Contains(t *testing.T) {
	sut := dns.NewSinkhole(slog.Default())
	blockedDomain := "xxx.yyy"

	sut.Register(blockedDomain)

	assert.True(t, sut.Contains(blockedDomain))
	assert.False(t, sut.Contains("federico.is"))
}

func TestSinkhole_Resolve(t *testing.T) {
	blockedDomain := "xxx.yyy"
	sut := dns.NewSinkhole(slog.Default())

	sut.Register(blockedDomain)

	query := message.Query{
		ID:               1,
		OpCode:           0,
		RecursionDesired: true,
		Question: message.Question{
			Name:  "federico.is",
			Type:  message.TypeAAAA,
			Class: message.ClassInternetAddress,
		},
	}
	res, ok := sut.Resolve(&query)
	assert.False(t, ok)
	assert.Nil(t, res)

	query = message.Query{
		ID:               2,
		OpCode:           0,
		RecursionDesired: true,
		Question: message.Question{
			Name:  blockedDomain,
			Type:  message.TypeA,
			Class: message.ClassInternetAddress,
		},
	}

	res, ok = sut.Resolve(&query)
	assert.True(t, ok)
	assert.EqualValues(t, 2, res.ID())
	assert.Len(t, res.Answers, 1)
	assert.EqualValues(t, message.TypeA, res.Answers[0].Type)
	assert.EqualValues(t, message.ClassInternetAddress, res.Answers[0].Class)
	assert.EqualValues(t, blockedDomain, res.Answers[0].DomainName)
	assert.EqualValues(t, dns.NonRoutableAddressIPv4[:], res.Answers[0].Data)
	assert.EqualValues(t, 4, res.Answers[0].Length)

	query = message.Query{
		ID:               2,
		OpCode:           0,
		RecursionDesired: true,
		Question: message.Question{
			Name:  blockedDomain,
			Type:  message.TypeAAAA,
			Class: message.ClassInternetAddress,
		},
	}

	res, ok = sut.Resolve(&query)
	assert.True(t, ok)
	assert.EqualValues(t, 2, res.ID())
	assert.Len(t, res.Answers, 1)
	assert.EqualValues(t, message.TypeAAAA, res.Answers[0].Type)
	assert.EqualValues(t, message.ClassInternetAddress, res.Answers[0].Class)
	assert.EqualValues(t, blockedDomain, res.Answers[0].DomainName)
	assert.EqualValues(t, dns.NonRoutableAddressIPv6[:], res.Answers[0].Data)
	assert.EqualValues(t, 16, res.Answers[0].Length)
}
