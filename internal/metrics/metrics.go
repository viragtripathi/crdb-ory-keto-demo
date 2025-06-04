package metrics

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	TupleInsertDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "tuple_insert_duration_seconds",
		Help:    "Duration of tuple inserts.",
		Buckets: prometheus.DefBuckets,
	})

	PermissionCheckCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "permission_check_total",
			Help: "Total permission checks run",
		},
		[]string{"result"},
	)
)

func Init() {
	prometheus.MustRegister(TupleInsertDuration)
	prometheus.MustRegister(PermissionCheckCounter)

	// Health and metrics endpoints
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Println("📡 Starting metrics HTTP server on :2112")
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Fatalf("❌ Metrics server failed: %v", err)
		}
	}()
}
