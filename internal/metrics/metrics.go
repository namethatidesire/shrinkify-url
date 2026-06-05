package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RedirectsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shrinkify_redirects_total",
			Help: "Total number of successful redirects.",
		},
		[]string{"code"},
	)

	ShortenRequestsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "shrinkify_shorten_requests_total",
			Help: "Total number of shorten requests.",
		},
	)

	RedirectLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name: "shrinkify_redirect_latency_seconds",
			Help: "Latency of redirect handler in seconds.",
			Buckets: prometheus.DefBuckets,
		},
	)

	SQSMessagesSentTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "shrinkify_sqs_messages_sent_total",
			Help: "Total number of SQS click events sent.",
		},
	)
)