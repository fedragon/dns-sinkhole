package dns

import (
	"log/slog"
	"strconv"
	"strings"

	p "github.com/prometheus/client_golang/prometheus"

	"github.com/fedragon/sinkhole/internal/dns/message"
	"github.com/fedragon/sinkhole/internal/metrics"
)

var (
	nonRoutableAddress = []byte{0x00, 0x2A} // "0.0.0.42"
)

// Sinkhole is a DNS server that receives queries and, if they are related to domains belonging to its internal registry, resolves them to non-routable addresses.
type Sinkhole struct {
	registry map[string]*Domain
	logger   *slog.Logger
}

func NewSinkhole(logger *slog.Logger) *Sinkhole {
	return &Sinkhole{
		registry: make(map[string]*Domain),
		logger:   logger.With("source", "sinkhole"),
	}
}

// Register registers a domain with the sinkhole.
func (s *Sinkhole) Register(domain string) error {
	d, err := NewDomain(domain)
	if err != nil {
		return err
	}

	found, ok := s.registry[d.name]
	if !ok {
		s.registry[d.name] = d
		return nil
	}

	return found.Register(domain)
}

// Resolve resolves a query to a non-routable address, if the domain belongs to its registry.
func (s *Sinkhole) Resolve(query *message.Query) (*message.Response, bool) {
	if query.OpCode() != 0 {
		metrics.UnsupportedOpCodeQueries.With(p.Labels{"opcode": strconv.Itoa(int(query.OpCode()))}).Inc()
		return nil, false
	}

	if !query.IsRecursionDesired() {
		metrics.NonRecursiveQueries.Inc()
		return nil, false
	}

	var answers []message.Record
	question := query.Question()
	if question.Class != message.ClassInternetAddress {
		metrics.UnsupportedClassQueries.With(p.Labels{"class": strconv.Itoa(int(question.Class))}).Inc()
		return nil, false
	}

	if question.Type != message.TypeA {
		metrics.UnsupportedTypeQueries.With(p.Labels{"type": strconv.Itoa(int(question.Type))}).Inc()
		return nil, false
	}

	metrics.SupportedQueries.With(p.Labels{"type": strconv.Itoa(int(question.Type))}).Inc()

	if s.Contains(question.Name) {
		answer := message.Record{
			DomainName: question.Name,
			Type:       message.TypeA,
			Class:      message.ClassInternetAddress,
			TTL:        3600,
			Length:     4,
			Data:       nonRoutableAddress,
		}

		answers = append(answers, answer)
	}

	if len(answers) > 0 {
		return message.NewResponse(query, answers), true
	}

	return nil, false
}

// Contains returns true if the domain belongs to the sinkhole's registry
func (s *Sinkhole) Contains(domain string) bool {
	idx := strings.LastIndex(domain, ".")
	if idx == -1 {
		return false
	}

	d, ok := s.registry[domain[idx+1:]]
	if !ok {
		return false
	}

	return d.Contains(domain)
}
