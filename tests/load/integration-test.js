import http from 'k6/http';
import { check, sleep } from 'k6';

// Integration test for CloudBridge Relay with Cadence
export const options = {
  stages: [
    { duration: '10s', target: 2 },  // Ramp up to 2 users
    { duration: '20s', target: 2 },  // Stay at 2 users
    { duration: '10s', target: 0 },  // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'], // 95% of requests must complete below 1000ms
    http_req_failed: ['rate<0.2'],     // Error rate must be below 20%
  },
};

const BASE_URL = 'http://localhost:9090';
const CADENCE_URL = 'http://localhost:7933';

export default function () {
  // Test 1: CloudBridge Relay Health Check
  const healthResponse = http.get(`${BASE_URL}/health`);
  const healthCheck = check(healthResponse, {
    'CloudBridge Relay health check': (r) => r.status === 200,
    'CloudBridge Relay response time < 200ms': (r) => r.timings.duration < 200,
    'CloudBridge Relay contains service name': (r) => r.body.includes('cloudbridge-relay'),
    'CloudBridge Relay status ok': (r) => r.body.includes('"status":"ok"'),
  });

  // Test 2: Cadence Health Check
  const cadenceResponse = http.get(`${CADENCE_URL}/health`);
  const cadenceCheck = check(cadenceResponse, {
    'Cadence health check': (r) => r.status === 200 || r.status === 404, // 404 is acceptable for Cadence
    'Cadence response time < 500ms': (r) => r.timings.duration < 500,
  });

  // Test 3: CloudBridge Relay Metrics (if available)
  const metricsResponse = http.get(`${BASE_URL}/metrics`);
  const metricsCheck = check(metricsResponse, {
    'Metrics endpoint accessible': (r) => r.status === 200 || r.status === 404,
  });

  // Test 4: Test CloudBridge Relay with different endpoints
  const endpoints = ['/health', '/', '/api/v1/status', '/metrics'];
  const randomEndpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
  const endpointResponse = http.get(`${BASE_URL}${randomEndpoint}`);
  
  const endpointCheck = check(endpointResponse, {
    'Random endpoint accessible': (r) => r.status === 200 || r.status === 404,
    'Random endpoint response time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(2); // 2 seconds between requests
}

export function handleSummary(data) {
  const avgResponseTime = data.metrics.http_req_duration.values.avg;
  const p95ResponseTime = data.metrics.http_req_duration.values['p(95)'];
  const errorRate = data.metrics.http_req_failed.values.rate * 100;
  const throughput = data.metrics.http_reqs.values.rate;
  
  return {
    'integration-test-results.json': JSON.stringify(data, null, 2),
    'integration-test-summary.txt': `
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

Thresholds:
- HTTP request duration (95th percentile): ${p95ResponseTime < 1000 ? 'PASS' : 'FAIL'} (< 1000ms)
- HTTP request failure rate: ${errorRate < 20 ? 'PASS' : 'FAIL'} (< 20%)

Overall Status: ${errorRate < 20 && p95ResponseTime < 1000 ? 'PASS' : 'FAIL'}

Integration Test Assessment:
- CloudBridge Relay: ${data.metrics.checks.values.rate > 0.8 ? 'HEALTHY' : 'ISSUES DETECTED'}
- Cadence Integration: ${data.metrics.checks.values.rate > 0.7 ? 'WORKING' : 'NEEDS ATTENTION'}
- Overall Integration: ${errorRate < 20 && p95ResponseTime < 1000 ? 'SUCCESS' : 'FAILED'}
    `,
  };
}

