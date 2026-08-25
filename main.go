package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"openrtb-bidder/engine"
	"openrtb-bidder/metrics"
	"openrtb-bidder/model"
)

type Server struct {
	decisionEngine *engine.DecisionEngine
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}

	server := &Server{
		decisionEngine: engine.NewDecisionEngine(baseURL),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /bid", server.handleBid)
	mux.HandleFunc("GET /win", server.handleWin)
	mux.HandleFunc("GET /healthz", server.handleHealth)
	mux.HandleFunc("GET /metrics", metrics.DefaultMetrics.Handler())

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  50 * time.Millisecond,
		WriteTimeout: 50 * time.Millisecond,
		IdleTimeout:  10 * time.Second,
	}

	log.Printf("OpenRTB DSP Bidder listening on port %s...", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func (s *Server) handleBid(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Header.Get("Content-Type") != "application/json" {
		metrics.DefaultMetrics.RecordRequest(time.Since(start), false, true)
		http.Error(w, "Invalid Content-Type, expected application/json", http.StatusBadRequest)
		return
	}

	var bidReq model.BidRequest
	if err := json.NewDecoder(r.Body).Decode(&bidReq); err != nil {
		metrics.DefaultMetrics.RecordRequest(time.Since(start), false, true)
		http.Error(w, "Malformed JSON request", http.StatusBadRequest)
		return
	}

	bidResp := s.decisionEngine.Evaluate(&bidReq)
	duration := time.Since(start)

	if bidResp == nil || len(bidResp.SeatBid) == 0 {
		metrics.DefaultMetrics.RecordRequest(duration, false, false)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	metrics.DefaultMetrics.RecordRequest(duration, true, false)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(bidResp)
}

func (s *Server) handleWin(w http.ResponseWriter, r *http.Request) {
	price := r.URL.Query().Get("price")
	impID := r.URL.Query().Get("impid")
	bidID := r.URL.Query().Get("bidid")

	log.Printf("[WIN NOTIFICATION] BidID: %s | ImpID: %s | Cleared Price: $%s", bidID, impID, price)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
