package engine

import (
	"testing"

	"openrtb-bidder/internal/model"
)

func TestDecisionEngine_Evaluate_Match(t *testing.T) {
	eng := NewDecisionEngine("http://localhost:8080")

	req := &model.BidRequest{
		ID: "req-1",
		Site: &model.Site{
			ID:     "site-1",
			Domain: "news-example.com",
		},
		Imp: []model.Impression{
			{
				ID: "imp-1",
				Banner: &model.Banner{
					W: 300,
					H: 250,
				},
				BidFloor: 1.0,
			},
		},
	}

	resp := eng.Evaluate(req)
	if resp == nil {
		t.Fatalf("expected non-nil response for matching bid request")
	}
	if len(resp.SeatBid) == 0 || len(resp.SeatBid[0].Bid) == 0 {
		t.Fatalf("expected seatbid with bids")
	}
	bid := resp.SeatBid[0].Bid[0]
	if bid.Price < 1.0 {
		t.Errorf("expected bid price >= floor price (1.0), got %f", bid.Price)
	}
	if bid.ImpID != "imp-1" {
		t.Errorf("expected impid imp-1, got %s", bid.ImpID)
	}
}

func TestDecisionEngine_Evaluate_NoMatch(t *testing.T) {
	eng := NewDecisionEngine("http://localhost:8080")

	req := &model.BidRequest{
		ID: "req-2",
		Site: &model.Site{
			ID:     "site-2",
			Domain: "unmatched-domain.com",
		},
		Imp: []model.Impression{
			{
				ID: "imp-2",
				Banner: &model.Banner{
					W: 500,
					H: 500, // Non-matching size
				},
				BidFloor: 1.0,
			},
		},
	}

	resp := eng.Evaluate(req)
	if resp != nil {
		t.Fatalf("expected nil (HTTP 204) response for non-matching bid request, got %+v", resp)
	}
}

func TestDecisionEngine_Evaluate_FloorTooHigh(t *testing.T) {
	eng := NewDecisionEngine("http://localhost:8080")

	req := &model.BidRequest{
		ID: "req-3",
		Site: &model.Site{
			ID:     "site-3",
			Domain: "news-example.com",
		},
		Imp: []model.Impression{
			{
				ID: "imp-3",
				Banner: &model.Banner{
					W: 300,
					H: 250,
				},
				BidFloor: 50.0, // Exceeds max CPM of 3.50
			},
		},
	}

	resp := eng.Evaluate(req)
	if resp != nil {
		t.Fatalf("expected nil for floor exceeding max budget, got %+v", resp)
	}
}
