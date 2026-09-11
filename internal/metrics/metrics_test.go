package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMetrics_RecordRequest(t *testing.T) {
	m := &Metrics{}

	m.RecordRequest(5*time.Millisecond, true, false)
	m.RecordRequest(3*time.Millisecond, false, false)
	m.RecordRequest(1*time.Millisecond, false, true)

	if m.TotalRequests != 3 {
		t.Errorf("expected 3 total requests, got %d", m.TotalRequests)
	}
	if m.BidResponses != 1 {
		t.Errorf("expected 1 bid response, got %d", m.BidResponses)
	}
	if m.NoBidResponses != 1 {
		t.Errorf("expected 1 no-bid response, got %d", m.NoBidResponses)
	}
	if m.ErrorRequests != 1 {
		t.Errorf("expected 1 error request, got %d", m.ErrorRequests)
	}
}

func TestMetrics_Handler(t *testing.T) {
	m := &Metrics{}
	m.RecordRequest(2*time.Millisecond, true, false)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	m.Handler()(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "rtb_requests_total 1") {
		t.Errorf("expected metrics body to contain 'rtb_requests_total 1', got: %s", body)
	}
	if !strings.Contains(body, `rtb_responses_total{status="bid"} 1`) {
		t.Errorf("expected metrics body to contain status=bid, got: %s", body)
	}
}
