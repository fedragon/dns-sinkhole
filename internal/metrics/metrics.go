package metrics

import (
	p "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	NonRoutableDomains = promauto.NewGauge(
		p.GaugeOpts{
			Namespace: "sinkhole",
			Name:      "non_routable_domains_total",
			Help:      "The total number of domains that we don't want to resolve",
		})

	queries = promauto.NewCounterVec(
		p.CounterOpts{
			Namespace: "sinkhole",
			Name:      "queries_total",
			Help:      "The total number of queries",
		},
		[]string{"blocked"})

	BlockedQueries  = queries.With(p.Labels{"blocked": "true"})
	UpstreamQueries = queries.With(p.Labels{"blocked": "false"})

	NonRecursiveQueries = promauto.NewCounter(
		p.CounterOpts{
			Namespace: "sinkhole",
			Name:      "non_recursive_queries_total",
			Help:      "The total number of queries that were discarded due to being non-recursive",
		})

	UnsupportedOpCodeQueries = promauto.NewCounterVec(
		p.CounterOpts{
			Namespace: "sinkhole",
			Name:      "unsupported_op_code_queries_total",
			Help:      "The total number of queries with unsupported opcode",
		},
		[]string{"opcode"})

	UnsupportedClassQueries = promauto.NewCounterVec(
		p.CounterOpts{
			Namespace: "sinkhole",
			Name:      "unsupported_class_queries_total",
			Help:      "The total number of queries with unsupported class",
		},
		[]string{"class"})

	UnsupportedTypeQueries = promauto.NewCounterVec(
		p.CounterOpts{
			Namespace: "sinkhole",
			Name:      "unsupported_type_queries_total",
			Help:      "The total number of queries with unsupported type",
		},
		[]string{"type"})

	QueryParsingErrors = promauto.NewCounter(
		p.CounterOpts{
			Namespace: "sinkhole",
			Name:      "query_parsing_errors_total",
			Help:      "The total number of query parsing errors",
		},
	)

	ResponseMarshallingErrors = promauto.NewCounter(
		p.CounterOpts{
			Namespace: "sinkhole",
			Name:      "response_marshalling_errors_total",
			Help:      "The total number of response marshalling errors",
		},
	)

	UpstreamErrors = promauto.NewCounter(
		p.CounterOpts{
			Namespace: "sinkhole",
			Name:      "upstream_errors_total",
			Help:      "The total number of errors encountered in the upstream DNS server",
		},
	)

	WriteResponseErrors = promauto.NewCounter(
		p.CounterOpts{
			Namespace: "sinkhole",
			Name:      "write_response_errors_total",
			Help:      "The total number of errors encountered when writing a response to the client",
		},
	)
)
