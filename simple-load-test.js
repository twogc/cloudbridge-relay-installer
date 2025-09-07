import http from 'k6/http';
import { check, sleep } from 'k6';

// Simple load test for health endpoint
export const options = {
  stages: [
    { duration: '10s', target: 5 },  // Ramp up to 5 users
    { duration: '20s', target: 5 },  // Stay at 5 users
    { duration: '10s', target: 0 },  // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    http_req_failed: ['rate<0.1'],    // Error rate must be below 10%
  },
};

const BASE_URL = 'http://localhost:9090';

export default function () {
  const response = http.get(`${BASE_URL}/health`);
  
  check(response, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 200ms': (r) => r.timings.duration < 200,
    'health check contains service name': (r) => r.body.includes('cloudbridge-relay'),
    'health check contains status ok': (r) => r.body.includes('"status":"ok"'),
  });

  sleep(1); // 1 second between requests
}

export function handleSummary(data) {
  const avgResponseTime = data.metrics.http_req_duration.values.avg;
  const p95ResponseTime = data.metrics.http_req_duration.values['p(95)'];
  const errorRate = data.metrics.http_req_failed.values.rate * 100;
  const throughput = data.metrics.http_reqs.values.rate;
  
  return {
    'simple-load-test-results.json': JSON.stringify(data, null, 2),
    'simple-load-test-summary.txt': `
CloudBridge Relay Simple Load Test Results
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
- HTTP request duration (95th percentile): ${p95ResponseTime < 500 ? 'PASS' : 'FAIL'} (< 500ms)
- HTTP request failure rate: ${errorRate < 10 ? 'PASS' : 'FAIL'} (< 10%)

Overall Status: ${errorRate < 10 && p95ResponseTime < 500 ? 'PASS' : 'FAIL'}

Performance Assessment:
- Response Time: ${avgResponseTime < 50 ? 'Excellent' : avgResponseTime < 100 ? 'Good' : avgResponseTime < 200 ? 'Acceptable' : 'Poor'}
- Throughput: ${throughput > 20 ? 'Excellent' : throughput > 10 ? 'Good' : throughput > 5 ? 'Acceptable' : 'Poor'}
- Reliability: ${errorRate < 1 ? 'Excellent' : errorRate < 5 ? 'Good' : errorRate < 10 ? 'Acceptable' : 'Poor'}
    `,
  };
}

