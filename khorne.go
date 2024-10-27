package khorne

import (
	chaos "github.com/picodata/xk6-khorne/src/chaos"
	"go.k6.io/k6/js/modules"
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

func (p *Chaos) RunChaosExperiment(namespace string, configPath string) Result {
	if err := chaos.RunChaosExperiment(namespace, configPath); err != nil {
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
			"Sleep":              e.Sleep,
			"RunChaosExperiment": e.RunChaosExperiment,
			"ClearChaosCache":    e.ClearChaosCache,
			"CheckClusterHealth": e.CheckClusterHealth,
			"CheckPodsHealth":    e.CheckPodsHealth,
			"RestartPods":        e.RestartPods,
		},
	}
}
