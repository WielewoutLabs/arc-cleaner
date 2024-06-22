package app

import "time"

type Option interface {
	apply(*App)
}

type optionWithNamespace struct {
	namespace string
}

func (o optionWithNamespace) apply(a *App) {
	a.namespace = o.namespace
}

func WithNamespace(namespace string) optionWithNamespace {
	return optionWithNamespace{
		namespace: namespace,
	}
}

type optionWithPeriod struct {
	period time.Duration
}

func (o optionWithPeriod) apply(a *App) {
	a.period = o.period
}

func WithPeriod(period time.Duration) optionWithPeriod {
	return optionWithPeriod{
		period: period,
	}
}

type optionWithDryRun struct {
	dryRun bool
}

func (o optionWithDryRun) apply(a *App) {
	a.dryRun = o.dryRun
}

func WithDryRun(dryRun bool) optionWithDryRun {
	return optionWithDryRun{
		dryRun: dryRun,
	}
}
