package dns

import (
	"log/slog"
	"strings"

	"github.com/fedragon/sinkhole/internal/dns/message"
	"github.com/fedragon/sinkhole/internal/metrics"
)

var (
	nonRoutableAddress = []byte{0x00, 0x2A} // "0.0.0.42"
)

type Sinkhole struct {
	domains map[string]*Domain
	logger  *slog.Logger
}

func NewSinkhole(logger *slog.Logger) *Sinkhole {
	return &Sinkhole{
		domains: make(map[string]*Domain),
		logger:  logger.With("source", "sinkhole"),
	}
}

func (s *Sinkhole) Register(domain string) error {
	d, err := NewDomain(domain)
	if err != nil {
		return err
	}

	found, ok := s.domains[d.name]
	if !ok {
		s.domains[d.name] = d
		return nil
	}

	return found.Register(domain)
}

func (s *Sinkhole) Resolve(query *message.Query) (*message.Response, bool) {
	if query.OpCode() != 0 {
		metrics.NonStandardQueries.Inc()
		s.logger.Debug("Passing non-standard query to fallback DNS resolver", "query", query)
		return nil, false
	}

	if !query.IsRecursionDesired() {
		metrics.NonRecursiveQueries.Inc()
		s.logger.Debug("Passing non-recursive query to fallback DNS resolver", "query", query)
		return nil, false
	}

	var answers []message.Record
	for _, question := range query.Questions() {
		if question.Class != message.ClassInternetAddress {
			metrics.UnsupportedClassQueries.Inc()
			s.logger.Debug("Passing query with non-Internet class to fallback DNS resolver", "query", query)
			return nil, false
		}

		if question.Type != message.TypeA {
			metrics.UnsupportedTypeQueries.Inc()
			s.logger.Debug("Passing query with non-A type to fallback DNS resolver", "query", query)
			return nil, false
		}

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
	}

	if len(answers) > 0 {
		return message.NewResponse(query, answers), true
	}

	return nil, false
}

func (s *Sinkhole) Contains(domain string) bool {
	idx := strings.LastIndex(domain, ".")
	if idx == -1 {
		return false
	}

	d, ok := s.domains[domain[idx+1:]]
	if !ok {
		return false
	}

	return d.Contains(domain)
}
