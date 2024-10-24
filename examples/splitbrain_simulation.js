// This test checks cluster tolerance to pod failures
// We basically disable one slave node from storage replicaset
// And hope that replicaset will return in consistent state

import k6 from 'k6';
import khorne from "k6/x/khorne";

export const options = Object.assign({}, {}, {
  vus: 1,
  duration: "50s",
  iterations: 1
});

export default function (opts) {
  khorne.RunChaosExperiment("test-namespace", "./examples/chaosmesh/kill_master_node.yaml")
  khorne.ExperimentSleep("20s")
  khorne.ClearChaosCache("test-namespace")
  let result = khorne.CheckPodsHealth("test-namespace", ["storage-0-1"])

  if (result.success) {
    k6.fail("Master node became consistent, however new master has been elected. Splitbrain wasn't detected. " + result.error)
  } else {
    console.log("Splitbrain has been handled, old master status: " + result.error)
  }

}



