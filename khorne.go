package khorne

import (
	"fmt"

	chaos "github.com/picodata/xk6-khorne/src/chaos"
	"go.k6.io/k6/js/modules"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	modules.Register("k6/x/khorne", &ChaosRoot{})
}

var _ modules.Module = ChaosRoot{}

type ChaosRoot struct{}

func (k ChaosRoot) NewModuleInstance(vu modules.VU) modules.Instance {
	return &Chaos{
		vu: vu,
	}
}

// This exposes khorne metadata for use in displaying results.
type Chaos struct {
	vu modules.VU
}

type Result struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (p *Chaos) RunChaosExperiment(namespace string, manifest map[string]interface{}) (bool, error) {
	// Fill up constant fields in manifest to save some user time
	manifest["apiVersion"] = "chaos-mesh.org/v1alpha1"
	manifest["metadata"] = map[string]interface{}{
		"namespace": namespace,
		"name":      "chaos-experiment",
	}

	if spec, ok := manifest["spec"].(map[string]interface{}); !ok {
		return false, fmt.Errorf("spec field is missing in experiment manifest")
	} else {
		if selector, ok := spec["selector"].(map[string]interface{}); !ok {
			return false, fmt.Errorf("selector field is missing in experiment manifest")
		} else {
			selector["namespaces"] = []string{namespace}
		}
	}
	// Create chaos experiment object and pass it to executor
	manifestStruct := &unstructured.Unstructured{
		Object: manifest,
	}
	if err := chaos.RunChaosExperiment(namespace, manifestStruct); err != nil {
		return false, err
	}

	return true, nil
}

func (p *Chaos) RunChaosExperimentFile(namespace string, manifestPath string) Result {
	if err := chaos.RunChaosExperimentFile(namespace, manifestPath); err != nil {
		return Result{false, err.Error()}
	}
	return Result{Success: true}
}

func (p *Chaos) Sleep(duration string) Result {
	if err := chaos.Sleep(duration); err != nil {
		return Result{false, err.Error()}
	}
	return Result{Success: true}

}

func (p *Chaos) ClearChaosCache(namespace string) Result {
	if err := chaos.ClearChaosCache(namespace); err != nil {
		return Result{false, err.Error()}
	}
	return Result{Success: true}
}

func (p *Chaos) CheckClusterHealth(namespace string) Result {
	if err := chaos.CheckClusterHealth(namespace); err != nil {
		return Result{false, err.Error()}
	}
	return Result{Success: true}
}

func (p *Chaos) CheckPodsHealth(namespace string, pods []string) Result {
	if err := chaos.CheckClusterHealth(namespace); err != nil {
		return Result{false, err.Error()}
	}
	return Result{Success: true}
}

func (p *Chaos) RestartPods(namespace string, pods []string) Result {
	if err := chaos.RestartPods(namespace, pods); err != nil {
		return Result{false, err.Error()}
	}
	return Result{Success: true}
}

func (e *Chaos) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"Sleep":                  e.Sleep,
			"RunChaosExperiment":     e.RunChaosExperiment,
			"RunChaosExperimentFile": e.RunChaosExperimentFile,
			"ClearChaosCache":        e.ClearChaosCache,
			"CheckClusterHealth":     e.CheckClusterHealth,
			"CheckPodsHealth":        e.CheckPodsHealth,
			"RestartPods":            e.RestartPods,
		},
	}
}
