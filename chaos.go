package chaos

import (
	"fmt"

	_ "xk6-khorne/internal/experiments" // Register the experiments module as well
	_ "xk6-khorne/internal/k8s"         // Register the k8s module as well

	"go.k6.io/k6/js/modules"
)

const version = "v0.0.2"

func init() {
	modules.Register("k6/x/chaos", &Chaos{
		Version: version,
	})
	fmt.Println("Running k6io/xk6-chaos@$" + version)
}

// Chaos is the main export of the chaos engineering extension
type Chaos struct {
	Version string
}
