package steps

import (
	"context"

	"github.com/cucumber/godog"
	"github.com/wielewout/arc-cleaner/test/acceptance/dsl"
	"github.com/wielewout/arc-cleaner/test/acceptance/system"
)

func InitializeHealthScenario(ctx *godog.ScenarioContext, systemConfig system.Config) {
	health := &health{
		dsl: dsl.NewDsl(systemConfig),
	}

	ctx.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
		health.err = nil
		return ctx, nil
	})

	ctx.Step(`^a liveness request comes in$`, health.aLivenessRequestComesIn)
	ctx.Step(`^a readiness request comes in$`, health.aReadinessRequestComesIn)
	ctx.Then(`^a successful health response is returned$`, health.aSuccessfulHealthResponseIsReturned)
}

type health struct {
	dsl dsl.Dsl
	err error
}

func (h *health) aLivenessRequestComesIn() error {
	h.err = h.dsl.Health.Live()
	return nil
}

func (h *health) aReadinessRequestComesIn() error {
	h.err = h.dsl.Health.Ready()
	return nil
}

func (h *health) aSuccessfulHealthResponseIsReturned(ctx context.Context) error {
	return h.err
}
