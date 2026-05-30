import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "15s", target: 10 },
    { duration: "15s", target: 200 },
    { duration: "30s", target: 200 },
    { duration: "15s", target: 10 },
    { duration: "15s", target: 0 },
  ],
  thresholds: {
    http_req_failed: ["rate<0.10"],
    http_req_duration: ["p(95)<750", "p(99)<1200"],
  },
};

const baseURL = __ENV.API_BASE_URL || "http://localhost:8080";
const apiKey = __ENV.API_KEY || "bp_sandbox_full_access_key";

export default function () {
  const response = http.get(`${baseURL}/v1/sandbox/scenarios`, {
    headers: {
      Authorization: `Bearer ${apiKey}`,
      "X-Correlation-ID": `k6-spike-${__VU}-${__ITER}`,
    },
  });
  check(response, {
    "sandbox read is stable or protected": (r) => r.status === 200 || r.status === 429,
  });
  sleep(0.05);
}
