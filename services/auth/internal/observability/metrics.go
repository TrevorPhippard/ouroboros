package observability

import "auth/internal/config"

func NewMetrics(cfg config.Config) *Metrics {
	return &Metrics{
		// Initialize your metrics here, e.g., using Prometheus client
	}
}

type Metrics struct {
	// Define your metrics fields here, e.g., counters, histograms, etc.
}
