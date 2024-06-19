package actionsgithubcom

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/wielewout/arc-cleaner/internal/logging"
)

type EphemeralRunnerReconciler struct {
	k8sClient client.Client
}

func NewEphemeralRunnerReconciler(k8sClient client.Client) *EphemeralRunnerReconciler {
	return &EphemeralRunnerReconciler{
		k8sClient: k8sClient,
	}
}

func (r *EphemeralRunnerReconciler) Reconcile(ctx context.Context, namespacedName types.NamespacedName) error {
	logger := logging.FromContext(ctx).
		With("namespace", namespacedName.Namespace).
		With("ephemeral-runner", namespacedName.Name)

	runnerPod, err := r.getRunnerPod(ctx, namespacedName)
	if err != nil {
		logger.Debug("runner pod does not exist")
	} else {
		runnerPodStatus := getPodStatus(runnerPod)
		logger.Debug(fmt.Sprintf("runner pod status is %s", strings.ToLower(string(runnerPodStatus))))
		if runnerPodStatus != corev1.PodPending {
			logger.Debug("skipping", "reason", "runner pod is not pending")
			return nil
		}
	}

	workflowPod, err := r.getWorkflowPod(ctx, namespacedName)
	if err != nil {
		logger.Debug("skipping", "reason", "workflow pod does not exist", "error", err.Error())
		return nil
	}

	workflowPodStatus := getPodStatus(workflowPod)
	logger.Debug(fmt.Sprintf("workflow pod status is %s", strings.ToLower(string(workflowPodStatus))))

	return r.deleteWorkflowPod(ctx, workflowPod)
}

func (r *EphemeralRunnerReconciler) getRunnerPod(ctx context.Context, namespacedName types.NamespacedName) (*corev1.Pod, error) {
	logger := logging.FromContext(ctx)

	pod := new(corev1.Pod)
	err := r.k8sClient.Get(
		ctx,
		types.NamespacedName{
			Namespace: namespacedName.Namespace,
			Name:      namespacedName.Name,
		},
		pod,
	)
	if err != nil {
		logger.Debug("unable to get runner pod", "error", err.Error())
		return nil, err
	}

	return pod, nil
}

func (r *EphemeralRunnerReconciler) getWorkflowPod(ctx context.Context, namespacedName types.NamespacedName) (*corev1.Pod, error) {
	logger := logging.FromContext(ctx)

	pod := new(corev1.Pod)
	err := r.k8sClient.Get(
		ctx,
		types.NamespacedName{
			Name:      fmt.Sprintf("%s-workflow", namespacedName.Name),
			Namespace: namespacedName.Namespace,
		},
		pod,
	)
	if err != nil {
		logger.Debug("unable to get workflow pod", "error", err.Error())
		return nil, err
	}

	return pod, nil
}

func getPodStatus(pod *corev1.Pod) corev1.PodPhase {
	return pod.Status.Phase
}

func (r *EphemeralRunnerReconciler) deleteWorkflowPod(ctx context.Context, workflowPod *corev1.Pod) error {
	logger := logging.FromContext(ctx)

	opts := &client.DeleteOptions{}
	if viper.GetBool("dryrun") {
		opts.DryRun = []string{metav1.DryRunAll}
		logger.Debug("dry run to delete worflow pod")
	}
	err := r.k8sClient.Delete(ctx, workflowPod, opts)
	if err != nil {
		logger.Error("failed deleting workflow pod", "error", err.Error())
		return err
	}
	logger.Info("deleted workflow pod")
	return nil
}
