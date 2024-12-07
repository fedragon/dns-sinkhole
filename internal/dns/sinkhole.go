package dns

import (
	"log/slog"
	"net/netip"
	"strconv"

	p "github.com/prometheus/client_golang/prometheus"

	"github.com/fedragon/sinkhole/internal/dns/message"
	"github.com/fedragon/sinkhole/internal/metrics"
)

var (
	nonRoutableAddress     = netip.MustParseAddr("0.0.0.42")
	NonRoutableAddressIPv4 = nonRoutableAddress.As4()
	NonRoutableAddressIPv6 = nonRoutableAddress.As16()
)

// Sinkhole is a DNS server that receives queries and, if they are related to domains belonging to its internal registry, resolves them to non-routable addresses.
type Sinkhole struct {
	registry map[string]struct{}
	logger   *slog.Logger
}

func NewSinkhole(logger *slog.Logger) *Sinkhole {
	return &Sinkhole{
		registry: make(map[string]struct{}),
		logger:   logger.With("source", "sinkhole"),
	}
}

// Register registers a domain with the sinkhole.
func (s *Sinkhole) Register(domain string) {
	s.registry[domain] = struct{}{}
}

// Resolve resolves a query to a non-routable address, if the domain belongs to its registry.
func (s *Sinkhole) Resolve(query *message.Query) (*message.Response, bool) {
	if query.OpCode != 0 {
		metrics.UnsupportedOpCodeQueries.With(p.Labels{"opcode": strconv.Itoa(int(query.OpCode))}).Inc()
		return nil, false
	}

	if !query.RecursionDesired {
		metrics.NonRecursiveQueries.Inc()
		return nil, false
	}

	question := query.Question
	if question.Class != message.ClassInternetAddress {
		metrics.UnsupportedClassQueries.With(p.Labels{"class": strconv.Itoa(int(question.Class))}).Inc()
		return nil, false
	}

	if question.Type != message.TypeA && question.Type != message.TypeAAAA {
		metrics.UnsupportedTypeQueries.With(p.Labels{"type": strconv.Itoa(int(question.Type))}).Inc()
		return nil, false
	}

	metrics.SupportedQueries.With(p.Labels{"type": strconv.Itoa(int(question.Type))}).Inc()

	if s.Contains(question.Name) {
		answer := message.Record{
			DomainName: question.Name,
			Class:      message.ClassInternetAddress,
			TTL:        3600,
		}

		if question.Type == message.TypeA {
			answer.Type = message.TypeA
			answer.Data = NonRoutableAddressIPv4[:]
			answer.Length = 4
		} else {
			answer.Type = message.TypeAAAA
			answer.Data = NonRoutableAddressIPv6[:]
			answer.Length = 16
		}

		return message.NewResponse(query, answer), true
	}

	return nil, false
}

// Contains returns true if the domain belongs to the sinkhole's registry
func (s *Sinkhole) Contains(domain string) bool {
	timer := p.NewTimer(metrics.ResponseTimesInternalResolve)
	defer timer.ObserveDuration()

	_, ok := s.registry[domain]
	return ok
}
