import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';

const soldOut   = new Counter('sold_out_responses');
const successOk = new Counter('successful_reservations');
const p99trend  = new Trend('reserve_p99');

export const options = {
  scenarios: {
    flash_rush: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 500 },
        { duration: '30s', target: 1000 },
        { duration: '10s', target: 0 },
      ],
    },
  },
  thresholds: {
    'http_req_duration{p(99)}': ['p(99)<50'],
    'sold_out_responses':       ['count>=0'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const userID = `user_${Math.floor(Math.random() * 100000)}`;
  const payload = JSON.stringify({ user_id: userID, sale_id: 1 });
  const params  = { headers: { 'Content-Type': 'application/json' } };

  const res = http.post(`${BASE_URL}/api/reserve`, payload, params);
  p99trend.add(res.timings.duration);

  check(res, {
    'status is 200 or 409 or 410': (r) => [200, 409, 410].includes(r.status),
  });

  if (res.status === 200)        successOk.add(1);
  else if (res.status === 409)   soldOut.add(1);

  sleep(0.001);
}