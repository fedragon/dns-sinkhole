package dns

import (
	"log/slog"
	"strings"

	"github.com/fedragon/sinkhole/internal/dns/message"
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
	d, ok := s.domains[domain]
	if !ok {
		d, err := NewDomain(domain)
		if err != nil {
			return err
		}
		s.domains[d.name] = d
		return nil
	}

	return d.Register(domain)
}

func (s *Sinkhole) Handle(query *message.Query) (*message.Response, bool) {
	if !query.IsRecursionDesired() {
		s.logger.Debug("passing non-recursion query to fallback DNS server", "query", query)
		return nil, false
	}

	if len(query.Questions()) != 1 {
		s.logger.Debug("passing query with multiple questions to fallback DNS", "query", query)
		return nil, false
	}

	question := query.Questions()[0]
	if question.Type != 1 || question.Class != 1 {
		s.logger.Debug("passing query with non-A questions to fallback DNS", "query", query)
		return nil, false
	}

	idx := strings.LastIndex(question.Name, ".")
	if idx == -1 {
		return nil, false
	}

	d, ok := s.domains[question.Name[idx+1:]]
	if !ok {
		return nil, false
	}

	if d.Contains(question.Name) {
		answer := message.Record{
			DomainName: question.Name,
			Type:       message.TypeA,
			Class:      message.ClassInternetAddress,
			TTL:        60,
			Length:     4,
			Data:       nonRoutableAddress,
		}

		return message.NewResponse(query, answer), true
	}

	return nil, false
}
