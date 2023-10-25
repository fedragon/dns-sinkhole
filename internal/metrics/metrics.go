package metrics

import (
	p "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	BlacklistedDomains = promauto.NewGauge(
		p.GaugeOpts{
			Namespace: "sinkhole",
			Name:      "blacklisted_domains",
			Help:      "The total number of blacklisted domains",
		})

	queries = promauto.NewCounterVec(
		p.CounterOpts{
			Namespace: "sinkhole",
			Name:      "blacklisted_queries",
			Help:      "The total number of blacklisted queries",
		},
		[]string{"blocked"})

	BlockedQueries = queries.With(p.Labels{"blocked": "true"})
	LegitQueries   = queries.With(p.Labels{"blocked": "false"})

	QueryParsingErrors = promauto.NewCounter(
		p.CounterOpts{
			Namespace: "sinkhole",
			Name:      "query_parsing_errors",
			Help:      "The total number of query parsing errors",
		},
	)

	ResponseMarshallingErrors = promauto.NewCounter(
		p.CounterOpts{
			Namespace: "sinkhole",
			Name:      "response_marshalling_errors",
			Help:      "The total number of response marshalling errors",
		},
	)

	FallbackErrors = promauto.NewCounter(
		p.CounterOpts{
			Namespace: "sinkhole",
			Name:      "fallback_errors",
			Help:      "The total number of fallback errors",
		},
	)
)
