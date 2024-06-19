package actionsgithubcom

type EphemeralRunnerReconcilerOption interface {
	apply(*EphemeralRunnerReconciler)
}

type optionWithDryRun struct {
	dryRun bool
}

func (o optionWithDryRun) apply(c *EphemeralRunnerReconciler) {
	c.DryRun = o.dryRun
}

func WithDryRun(dryRun bool) optionWithDryRun {
	return optionWithDryRun{
		dryRun: dryRun,
	}
}
