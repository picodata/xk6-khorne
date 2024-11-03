package khorne

import (
	"github.com/dop251/goja"
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

func (p *Chaos) RunChaosExperiment(call goja.FunctionCall, runtime *goja.Runtime) goja.Value {
	namespace := call.Argument(0).String()
	configObj := call.Argument(1).Export()

	manifestObj, ok := configObj.(map[string]interface{})
	if !ok {
		panic("Invalid manifest object")
	}
	manifestStruct := &unstructured.Unstructured{
		Object: manifestObj,
	}

	response := map[string]interface{}{
		"success": true,
		"error":   "",
	}
	if err := chaos.RunChaosExperiment(namespace, manifestStruct); err != nil {
		response["success"] = false
		response["error"] = err.Error()
		return runtime.ToValue(response)
	}

	return runtime.ToValue(response)
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
