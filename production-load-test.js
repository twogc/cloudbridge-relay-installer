import http from 'k6/http';
import { check, sleep } from 'k6';

// Production load test for CloudBridge Relay
export const options = {
  stages: [
    { duration: '10s', target: 5 },  // Ramp up to 5 users
    { duration: '30s', target: 10 }, // Stay at 10 users
    { duration: '10s', target: 15 }, // Ramp up to 15 users
    { duration: '30s', target: 15 }, // Stay at 15 users
    { duration: '10s', target: 0 },  // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    http_req_failed: ['rate<0.05'],   // Error rate must be below 5%
  },
};

const BASE_URL = 'http://192.168.49.2:31257';

export default function () {
  // Test 1: Health Check
  const healthResponse = http.get(`${BASE_URL}/health`);
  check(healthResponse, {
    'Production health check status 200': (r) => r.status === 200,
    'Production health check response time < 200ms': (r) => r.timings.duration < 200,
    'Production health check contains service name': (r) => r.body.includes('cloudbridge-relay'),
    'Production health check status ok': (r) => r.body.includes('"status":"ok"'),
    'Production health check contains components': (r) => r.body.includes('"components":'),
    'Production health check contains uptime': (r) => r.body.includes('"uptime":'),
  });

  // Test 2: Metrics Endpoint (on port 8084)
  const metricsResponse = http.get(`http://192.168.49.2:32587/metrics`);
  check(metricsResponse, {
    'Production metrics endpoint accessible': (r) => r.status === 200,
    'Production metrics response time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(1);
}
