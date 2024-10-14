import { check, fail } from "k6";
import { randomSeed } from "k6";
import http from "k6/http";
import vars from "./vars.js";

let authToken = __ENV.TOKEN;
let seed = Number(__ENV.RANDOM_SEED);

export function getRandom(max) {
  return Math.floor(Math.random() * max) + 1;
}

export function setRandomSeed() {
  seed = seed || getRandom(Number.MAX_SAFE_INTEGER);
  randomSeed(seed);
  console.info(`Generated random seed: ${seed}`);
}

export function respCheck(resp, opts) {
  const { httpCode = 200, test-namespaceCode = 200, dataChecker = null, needFail = true } = opts || {};
  let checks = {
    "is status ok": (r) => r.status === httpCode,
  };
  if (test-namespaceCode) {
    checks["is result code ok"] = (r) => r.body && r.json().status.code === test-namespaceCode;
  }
  if (dataChecker) {
    checks["are returned data expected"] = (r) => r.body && dataChecker(r.json().data);
  }
  let result = check(resp, checks);
  if (!result) {
    console.warn(resp.body);
    if (needFail) fail();
  }
  return result;
}

export function getBaseUrl() {
  let node = typeof __ITER == "undefined" ? vars.hosts[0] : vars.hosts[__ITER % vars.hosts.length];
  return `http://${node}`;
}

function getAuthToken(expired_at) {
  if (expired_at === undefined) {
    expired_at = Math.round(Date.now() / 1000) + 3600;
  };
  if (authToken) {
    return authToken;
  }
  let resp = http.put(`${getBaseUrl()}/admin/bauth/users/test:perf/password`, '{ "password": "pwd" }');
  respCheck(resp, { httpCode: 201, test-namespaceCode: null });

  resp = http.put(`${getBaseUrl()}/admin/bauth/users/test:perf/token`, '{ "expires_at": 8015414400 }');
  respCheck(resp, { httpCode: 201, test-namespaceCode: null });
  if (!resp.body) fail();

  authToken = resp.json().token;
  if (!authToken) fail();

  return authToken;
}

export function getDefaultHTTPHeaders(expired_at) {
  return {
    "Content-Type": "application/json",
    "Bauth-token": getAuthToken(expired_at),
  };
}
