// This test checks cluster tolerance to pod failures
// We basically disable one slave node from storage replicaset
// And hope that replicaset will return in consistent state in 60s


import { ChaosController } from 'k6/x/chaos/experiments';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';
import http from "k6/http";
import { Trend } from "k6/metrics";
import { respCheck, getDefaultHTTPHeaders, setRandomSeed, getBaseUrl } from "./utils.js";
import exec from "k6/execution";
import vars from "./vars.js";
import { fail } from 'k6';

export const options = Object.assign({}, vars.generalOpts, {
  vus: 1,
  duration: "20s",
});

const latency = new Trend("write_latency");
const NO_RECOVER = false

export function setup() {
  setRandomSeed();
  networkHarass("test-namespace")
  return { httpParams: { headers: getDefaultHTTPHeaders() } };
}

export function teardown(opts) {
  // TODO here goes the check for cluster performance under
  // network failures

  // Creating another instance of controller to
  // delete chaosmesh traces from namespace
  // New instance is created, because of strange
  // memory manipulations of JS

  const chaosController = new ChaosController("test-namespace");
  if (chaosController == null) {
    fail("couldn't create chaos controller instance")
  }

  chaosController.clearExperimentData("networkchaos")

  respCheck(http.del(`${getBaseUrl()}/ucp/0001/perf/plain`, null, opts.httpParams));
}

export default function (opts) {
  let uid = exec.scenario.iterationInTest * (opts.startId || 1);
  let data = `{"id":0,"params":[{"uid":${uid},"p1":1,"p2":2}]}`;
  let resp = http.put(`${getBaseUrl()}/ucp/0001/perf/plain/uid/${uid}`, data, opts.httpParams);
  respCheck(resp, { needFail: false });
  latency.add(resp.timings.waiting);
}

// The killPod function terminates a pod within a Kubernetes cluster according to specifications provided.
export function networkHarass(namespace) {
  const chaosController = new ChaosController(namespace);
  if (chaosController == null) {
    fail("couldn't create chaos controller instance")
  }
  console.log("Stargin exp")
  chaosController.runChaosExperiment("./examples/chaosmesh/master-network-loss.yaml", NO_RECOVER, "1s");
}



