package actionsgithubcom

import (
	"context"

	githubarcv1alpha1 "github.com/actions/actions-runner-controller/apis/actions.github.com/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/wielewout/arc-cleaner/internal/logging"
)

func getEphemeralRunner(ctx context.Context, k8sClient client.Client, namespacedName types.NamespacedName) (*githubarcv1alpha1.EphemeralRunner, error) {
	logger := logging.FromContext(ctx)

	ephemeralRunner := new(githubarcv1alpha1.EphemeralRunner)
	err := k8sClient.Get(
		ctx,
		namespacedName,
		ephemeralRunner,
	)

	if err != nil {
		logger.Error("failed to get ephemeral runners", "error", err.Error())
		return nil, err
	}

	logger.Debug("got ephemeral runners", "name", namespacedName.Name, "namespace", namespacedName.Namespace)
	return ephemeralRunner, nil
}

func getRunnerPod(ctx context.Context, k8sClient client.Client, namespacedName types.NamespacedName) (*corev1.Pod, error) {
	logger := logging.FromContext(ctx)

	pod := new(corev1.Pod)
	err := k8sClient.Get(
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

func getWorkflowPod(ctx context.Context, k8sClient client.Client, namespacedName types.NamespacedName) (*corev1.Pod, error) {
	logger := logging.FromContext(ctx)

	pod := new(corev1.Pod)
	err := k8sClient.Get(
		ctx,
		types.NamespacedName{
			Name:      namespacedName.Name,
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

type deleteWorkflowPodOptions struct {
	dryRun bool
}

func deleteWorkflowPod(ctx context.Context, k8sClient client.Client, workflowPod *corev1.Pod, options deleteWorkflowPodOptions) error {
	logger := logging.FromContext(ctx)

	opts := &client.DeleteOptions{}
	if options.dryRun {
		opts.DryRun = []string{metav1.DryRunAll}
		logger.Debug("dry run to delete worflow pod")
	}
	err := k8sClient.Delete(ctx, workflowPod, opts)
	if err != nil {
		logger.Error("failed deleting workflow pod", "error", err.Error())
		return err
	}
	logger.Info("deleted workflow pod")
	return nil
}

func podStatus(pod *corev1.Pod) corev1.PodPhase {
	return pod.Status.Phase
}
