package actionsgithubcom

type Option interface {
	apply(*WorkflowPodReconciler)
}

type optionWithDryRun struct {
	dryRun bool
}

func (o optionWithDryRun) apply(c *WorkflowPodReconciler) {
	c.DryRun = o.dryRun
}

func WithDryRun(dryRun bool) optionWithDryRun {
	return optionWithDryRun{
		dryRun: dryRun,
	}
}
