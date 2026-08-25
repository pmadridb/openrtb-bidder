import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    bid_load: {
      executor: 'ramping-arrival-rate',
      startRate: 500,
      timeUnit: '1s',
      preAllocatedVUs: 100,
      maxVUs: 400,
      stages: [
        { duration: '10s', target: 2000 },
        { duration: '30s', target: 5000 },
        { duration: '10s', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(99)<25', 'p(95)<10'],
    http_req_failed: ['rate<0.001'],
  },
};

const URL = 'http://bidder.yourdomain.com/bid';
const HEADERS = { 'Content-Type': 'application/json' };

const validPayload = JSON.stringify({
  id: 'k6-req-valid',
  imp: [{ id: '1', banner: { w: 300, h: 250 }, bidfloor: 1.0 }],
  site: { id: 'pub-101', domain: 'news-example.com' },
});

export default function () {
  const res = http.post(URL, validPayload, { headers: HEADERS });
  check(res, {
    'status is 200': (r) => r.status === 200,
    'has valid bid response': (r) => r.json('seatbid.0.bid.0.price') > 0,
  });
}