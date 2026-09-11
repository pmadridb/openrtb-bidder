package engine

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
	"sync/atomic"

	"openrtb-bidder/internal/model"
)

// TargetSize represents supported banner dimensions.
type TargetSize struct {
	W int
	H int
}

// Campaign holds targeting rules, pricing, and creative assets for an advertiser.
type Campaign struct {
	ID            string
	AdvertiserID  string
	CreativeID    string
	MaxCPM        float64
	TargetDomains []string
	TargetSizes   []TargetSize
	HTMLTemplate  string
}

// DecisionEngine evaluates BidRequests against targeted campaigns.
type DecisionEngine struct {
	campaigns  atomic.Pointer[[]Campaign]
	winBaseURL string
}

func NewDecisionEngine(winBaseURL string) *DecisionEngine {
	engine := &DecisionEngine{
		winBaseURL: winBaseURL,
	}
	engine.loadSeedCampaigns()
	return engine
}

// Evaluate loops through all impressions in a request and scores potential bids.
func (e *DecisionEngine) Evaluate(req *model.BidRequest) *model.BidResponse {
	if req == nil || len(req.Imp) == 0 {
		return nil
	}

	var bids []model.Bid

	for _, imp := range req.Imp {
		// Only support banner inventory in this step
		if imp.Banner == nil {
			continue
		}

		matchedCampaign := e.findMatchingCampaign(req, imp)
		if matchedCampaign == nil {
			continue
		}

		// Calculate dynamic price based on auction floor
		bidPrice := e.calculateBidPrice(matchedCampaign, imp.BidFloor)
		if bidPrice < imp.BidFloor {
			continue // Floor too high
		}

		bid := e.buildBid(matchedCampaign, imp, bidPrice)
		bids = append(bids, bid)
	}

	if len(bids) == 0 {
		return nil // Signal HTTP 204 No Content
	}

	return &model.BidResponse{
		ID:    req.ID,
		BidID: generateUUID(),
		Cur:   "USD",
		SeatBid: []model.SeatBid{
			{
				Bid:  bids,
				Seat: "dsp-seat-1",
			},
		},
	}
}

func (e *DecisionEngine) findMatchingCampaign(req *model.BidRequest, imp model.Impression) *Campaign {
	for _, campaign := range *e.campaigns.Load() {
		if !e.matchesDomain(campaign, req.Site) {
			continue
		}
		if !e.matchesSize(campaign, imp.Banner) {
			continue
		}
		return &campaign
	}
	return nil
}

func (e *DecisionEngine) matchesDomain(c Campaign, site *model.Site) bool {
	if len(c.TargetDomains) == 0 {
		return true // Broad targeting
	}
	if site == nil || site.Domain == "" {
		return false
	}
	for _, domain := range c.TargetDomains {
		if strings.Contains(strings.ToLower(site.Domain), strings.ToLower(domain)) {
			return true
		}
	}
	return false
}

func (e *DecisionEngine) matchesSize(c Campaign, banner *model.Banner) bool {
	for _, size := range c.TargetSizes {
		if size.W == banner.W && size.H == banner.H {
			return true
		}
	}
	return false
}

func (e *DecisionEngine) calculateBidPrice(c *Campaign, bidFloor float64) float64 {
	targetPrice := c.MaxCPM * 0.85 // Target 15% margin below max budget

	if bidFloor > 0 {
		if targetPrice < bidFloor {
			// Bid slightly above the floor if within budget
			if c.MaxCPM >= bidFloor+0.05 {
				return math.Round((bidFloor+0.05)*100) / 100
			}
			return 0 // Cannot meet floor
		}
	}

	return math.Round(targetPrice*100) / 100
}

func (e *DecisionEngine) buildBid(c *Campaign, imp model.Impression, price float64) model.Bid {
	bidID := generateUUID()
	winURL := fmt.Sprintf("%s/win?price=${AUCTION_PRICE}&impid=%s&bidid=%s", e.winBaseURL, imp.ID, bidID)
	admMarkup := fmt.Sprintf(c.HTMLTemplate, imp.Banner.W, imp.Banner.H)

	return model.Bid{
		ID:    bidID,
		ImpID: imp.ID,
		Price: price,
		AdID:  c.CreativeID,
		NURL:  winURL,
		ADM:   admMarkup,
		CID:   c.ID,
		CRID:  c.CreativeID,
		W:     imp.Banner.W,
		H:     imp.Banner.H,
	}
}

func (e *DecisionEngine) loadSeedCampaigns() {
	campaigns := []Campaign{
		{
			ID:            "cmp-tech-01",
			AdvertiserID:  "adv-cloud-corp",
			CreativeID:    "cr-300x250-cloud",
			MaxCPM:        3.50,
			TargetDomains: []string{"news-example.com", "techblog.io"},
			TargetSizes:   []TargetSize{{W: 300, H: 250}, {W: 728, H: 90}},
			HTMLTemplate:  `<a href="https://cloud.example.com"><img src="https://cdn.example.com/300x250.png" width="%d" height="%d"/></a>`,
		},
		{
			ID:            "cmp-auto-02",
			AdvertiserID:  "adv-motors",
			CreativeID:    "cr-728x90-auto",
			MaxCPM:        2.10,
			TargetDomains: []string{}, // Targets all domains
			TargetSizes:   []TargetSize{{W: 728, H: 90}},
			HTMLTemplate:  `<a href="https://motors.example.com"><img src="https://cdn.example.com/728x90.png" width="%d" height="%d"/></a>`,
		},
	}
	e.campaigns.Store(&campaigns)
}

func generateUUID() string {
	return fmt.Sprintf("%08x-%08x", rand.Uint32(), rand.Uint32())
}
