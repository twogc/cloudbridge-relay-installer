import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 10 }, // Ramp up to 10 users
    { duration: '1m', target: 10 },  // Stay at 10 users
    { duration: '30s', target: 20 }, // Ramp up to 20 users
    { duration: '1m', target: 20 },  // Stay at 20 users
    { duration: '30s', target: 50 }, // Ramp up to 50 users
    { duration: '1m', target: 50 },  // Stay at 50 users
    { duration: '30s', target: 0 },  // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    http_req_failed: ['rate<0.1'],    // Error rate must be below 10%
    errors: ['rate<0.1'],             // Custom error rate must be below 10%
  },
};

// Test data
const BASE_URL = 'http://localhost:9090';

export default function () {
  // Test 1: Health check
  const healthResponse = http.get(`${BASE_URL}/health`);
  const healthCheck = check(healthResponse, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 200ms': (r) => r.timings.duration < 200,
    'health check contains service name': (r) => r.body.includes('cloudbridge-relay'),
    'health check contains status ok': (r) => r.body.includes('"status":"ok"'),
  });
  
  errorRate.add(!healthCheck);

  // Test 2: Metrics endpoint (if available)
  const metricsResponse = http.get(`${BASE_URL}/metrics`);
  const metricsCheck = check(metricsResponse, {
    'metrics endpoint accessible': (r) => r.status === 200 || r.status === 404,
  });

  // Test 3: Root endpoint
  const rootResponse = http.get(`${BASE_URL}/`);
  const rootCheck = check(rootResponse, {
    'root endpoint accessible': (r) => r.status === 200 || r.status === 404,
  });

  // Test 4: API endpoints (if available)
  const apiResponse = http.get(`${BASE_URL}/api/v1/status`);
  const apiCheck = check(apiResponse, {
    'api endpoint accessible': (r) => r.status === 200 || r.status === 404,
  });

  // Test 5: WebSocket connection test (if available)
  const wsResponse = http.get(`${BASE_URL}/ws`);
  const wsCheck = check(wsResponse, {
    'websocket endpoint accessible': (r) => r.status === 200 || r.status === 404 || r.status === 426,
  });

  // Random sleep between requests
  sleep(Math.random() * 2 + 0.5); // 0.5-2.5 seconds
}

export function handleSummary(data) {
  return {
    'load-test-results.json': JSON.stringify(data, null, 2),
    'load-test-summary.txt': `
CloudBridge Relay Load Test Results
===================================

Test Duration: ${data.state.testRunDurationMs / 1000}s
Total Requests: ${data.metrics.http_reqs.values.count}
Failed Requests: ${data.metrics.http_req_failed.values.count}
Error Rate: ${(data.metrics.http_req_failed.values.rate * 100).toFixed(2)}%

Response Times:
- Average: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms
- Median: ${data.metrics.http_req_duration.values.med.toFixed(2)}ms
- 95th percentile: ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms
- 99th percentile: ${data.metrics.http_req_duration.values['p(99)'].toFixed(2)}ms

Throughput:
- Requests per second: ${data.metrics.http_reqs.values.rate.toFixed(2)}
- Data received: ${(data.metrics.data_received.values.count / 1024).toFixed(2)} KB
- Data sent: ${(data.metrics.data_sent.values.count / 1024).toFixed(2)} KB

Thresholds:
- HTTP request duration (95th percentile): ${data.metrics.http_req_duration.values['p(95)'] < 500 ? 'PASS' : 'FAIL'} (< 500ms)
- HTTP request failure rate: ${data.metrics.http_req_failed.values.rate < 0.1 ? 'PASS' : 'FAIL'} (< 10%)
- Custom error rate: ${data.metrics.errors ? (data.metrics.errors.values.rate < 0.1 ? 'PASS' : 'FAIL') : 'N/A'} (< 10%)

Status: ${data.metrics.http_req_failed.values.rate < 0.1 && data.metrics.http_req_duration.values['p(95)'] < 500 ? 'PASS' : 'FAIL'}
    `,
  };
}

