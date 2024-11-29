package routes

import (
	"net/http"

	"github.com/pandorasNox/lettr/pkg/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metrics struct {
	honeyTrapped prometheus.Gauge
}

func GetMetrics(server *server.Server) http.HandlerFunc {
	// promhandle := promhttp.Handler()

	reg := prometheus.NewRegistry()
	m := NewMetrics()
	reg.MustRegister(m.honeyTrapped)

	// add some defaults from prometheus package
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewBuildInfoCollector())

	return func(w http.ResponseWriter, r *http.Request) {
		m.honeyTrapped.Set(float64(server.Metrics().HoneyTrapped()))

		promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

		promHandler.ServeHTTP(w, r)
	}
}

func NewMetrics() *metrics {
	m := &metrics{
		honeyTrapped: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "lettr",
			Name:      "honey_trapped",
			Help:      "Number of request send via our suggest form (message field) honey trap.",
		}),
	}
	return m
}
