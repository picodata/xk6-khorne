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

func (p *Chaos) RunChaosExperiment(namespace string, configPath string) error {
	return chaos.RunChaosExperiment(namespace, configPath)
}

func (p *Chaos) ExperimentSleep(duration string) {
	chaos.ExperimentSleep(duration)
}

func (p *Chaos) ClearChaosCache(namespace string) string {
	return chaos.ClearChaosCache(namespace)
}

func (p *Chaos) CheckClusterHealth(namespace string) string {
	return chaos.CheckClusterHealth(namespace)
}

func (p *Chaos) CheckPodsHealth(namespace string, pods []string) string {
	return chaos.CheckPodsHealth(namespace, pods)
}

func (p *Chaos) RestartPods(namespace string, pods []string) string {
	return chaos.RestartPods(namespace, pods)
}

func (e *Chaos) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"ExperimentSleep":    e.ExperimentSleep,
			"RunChaosExperiment": e.RunChaosExperiment,
			"ClearChaosCache":    e.ClearChaosCache,
			"CheckClusterHealth": e.CheckClusterHealth,
			"CheckPodsHealth":    e.CheckPodsHealth,
			"RestartPods":        e.RestartPods,
		},
	}
}
