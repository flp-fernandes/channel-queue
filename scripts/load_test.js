import http from 'k6/http';
import { check } from 'k6';
import { Counter } from 'k6/metrics';

const errors = new Counter('errors');

export const options = {
  vus: 50,
  iterations: 1000000,
};

const payload = JSON.stringify({ product_id: 1 });
const headers = { 'Content-Type': 'application/json' };

export default function () {
  const res = http.post('http://localhost:3000/events/product/view', payload, { headers });

  const ok = check(res, {
    '202 Accepted': (r) => r.status === 202,
  });

  if (!ok) errors.add(1);
}
