package metrics

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// 📊 Общее количество выполнений usecase'ов
	UsecaseTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bot_usecase_total",
			Help: "Total number of usecases executed",
		},
		[]string{"usecase"},
	)

	// ⏱️ Время выполнения usecase'ов
	UsecaseDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bot_usecase_duration_seconds",
			Help:    "Duration of usecase execution in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"usecase"},
	)

	// 🧍 Сила игрока (обновляется после анализа состояния)
	GamerPowerGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bot_gamer_power",
			Help: "Power of the current gamer",
		},
		[]string{"gamer"},
	)

	// 🔥 Уровень печки игрока
	GamerFurnaceLevel = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bot_gamer_furnace_level",
			Help: "Furnace building level of the gamer",
		},
		[]string{"gamer"},
	)

	// ❌ Ошибки взаимодействия с ADB
	ADBErrorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bot_adb_error_total",
			Help: "Number of ADB-related errors (click/screenshot/swipe)",
		},
		[]string{"device_id", "type"},
	)
)

// 🚀 Регистрируем все метрики при запуске
func Init() {
	prometheus.MustRegister(
		UsecaseTotal,
		UsecaseDuration,
		GamerPowerGauge,
		GamerFurnaceLevel,
		ADBErrorTotal,
	)
}

// 🌐 Запуск HTTP-сервера для экспорта метрик Prometheus
func StartExporter() {
	Init()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("📈 Prometheus metrics available at http://localhost:2112/metrics")
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Fatalf("❌ Failed to start metrics exporter: %v", err)
		}
	}()
}
