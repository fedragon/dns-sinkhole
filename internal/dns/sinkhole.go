package dns

import (
	"fmt"

	"github.com/fedragon/sinkhole/internal/dns/message"
)

var (
	nonRoutableAddress = []byte{0x00, 0x2A} // "0.0.0.42"
)

type Sinkhole struct {
	domains map[string]*Domain
}

func NewSinkhole() *Sinkhole {
	return &Sinkhole{
		domains: make(map[string]*Domain),
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
		fmt.Printf("passing non-recursion query to fallback DNS server. query: %v\n", query)
		return nil, false
	}

	if len(query.Questions()) == 0 {
		fmt.Printf("no questions in query. query: %v\n", query)
		return nil, false
	}

	if len(query.Questions()) > 0 {
		fmt.Printf("currently unable to answer multiple questions in a single query. query: %v\n", query)
		return nil, false
	}

	question := query.Questions()[0]
	if question.Type != 1 || question.Class != 1 {
		fmt.Printf("currently only able to answer A-type questions. query: %v\n", query)
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
