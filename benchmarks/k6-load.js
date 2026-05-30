import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "1m", target: 20 },
    { duration: "3m", target: 20 },
    { duration: "1m", target: 0 },
  ],
  thresholds: {
    http_req_failed: ["rate<0.02"],
    http_req_duration: ["p(95)<250", "p(99)<500"],
  },
};

const baseURL = __ENV.API_BASE_URL || "http://localhost:8080";
const apiKey = __ENV.API_KEY || "bp_sandbox_full_access_key";

export default function () {
  const headers = {
    Authorization: `Bearer ${apiKey}`,
    "Content-Type": "application/json",
    "X-Correlation-ID": `k6-load-${__VU}-${__ITER}`,
  };

  const read = http.get(`${baseURL}/v1/accounts/acct_sandbox_001/balance`, { headers });
  check(read, { "balance read ok": (r) => r.status === 200 });

  if (__ITER % 5 === 0) {
    const write = http.post(
      `${baseURL}/v1/pix/transfers`,
      JSON.stringify({
        source_account_id: "acct_sandbox_001",
        amount_cents: 1,
        currency: "BRL",
        pix_key: "merchant@example.com",
        description: "k6 load transfer",
      }),
      { headers: { ...headers, "Idempotency-Key": `load-${__VU}-${__ITER}` } },
    );
    check(write, { "pix write accepted": (r) => r.status === 202 });
  }
  sleep(0.2);
}
