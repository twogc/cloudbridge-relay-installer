import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Quick test configuration
export const options = {
  stages: [
    { duration: '10s', target: 5 },  // Ramp up to 5 users
    { duration: '20s', target: 5 },  // Stay at 5 users
    { duration: '10s', target: 10 }, // Ramp up to 10 users
    { duration: '20s', target: 10 }, // Stay at 10 users
    { duration: '10s', target: 0 },  // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'], // 95% of requests must complete below 1000ms
    http_req_failed: ['rate<0.2'],     // Error rate must be below 20%
    errors: ['rate<0.2'],              // Custom error rate must be below 20%
  },
};

// Test data
const BASE_URL = 'http://localhost:9090';

export default function () {
  // Test 1: Health check
  const healthResponse = http.get(`${BASE_URL}/health`);
  const healthCheck = check(healthResponse, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 500ms': (r) => r.timings.duration < 500,
    'health check contains service name': (r) => r.body.includes('cloudbridge-relay'),
    'health check contains status ok': (r) => r.body.includes('"status":"ok"'),
  });
  
  errorRate.add(!healthCheck);

  // Test 2: Root endpoint (expected to return 404)
  const rootResponse = http.get(`${BASE_URL}/`);
  const rootCheck = check(rootResponse, {
    'root endpoint returns 404': (r) => r.status === 404,
  });

  // Random sleep between requests
  sleep(Math.random() * 1 + 0.5); // 0.5-1.5 seconds
}

export function handleSummary(data) {
  return {
    'quick-load-test-results.json': JSON.stringify(data, null, 2),
    'quick-load-test-summary.txt': `
CloudBridge Relay Quick Load Test Results
=========================================

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
- HTTP request duration (95th percentile): ${data.metrics.http_req_duration.values['p(95)'] < 1000 ? 'PASS' : 'FAIL'} (< 1000ms)
- HTTP request failure rate: ${data.metrics.http_req_failed.values.rate < 0.2 ? 'PASS' : 'FAIL'} (< 20%)
- Custom error rate: ${data.metrics.errors ? (data.metrics.errors.values.rate < 0.2 ? 'PASS' : 'FAIL') : 'N/A'} (< 20%)

Status: ${data.metrics.http_req_failed.values.rate < 0.2 && data.metrics.http_req_duration.values['p(95)'] < 1000 ? 'PASS' : 'FAIL'}
    `,
  };
}
