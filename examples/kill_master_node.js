// In this test we disable master node of replicaset
// Master node is excepted to enter the inconsistent state

import k6 from "k6";
import khorne from "k6/x/khorne";

export const options = Object(
  {},
  {},
  {
    vus: 1,
    iterations: 1,
  }
);

export default function (opts) {
  khorne.RunChaosExperiment(
    "test-namespace",
    "./examples/chaosmesh/kill_master_node.yaml"
  );
  khorne.Sleep("20s");
  khorne.ClearChaosCache("test-namespace");
  let result = khorne.CheckPodsHealth("test-namespace", ["storage-0-1"]);

  if (result.success) {
    k6.fail(
      "Master node became consistent, however new master has been elected. " +
        result.error
    );
  } else {
    console.log(
      "Everything went as planned. Old master status: " + result.error
    );
  }
}
