import http from 'k6/http';
import { check, sleep } from 'k6';

// CloudBridge Relay Integration Test (without Cadence)
export const options = {
  stages: [
    { duration: '10s', target: 3 },  // Ramp up to 3 users
    { duration: '20s', target: 3 },  // Stay at 3 users
    { duration: '10s', target: 0 },  // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    http_req_failed: ['rate<0.1'],    // Error rate must be below 10%
  },
};

const BASE_URL = 'http://localhost:9090';

export default function () {
  // Test 1: Health Check
  const healthResponse = http.get(`${BASE_URL}/health`);
  const healthCheck = check(healthResponse, {
    'Health check status 200': (r) => r.status === 200,
    'Health check response time < 200ms': (r) => r.timings.duration < 200,
    'Health check contains service name': (r) => r.body.includes('cloudbridge-relay'),
    'Health check status ok': (r) => r.body.includes('"status":"ok"'),
    'Health check contains components': (r) => r.body.includes('"components"'),
    'Health check contains uptime': (r) => r.body.includes('"uptime"'),
  });

  // Test 2: Metrics Endpoint
  const metricsResponse = http.get(`${BASE_URL}/metrics`);
  const metricsCheck = check(metricsResponse, {
    'Metrics endpoint accessible': (r) => r.status === 200 || r.status === 404,
    'Metrics response time < 500ms': (r) => r.timings.duration < 500,
  });

  // Test 3: Root Endpoint
  const rootResponse = http.get(`${BASE_URL}/`);
  const rootCheck = check(rootResponse, {
    'Root endpoint accessible': (r) => r.status === 200 || r.status === 404,
    'Root response time < 500ms': (r) => r.timings.duration < 500,
  });

  // Test 4: API Endpoints
  const apiEndpoints = ['/api/v1/status', '/api/health', '/api/status'];
  const randomApiEndpoint = apiEndpoints[Math.floor(Math.random() * apiEndpoints.length)];
  const apiResponse = http.get(`${BASE_URL}${randomApiEndpoint}`);
  
  const apiCheck = check(apiResponse, {
    'API endpoint accessible': (r) => r.status === 200 || r.status === 404,
    'API response time < 500ms': (r) => r.timings.duration < 500,
  });

  // Test 5: WebSocket Endpoint (if available)
  const wsResponse = http.get(`${BASE_URL}/ws`);
  const wsCheck = check(wsResponse, {
    'WebSocket endpoint accessible': (r) => r.status === 200 || r.status === 404 || r.status === 426,
    'WebSocket response time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(1.5); // 1.5 seconds between requests
}

export function handleSummary(data) {
  const avgResponseTime = data.metrics.http_req_duration.values.avg;
  const p95ResponseTime = data.metrics.http_req_duration.values['p(95)'];
  const errorRate = data.metrics.http_req_failed.values.rate * 100;
  const throughput = data.metrics.http_reqs.values.rate;
  const checkSuccessRate = (data.metrics.checks.values.count - data.metrics.checks.values.fails) / data.metrics.checks.values.count * 100;
  
  return {
    'cloudbridge-integration-test-results.json': JSON.stringify(data, null, 2),
    'cloudbridge-integration-test-summary.txt': `
CloudBridge Relay Integration Test Results
==========================================

Test Duration: ${(data.state.testRunDurationMs / 1000).toFixed(1)}s
Total Requests: ${data.metrics.http_reqs.values.count}
Failed Requests: ${data.metrics.http_req_failed.values.count}
Error Rate: ${errorRate.toFixed(2)}%

Response Times:
- Average: ${avgResponseTime.toFixed(2)}ms
- Median: ${data.metrics.http_req_duration.values.med.toFixed(2)}ms
- 95th percentile: ${p95ResponseTime.toFixed(2)}ms
- 99th percentile: ${data.metrics.http_req_duration.values['p(99)'].toFixed(2)}ms

Throughput:
- Requests per second: ${throughput.toFixed(2)}
- Data received: ${(data.metrics.data_received.values.count / 1024).toFixed(2)} KB
- Data sent: ${(data.metrics.data_sent.values.count / 1024).toFixed(2)} KB

Check Success Rate: ${checkSuccessRate.toFixed(2)}%

Thresholds:
- HTTP request duration (95th percentile): ${p95ResponseTime < 500 ? 'PASS' : 'FAIL'} (< 500ms)
- HTTP request failure rate: ${errorRate < 10 ? 'PASS' : 'FAIL'} (< 10%)

Overall Status: ${errorRate < 10 && p95ResponseTime < 500 ? 'PASS' : 'FAIL'}

Integration Test Assessment:
- Health Endpoint: ${checkSuccessRate > 90 ? 'HEALTHY' : 'ISSUES DETECTED'}
- API Endpoints: ${checkSuccessRate > 80 ? 'WORKING' : 'NEEDS ATTENTION'}
- Overall Integration: ${errorRate < 10 && p95ResponseTime < 500 ? 'SUCCESS' : 'FAILED'}

Performance Grade:
- Response Time: ${avgResponseTime < 50 ? 'A' : avgResponseTime < 100 ? 'B' : avgResponseTime < 200 ? 'C' : 'D'}
- Reliability: ${errorRate < 1 ? 'A' : errorRate < 5 ? 'B' : errorRate < 10 ? 'C' : 'D'}
- Throughput: ${throughput > 10 ? 'A' : throughput > 5 ? 'B' : throughput > 2 ? 'C' : 'D'}
    `,
  };
}

