package experiments

import (
	"log"
	"xk6-khorne/internal/experiments/chaos_controller"

	"github.com/dop251/goja"
	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/chaos/experiments", &ExperimentsRoot{})
}

var _ modules.Module = ExperimentsRoot{}

type ExperimentsRoot struct{}

func (k ExperimentsRoot) NewModuleInstance(vu modules.VU) modules.Instance {
	return &Experiments{
		vu: vu,
	}
}

// This exposes experiment metadata for use in displaying results.
type Experiments struct {
	ChaosController *chaos_controller.ChaosController
	vu              modules.VU
}

func (e *Experiments) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"ChaosController": e.XChaosController,
		},
	}
}

// XChaosController serves as a constructor of the ChaosController js class
// Expected constructor arguments:
// 1) namespace: string
func (e *Experiments) XChaosController(call goja.ConstructorCall) *goja.Object {
	rt := e.vu.Runtime()

	// parse cmd. line arguments
	args := call.Arguments
	var namespace string
	if len(args) > 0 {
		namespace = args[0].String()
	} else {
		log.Fatalf("no arguments passed to chaos controller constructor")
		return nil
	}

	p := chaos_controller.New(namespace)

	return rt.ToValue(p).ToObject(rt)
}
