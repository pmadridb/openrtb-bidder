# OpenRTB 2.5 Real-Time Bidding (DSP) Engine in Go

A high-throughput, low-latency Demand-Side Platform (DSP) decision engine built in Go. Designed to process OpenRTB 2.5 bid requests within strict millisecond SLAs, execute multi-parameter campaign matching, handle HTTP 204 No-Content signals, and export operational metrics for Prometheus monitoring.

---

## Architecture & Features

* **High-Performance Bid Decisioning:** Thread-safe, low-allocation campaign evaluation pipeline executing in `< 1ms`.
* **OpenRTB 2.5 Specification Compliance:** Supports standard objects including `BidRequest`, `Imp`, `Banner`, `Geo`, and `Device`.
* **Zero-Payload No-Bids (`204 No Content`):** Returns immediate `HTTP 204` responses on campaign mismatches or floor price failures to minimize outbound bandwidth across high QPS pipelines.
* **Dynamic Macro Resolution:** Injects standard OpenRTB macros (`${AUCTION_PRICE}`) into win notice URLs (`nurl`) for post-auction settlement.
* **Low-Overhead Metrics:** Uses atomic counters (`sync/atomic`) to expose lock-free Prometheus telemetry at `/metrics`.
* **Production Containerization & Orchestration:** Built-in multi-stage `Dockerfile` and Kubernetes manifests for seamless cluster deployments with readiness and liveness checks.

---

## System Architecture Flow

```
[ Exchange / SSP ] 
       │
       │  1. POST /bid (OpenRTB 2.5 JSON)
       ▼
┌────────────────────────────────────────────────────────┐
│ OpenRTB Bidder Service                                 │
│                                                        │
│  ├── Timeout Guard (50ms Read/Write Limits)            │
│  ├── Engine Evaluation Pipeline                        │
│  │     ├── Floor Price Filtering                        │
│  │     ├── Format & Dimension Matching (e.g., 300x250) │
│  │     └── Geo & Device Targeting                       │
│  │                                                     │
│  ├── [Match Found]    ──► HTTP 200 OK (Bid Object)     │
│  └── [No Match/Floor] ──► HTTP 204 No Content          │
└────────────────────────────────────────────────────────┘
       │
       │  2. Win Notification Callback
       ▼
  GET /win?impid={ID}&price=${AUCTION_PRICE}

```

---

## API Endpoints

| Endpoint | Method | Description | Response Codes |
| --- | --- | --- | --- |
| `/bid` | `POST` | Core OpenRTB auction decisioning engine | `200 OK` (Bid), `204 No Content` (No-Bid), `400 Bad Request` |
| `/win` | `GET` | Billing and auction clearance price logger | `200 OK` |
| `/metrics` | `GET` | Real-time Prometheus metrics endpoint | `200 OK` |
| `/healthz` | `GET` | Kubernetes liveness and readiness probe | `200 OK` |

---

## Campaign Targeting Logic & Floor Rules

The bidder evaluates incoming impressions against active inventory campaigns using a strict cascade:

1. **Floor Price Compliance:** Bids are dropped if `Campaign.ECPM < Imp.BidFloor`.
2. **Ad Format & Size Matching:** Requires explicit width (`w`) and height (`h`) parity for banner creatives.
3. **Geographic Filtering:** Filters bids based on ISO country code matches in `Device.Geo.Country`.
4. **Device Type Validation:** Matches allowed device classifications (e.g., Mobile/Tablet vs. Desktop).

If no campaign satisfies all constraints, the bidder suppresses response generation and instantly emits `HTTP 204`.

---

## Quickstart & Local Development

### Prerequisites

* Go 1.22+ installed locally

### Running Locally

```bash
# Clone the repository
git clone https://github.com/your-org/openrtb-bidder.git
cd openrtb-bidder

# Run the bidder server
go run main.go

```

The server listens on port `8080` by default.

---

## Testing & Verification

### 1. Send a Valid Matching Bid Request (`200 OK`)

```bash
curl -i -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d '{
    "id": "req-12345",
    "imp": [{
      "id": "imp-1",
      "banner": { "w": 300, "h": 250 },
      "bidfloor": 1.50
    }],
    "device": {
      "devicetype": 4,
      "geo": { "country": "USA" }
    }
  }'

```

**Expected Response (`200 OK`):**

```json
{
  "id": "req-12345",
  "seatbid": [
    {
      "bid": [
        {
          "id": "bid-imp-1",
          "impid": "imp-1",
          "price": 2.50,
          "adid": "camp-300x250-usa",
          "nurl": "http://localhost:8080/win?bidid=bid-imp-1&impid=imp-1&price=${AUCTION_PRICE}",
          "adm": "<script src=\"https://cdn.adserver.com/tag.js\"></script>",
          "w": 300,
          "h": 250
        }
      ],
      "seat": "dsp-seat-1"
    }
  ],
  "bidid": "resp-req-12345",
  "cur": "USD"
}

```

### 2. Send a Non-Matching Request (`204 No Content`)

```bash
curl -i -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d '{
    "id": "req-67890",
    "imp": [{
      "id": "imp-2",
      "banner": { "w": 728, "h": 90 },
      "bidfloor": 50.00
    }]
  }'

```

**Expected Response:**

```http
HTTP/1.1 204 No Content
Date: Sun, 23 Aug 2026 14:18:51 GMT

```

### 3. Inspect Prometheus Metrics

```bash
curl http://localhost:8080/metrics

```

**Output:**

```text
# HELP rtb_requests_total Total incoming bid requests
# TYPE rtb_requests_total counter
rtb_requests_total 2

# HELP rtb_responses_total Total bid responses by status
# TYPE rtb_responses_total counter
rtb_responses_total{status="bid"} 1
rtb_responses_total{status="nobid"} 1
rtb_responses_total{status="error"} 0

# HELP rtb_processing_latency_avg_ms Average request processing latency in milliseconds
# TYPE rtb_processing_latency_avg_ms gauge
rtb_processing_latency_avg_ms 0.420

```

---

## Deployment Options

### Docker Deployment

```bash
# Build lightweight container image
docker build -t openrtb-bidder:latest .

# Run container exposing port 8080
docker run -d -p 8080:8080 --name bidder openrtb-bidder:latest

```

### Kubernetes Deployment

Apply the manifests located in the `k8s/` directory to deploy the service with automated health probing and scaling:

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

```

---

## Load Testing with k6

Execute a performance test with 100 concurrent virtual users to verify latency bounds under load:

```bash
k6 run tests/load-test.js

```

---