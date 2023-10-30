package metrics

import (
	p "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	NonRoutableDomains = promauto.NewGauge(
		p.GaugeOpts{
			Namespace: "sinkhole",
			Name:      "non_routable_domains",
			Help:      "The total number of domains that we don't want to resolve",
		})

	queries = promauto.NewCounterVec(
		p.CounterOpts{
			Namespace: "sinkhole",
			Name:      "queries",
			Help:      "The total number of queries",
		},
		[]string{"blocked"})

	BlockedQueries  = queries.With(p.Labels{"blocked": "true"})
	FallbackQueries = queries.With(p.Labels{"blocked": "false"})

	discardedQueries = promauto.NewCounterVec(
		p.CounterOpts{
			Namespace: "sinkhole",
			Name:      "discarded_queries",
			Help:      "The total number of discarded queries",
		},
		[]string{"reason"})

	NonStandardQueries      = discardedQueries.With(p.Labels{"reason": "non-standard"})
	NonRecursiveQueries     = discardedQueries.With(p.Labels{"reason": "non-recursive"})
	UnsupportedClassQueries = discardedQueries.With(p.Labels{"reason": "unsupported-class"})
	UnsupportedTypeQueries  = discardedQueries.With(p.Labels{"reason": "unsupported-type"})

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
