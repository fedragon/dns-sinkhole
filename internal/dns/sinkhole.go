package dns

import (
	"log/slog"

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
		var err error
		d, err = NewDomain(domain)
		if err != nil {
			return err
		}
	}

	return d.Register(domain)
}

func (s *Sinkhole) Handle(query *message.Query) (*message.Response, bool) {
	if !query.IsRecursionDesired() {
		s.logger.Debug("passing non-recursion query to fallback DNS server", "query", query)
		return nil, false
	}

	if len(query.Questions()) == 0 {
		s.logger.Debug("no questions in query", "query", query)
		return nil, false
	}

	if len(query.Questions()) > 0 {
		s.logger.Debug("currently unable to answer multiple questions in a single query", "query", query)
		return nil, false
	}

	question := query.Questions()[0]
	if question.Type != 1 || question.Class != 1 {
		s.logger.Debug("currently only able to answer A-type questions", "query", query)
		return nil, false
	}

	d, ok := s.domains[question.Name]
	if !ok {
		return nil, false
	}

	var response *message.Response
	if d.Contains(question.Name) {
		answer := message.Record{
			DomainName: question.Name,
			Type:       message.TypeA,
			Class:      message.ClassInternetAddress,
			TTL:        60,
			Length:     uint16(len(nonRoutableAddress)),
			Data:       nonRoutableAddress,
		}

		response = message.BuildResponse(query, answer)
	}

	return response, true
}
