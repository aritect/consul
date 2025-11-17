package metrics

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	TelegramMessagesReceived = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "consul_telegram_bot_telegram_messages_received_total",
			Help: "Total number of Telegram messages received",
		},
		[]string{"chat_type", "command"},
	)

	TelegramMessagesSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "consul_telegram_bot_telegram_messages_sent_total",
			Help: "Total number of Telegram messages sent",
		},
		[]string{"chat_type", "status"},
	)

	TelegramCommandsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "consul_telegram_bot_telegram_commands_processed_total",
			Help: "Total number of Telegram commands processed",
		},
		[]string{"command", "status"},
	)

	TelegramActiveUsers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "consul_telegram_bot_telegram_active_users",
			Help: "Number of active Telegram users",
		},
	)

	LevelDBOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "consul_telegram_bot_leveldb_operations_total",
			Help: "Total number of LevelDB operations",
		},
		[]string{"operation", "status"},
	)

	LevelDBSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "consul_telegram_bot_leveldb_size_bytes",
			Help: "Current LevelDB database size in bytes",
		},
	)

	ProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "consul_telegram_bot_processing_duration_microseconds",
			Help: "Time spent processing operations in microseconds",
			Buckets: []float64{
				1,
				5,
				10,
				50,
				100,
				500,
				1000,
				5000,
				10000,
				50000,
				100000,
				500000,
				1000000,
			},
		},
		[]string{"operation"},
	)

	ErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "consul_telegram_bot_errors_total",
			Help: "Total number of errors",
		},
		[]string{"component", "error_type"},
	)

	UptimeSeconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "consul_telegram_bot_uptime_seconds",
			Help: "Service uptime in seconds",
		},
	)
)

func init() {
	prometheus.MustRegister(
		TelegramMessagesReceived,
		TelegramMessagesSent,
		TelegramCommandsProcessed,
		TelegramActiveUsers,
		LevelDBOperations,
		LevelDBSize,
		ProcessingDuration,
		ErrorsTotal,
		UptimeSeconds,
	)
}

func StartMetricsServer(port string) {
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("starting metrics server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("failed to start metrics server: %v", err)
	}
}
