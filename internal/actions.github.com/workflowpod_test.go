package actionsgithubcom_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	actionsgithubcom "github.com/wielewout/arc-cleaner/internal/actions.github.com"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestWorkflowPodReconciler(t *testing.T) {
	suite.Run(t, &WorkflowPodReconcilerTestSuite{})
}

type WorkflowPodReconcilerTestSuite struct {
	suite.Suite
	k8sClient              client.WithWatch
	runnerNamespacedName   types.NamespacedName
	workflowNamespacedName types.NamespacedName
}

func (ts *WorkflowPodReconcilerTestSuite) SetupTest() {
	ts.k8sClient = fake.NewClientBuilder().
		WithScheme(getScheme()).
		Build()

	runnerName := "gha-runner"
	namespace := "default"
	ts.runnerNamespacedName = types.NamespacedName{
		Name:      runnerName,
		Namespace: namespace,
	}
	ts.workflowNamespacedName = types.NamespacedName{
		Name:      fmt.Sprintf("%s-workflow", runnerName),
		Namespace: namespace,
	}

	_, err := createWorkflowPod(ts.k8sClient, ts.workflowNamespacedName)
	ts.Require().NoError(err)
}

func (ts *WorkflowPodReconcilerTestSuite) requireNoPod(namespacedName types.NamespacedName) {
	pod := new(corev1.Pod)
	err := ts.k8sClient.Get(context.Background(), namespacedName, pod)
	ts.Require().Error(err)
}

func (ts *WorkflowPodReconcilerTestSuite) requirePod(namespacedName types.NamespacedName) {
	pod := new(corev1.Pod)
	err := ts.k8sClient.Get(context.Background(), namespacedName, pod)
	ts.Require().NoError(err)
}

func (ts *WorkflowPodReconcilerTestSuite) TestShouldKeepRunnerPodWhenRunnerPodNameProvidedToReconciler() {
	_, err := createRunnerPod(ts.k8sClient, ts.runnerNamespacedName, corev1.PodRunning)
	ts.Require().NoError(err)

	controller := actionsgithubcom.NewWorkflowPodReconciler(ts.k8sClient)
	err = controller.Reconcile(context.Background(), ts.runnerNamespacedName)
	ts.Require().NoError(err)

	ts.requirePod(ts.runnerNamespacedName)
}

func (ts *WorkflowPodReconcilerTestSuite) TestShouldKeepNonWorkflowPodWithNameEndingWithWorkflowWhenProvidedToReconciler() {
	podNamespacedName := types.NamespacedName{
		Name:      "some-pod-workflow",
		Namespace: ts.workflowNamespacedName.Namespace,
	}
	_, err := createSomePod(ts.k8sClient, podNamespacedName)
	ts.Require().NoError(err)

	controller := actionsgithubcom.NewWorkflowPodReconciler(ts.k8sClient)
	err = controller.Reconcile(context.Background(), podNamespacedName)
	ts.Require().NoError(err)

	ts.requirePod(podNamespacedName)
}

func (ts *WorkflowPodReconcilerTestSuite) TestShouldDeleteWorkflowPodWhenRunnerNonExistent() {
	controller := actionsgithubcom.NewWorkflowPodReconciler(ts.k8sClient)
	err := controller.Reconcile(context.Background(), ts.workflowNamespacedName)
	ts.Require().NoError(err)

	ts.requireNoPod(ts.workflowNamespacedName)
}

func (ts *WorkflowPodReconcilerTestSuite) TestShouldDeleteWorkflowPodWhenRunnerExistsButRunnerPodNonExistent() {
	_, err := createEphemeralRunner(ts.k8sClient, ts.runnerNamespacedName)
	ts.Require().NoError(err)

	controller := actionsgithubcom.NewWorkflowPodReconciler(ts.k8sClient)
	err = controller.Reconcile(context.Background(), ts.workflowNamespacedName)
	ts.Require().NoError(err)

	ts.requireNoPod(ts.workflowNamespacedName)
}

func (ts *WorkflowPodReconcilerTestSuite) TestShouldDeleteWorkflowPodWhenRunnerExistsAndRunnerPodPending() {
	_, err := createEphemeralRunner(ts.k8sClient, ts.runnerNamespacedName)
	ts.Require().NoError(err)

	_, err = createRunnerPod(ts.k8sClient, ts.runnerNamespacedName, corev1.PodPending)
	ts.Require().NoError(err)

	controller := actionsgithubcom.NewWorkflowPodReconciler(ts.k8sClient)
	err = controller.Reconcile(context.Background(), ts.workflowNamespacedName)
	ts.Require().NoError(err)

	ts.requireNoPod(ts.workflowNamespacedName)
}

func (ts *WorkflowPodReconcilerTestSuite) TestShouldKeepWorkflowPodWhenRunnerExistsAndRunnerRunning() {
	_, err := createEphemeralRunner(ts.k8sClient, ts.runnerNamespacedName)
	ts.Require().NoError(err)

	_, err = createRunnerPod(ts.k8sClient, ts.runnerNamespacedName, corev1.PodRunning)
	ts.Require().NoError(err)

	controller := actionsgithubcom.NewWorkflowPodReconciler(ts.k8sClient)
	err = controller.Reconcile(context.Background(), ts.workflowNamespacedName)
	ts.Require().NoError(err)

	ts.requirePod(ts.workflowNamespacedName)
}

func (ts *WorkflowPodReconcilerTestSuite) TestShouldNotErrorWhenWorkflowPodAlreadyDeleted() {
	err := deleteWorkflowPod(ts.k8sClient, ts.workflowNamespacedName)
	ts.Require().NoError(err)
	ts.requireNoPod(ts.workflowNamespacedName)

	controller := actionsgithubcom.NewWorkflowPodReconciler(ts.k8sClient)
	err = controller.Reconcile(context.Background(), ts.workflowNamespacedName)

	ts.Require().NoError(err)
}

func (ts *WorkflowPodReconcilerTestSuite) TestShouldKeepWorkflowPodWhenDryRun() {
	controller := actionsgithubcom.NewWorkflowPodReconciler(ts.k8sClient, actionsgithubcom.WithDryRun(true))
	err := controller.Reconcile(context.Background(), ts.workflowNamespacedName)
	ts.Require().NoError(err)

	ts.requirePod(ts.workflowNamespacedName)
}
