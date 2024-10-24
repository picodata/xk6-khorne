package chaos

import (
	"time"
)

// Wait cluster to reconfigure after chaos experiment
func ExperimentSleep(durationStr string) error {
	experimentDuration, err := time.ParseDuration(durationStr)
	if err != nil {
		return err
	}

	time.Sleep(experimentDuration)

	return nil
}
