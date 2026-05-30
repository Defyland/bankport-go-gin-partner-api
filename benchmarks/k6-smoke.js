import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 1,
  duration: "30s",
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<200"],
  },
};

const baseURL = __ENV.API_BASE_URL || "http://localhost:8080";
const apiKey = __ENV.API_KEY || "bp_sandbox_full_access_key";

export default function () {
  const headers = {
    Authorization: `Bearer ${apiKey}`,
    "Content-Type": "application/json",
    "X-Correlation-ID": `k6-smoke-${__VU}-${__ITER}`,
  };
  const balance = http.get(`${baseURL}/v1/accounts/acct_sandbox_001/balance`, { headers });
  check(balance, {
    "balance is 200": (r) => r.status === 200,
    "balance has request id": (r) => Boolean(r.json("request_id")),
  });

  const transfer = http.post(
    `${baseURL}/v1/pix/transfers`,
    JSON.stringify({
      source_account_id: "acct_sandbox_001",
      amount_cents: 1,
      currency: "BRL",
      pix_key: "merchant@example.com",
      description: "k6 smoke transfer",
    }),
    { headers: { ...headers, "Idempotency-Key": `smoke-${__VU}-${__ITER}` } },
  );
  check(transfer, {
    "pix transfer accepted": (r) => r.status === 202,
  });
  sleep(1);
}
