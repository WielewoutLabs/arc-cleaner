package actionsgithubcom

import (
	"context"
	"fmt"
	"strings"

	"github.com/wielewout/arc-cleaner/internal/logging"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkflowPodReconciler struct {
	k8sClient client.Client
	DryRun    bool
}

func NewWorkflowPodReconciler(k8sClient client.Client, opts ...Option) *WorkflowPodReconciler {
	controller := &WorkflowPodReconciler{
		k8sClient: k8sClient,
		DryRun:    false,
	}

	for _, opt := range opts {
		opt.apply(controller)
	}

	return controller
}

func (r *WorkflowPodReconciler) Reconcile(ctx context.Context, namespacedName types.NamespacedName) error {
	logger := logging.FromContext(ctx).
		With("namespace", namespacedName.Namespace).
		With("pod-name", namespacedName.Name)

	logger.Debug("workflow pod reconciler")

	if !strings.HasSuffix(namespacedName.Name, "-workflow") {
		logger.Debug("skipping", "reason", "pod is no workflow pod without 'workflow' suffix in name")
		return nil
	}

	workflowPod, err := getWorkflowPod(ctx, r.k8sClient, namespacedName)
	if err != nil {
		logger.Warn("skipping", "reason", "workflow pod does not exist")
		return nil
	}

	runnerPodName, ok := workflowPod.ObjectMeta.Labels["runner-pod"]
	if !ok {
		logger.Debug("skipping", "reason", "pod is no workflow pod without 'runner-pod' label")
		return nil
	}

	runnerNamespacedName := types.NamespacedName{
		Name:      runnerPodName,
		Namespace: namespacedName.Namespace,
	}

	_, err = getEphemeralRunner(ctx, r.k8sClient, runnerNamespacedName)
	if err != nil {
		logger.Debug("ephemeral runner does not exist")
	} else {
		logger.Debug("ephemeral runner exists")

		runnerPod, err := getRunnerPod(ctx, r.k8sClient, runnerNamespacedName)
		if err != nil {
			logger.Debug("runner pod does not exist")
		} else {
			logger.Debug("runner pod exists")

			runnerPodStatus := podStatus(runnerPod)
			logger.Debug(fmt.Sprintf("runner pod status is %s", strings.ToLower(string(runnerPodStatus))))
			if runnerPodStatus != corev1.PodPending {
				logger.Debug("skipping", "reason", "runner pod is not pending")
				return nil
			}
		}
	}

	return deleteWorkflowPod(ctx, r.k8sClient, workflowPod, deleteWorkflowPodOptions{dryRun: r.DryRun})
}
