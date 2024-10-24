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
  khorne.RunChaosExperiment("test-namespace", "./examples/chaosmesh/kill_one_slave_node.yaml")
  khorne.ExperimentSleep("5s")
  khorne.ClearChaosCache("test-namespace")
  let result = khorne.CheckPodsHealth("test-namespace", ["storage-0-2"])

  if (!result.success) {
    k6.fail("Node didn't recover in time, error: " + result)
  }
}



