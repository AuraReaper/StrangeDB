package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// request metrics
	RequestTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "strangedb_requests_total",
		Help: "Total number of requests",
	},
		[]string{"operation", "ststus"},
	)

	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "stragedb_request_duration_seconds",
		Help:    "Request duration in seconds",
		Buckets: prometheus.DefBuckets,
	},
		[]string{"operation"},
	)

	//storage metrics
	KeysTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "strangedb_keys_total",
		Help: "Total number of keys strored",
	})

	StorageBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "strangedb_storage_bytes",
		Help:      "Storage size in bytes",
	})

	// cluster metrics
	NodesTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "strangedb_nodes_total",
		Help: "Total number of nodes i cluster",
	})

	GossipMessagesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "strangedb_gossip_messages_total",
		Help: "Total gossip messages",
	},
		[]string{"type"},
	)

	// replication metrics
	ReplicationLagSeconds = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "strangedb_replication_lag_seconds",
		Help: "Replication lag in seconds",
	},
		[]string{"node"},
	)

	ReadRepairsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "strangedb_read_repairs_total",
		Help: "Total read repairs performed",
	})
)

func RecordRequest(operation, status string, duration float64) {
	RequestTotal.WithLabelValues(operation, status).Inc()
	RequestDuration.WithLabelValues(operation).Observe(duration)
}
