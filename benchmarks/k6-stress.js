import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "1m", target: 50 },
    { duration: "2m", target: 100 },
    { duration: "1m", target: 150 },
    { duration: "1m", target: 0 },
  ],
  thresholds: {
    http_req_failed: ["rate<0.05"],
    http_req_duration: ["p(95)<500", "p(99)<900"],
  },
};

const baseURL = __ENV.API_BASE_URL || "http://localhost:8080";
const apiKey = __ENV.API_KEY || "bp_sandbox_full_access_key";

export default function () {
  const headers = {
    Authorization: `Bearer ${apiKey}`,
    "X-Correlation-ID": `k6-stress-${__VU}-${__ITER}`,
  };
  const response = http.get(`${baseURL}/v1/accounts/acct_sandbox_001/balance`, { headers });
  check(response, {
    "balance read is stable or rate limited": (r) => r.status === 200 || r.status === 429,
  });
  sleep(0.1);
}
