package signals

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func NewContextWithSignals() context.Context {
	return signals.SetupSignalHandler()
}
