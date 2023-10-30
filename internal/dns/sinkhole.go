package dns

import (
	"log/slog"
	"strings"

	"github.com/fedragon/sinkhole/internal/dns/message"
)

var (
	nonRoutableAddress = []byte{0x00, 0x2A} // "0.0.0.42"
)

type ResolveResult int

const (
	ResolveSuccess ResolveResult = iota
	UnresolvedNotFound
	UnresolvedNonStandard
	UnresolvedNonRecursive
	UnresolvedUnsupportedClass
	UnresolvedUnsupportedType
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
func (s *Sinkhole) Resolve(query *message.Query) (*message.Response, ResolveResult) {
	if query.OpCode() != 0 {
		return nil, UnresolvedNonStandard
	}

	if !query.IsRecursionDesired() {
		return nil, UnresolvedNonRecursive
	}

	var answers []message.Record
	for _, question := range query.Questions() {
		if question.Class != message.ClassInternetAddress {
			return nil, UnresolvedUnsupportedClass
		}

		if question.Type != message.TypeA {
			return nil, UnresolvedUnsupportedType
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
		return message.NewResponse(query, answers), ResolveSuccess
	}

	return nil, UnresolvedNotFound
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
