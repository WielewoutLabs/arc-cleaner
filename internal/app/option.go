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
