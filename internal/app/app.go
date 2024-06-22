package app

import (
	"context"

	githubarcv1alpha1 "github.com/actions/actions-runner-controller/apis/actions.github.com/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	actionsgithubcom "github.com/wielewout/arc-cleaner/internal/actions.github.com"
	"github.com/wielewout/arc-cleaner/internal/kubernetes"
	"github.com/wielewout/arc-cleaner/internal/logging"
)

type App struct {
	k8sClient *kubernetes.Client
	namespace string
	dryRun    bool
}

func New(k8sClient *kubernetes.Client, opts ...Option) *App {
	app := &App{
		k8sClient: k8sClient,
		namespace: "default",
		dryRun:    false,
	}

	for _, opt := range opts {
		opt.apply(app)
	}

	return app
}

func (app App) Start(ctx context.Context) {
	ephemeralRunnerList := app.getEphemeralRunnerList(ctx)
	for _, ephemeralRunner := range ephemeralRunnerList.Items {
		controller := actionsgithubcom.NewEphemeralRunnerReconciler(
			app.k8sClient,
			actionsgithubcom.WithDryRun(app.dryRun),
		)
		controller.Reconcile(ctx, types.NamespacedName{
			Name:      ephemeralRunner.GetName(),
			Namespace: ephemeralRunner.GetNamespace(),
		})
	}
}

func (app App) getEphemeralRunnerList(ctx context.Context) *githubarcv1alpha1.EphemeralRunnerList {
	logger := logging.FromContext(ctx).
		With("namespace", app.namespace)

	ephemeralRunnerList := new(githubarcv1alpha1.EphemeralRunnerList)
	err := app.k8sClient.List(
		ctx,
		ephemeralRunnerList,
		client.InNamespace(app.namespace),
	)

	if err != nil {
		logger.Error("failed to list ephemeral runners", "error", err.Error())
		return ephemeralRunnerList
	}

	logger.Debug("listed ephemeral runners", "length", len(ephemeralRunnerList.Items))
	return ephemeralRunnerList
}
