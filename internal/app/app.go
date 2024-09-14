package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	actionsgithubcom "github.com/wielewout/arc-cleaner/internal/actions.github.com"
	"github.com/wielewout/arc-cleaner/internal/kubernetes"
	"github.com/wielewout/arc-cleaner/internal/logging"
)

type App struct {
	listenAddress string
	k8sClient     *kubernetes.Client
	namespace     string
	period        time.Duration
	dryRun        bool
}

func New(k8sClient *kubernetes.Client, opts ...Option) *App {
	app := &App{
		listenAddress: ":8080",
		k8sClient:     k8sClient,
		namespace:     "default",
		period:        30 * time.Second,
		dryRun:        false,
	}

	for _, opt := range opts {
		opt.apply(app)
	}

	return app
}

func (app App) Start(ctx context.Context) {
	logger := logging.FromContext(ctx)

	ready := false
	go startHttpServer(ctx, app.listenAddress, &ready)

	ticker := time.NewTicker(app.period)
	logger.Debug("started periodic timer", "period", app.period)

	ready = true

	app.reconcile(ctx)

	for {
		select {
		case <-ticker.C:
			logger.Debug("triggered periodic timer")
			app.reconcile(ctx)
		case <-ctx.Done():
			ticker.Stop()
			logger.Debug("stopped periodic timer")
			return
		}
	}
}

func startHttpServer(ctx context.Context, listenAddress string, ready *bool) {
	logger := logging.FromContext(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if *ready {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	logger.Info("serving", "listenAddress", listenAddress)
	err := http.ListenAndServe(listenAddress, mux)
	if err != nil {
		slog.Error("failed listening and serving http", "error", err)
	}
}

func (app App) reconcile(ctx context.Context) {
	podList := app.getPodList(ctx)
	for _, pod := range podList.Items {
		controller := actionsgithubcom.NewWorkflowPodReconciler(
			app.k8sClient,
			actionsgithubcom.WithDryRun(app.dryRun),
		)
		_ = controller.Reconcile(ctx, types.NamespacedName{
			Name:      pod.GetName(),
			Namespace: pod.GetNamespace(),
		})
	}
}

func (app App) getPodList(ctx context.Context) *corev1.PodList {
	logger := logging.FromContext(ctx).
		With("namespace", app.namespace)

	podList := new(corev1.PodList)
	err := app.k8sClient.List(
		ctx,
		podList,
		client.InNamespace(app.namespace),
	)

	if err != nil {
		logger.Error("failed to list pods", "error", err.Error())
		return podList
	}

	logger.Debug("listed pods", "length", len(podList.Items))
	return podList
}
