package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// Metrics holds atomic counters for real-time bidder monitoring.
type Metrics struct {
	TotalRequests   uint64
	BidResponses    uint64
	NoBidResponses  uint64
	ErrorRequests   uint64
	TotalDurationUs uint64 // Total processing duration in microseconds
}

var DefaultMetrics = &Metrics{}

func (m *Metrics) RecordRequest(duration time.Duration, hasBid bool, isError bool) {
	atomic.AddUint64(&m.TotalRequests, 1)
	atomic.AddUint64(&m.TotalDurationUs, uint64(duration.Microseconds()))

	if isError {
		atomic.AddUint64(&m.ErrorRequests, 1)
		return
	}

	if hasBid {
		atomic.AddUint64(&m.BidResponses, 1)
	} else {
		atomic.AddUint64(&m.NoBidResponses, 1)
	}
}

// Handler returns metrics formatted in standard Prometheus text exposition format.
func (m *Metrics) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqs := atomic.LoadUint64(&m.TotalRequests)
		bids := atomic.LoadUint64(&m.BidResponses)
		nobids := atomic.LoadUint64(&m.NoBidResponses)
		errs := atomic.LoadUint64(&m.ErrorRequests)
		durationUs := atomic.LoadUint64(&m.TotalDurationUs)

		avgLatencyMs := 0.0
		if reqs > 0 {
			avgLatencyMs = float64(durationUs) / float64(reqs) / 1000.0
		}

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP rtb_requests_total Total incoming bid requests\n")
		fmt.Fprintf(w, "# TYPE rtb_requests_total counter\n")
		fmt.Fprintf(w, "rtb_requests_total %d\n\n", reqs)

		fmt.Fprintf(w, "# HELP rtb_responses_total Total bid responses by status\n")
		fmt.Fprintf(w, "# TYPE rtb_responses_total counter\n")
		fmt.Fprintf(w, "rtb_responses_total{status=\"bid\"} %d\n", bids)
		fmt.Fprintf(w, "rtb_responses_total{status=\"nobid\"} %d\n", nobids)
		fmt.Fprintf(w, "rtb_responses_total{status=\"error\"} %d\n\n", errs)

		fmt.Fprintf(w, "# HELP rtb_processing_latency_avg_ms Average request processing latency in milliseconds\n")
		fmt.Fprintf(w, "# TYPE rtb_processing_latency_avg_ms gauge\n")
		fmt.Fprintf(w, "rtb_processing_latency_avg_ms %.3f\n", avgLatencyMs)
	}
}