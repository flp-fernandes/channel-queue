import http from 'k6/http';
import { check } from 'k6';
import { Counter } from 'k6/metrics';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const errors = new Counter('errors');

export const options = {
  vus: 50,
  iterations: 1000000,
};

const headers = { 'Content-Type': 'application/json' };

export default function () {
  const payload = JSON.stringify({ product_id: uuidv4() });
  const res = http.post('http://localhost:3000/events/product/view', payload, { headers });

  const ok = check(res, {
    '202 Accepted': (r) => r.status === 202,
  });

  if (!ok) errors.add(1);
}
