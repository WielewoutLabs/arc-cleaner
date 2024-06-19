package actionsgithubcom_test

import (
	"context"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/stretchr/testify/suite"
	actionsgithubcom "github.com/wielewout/arc-cleaner/internal/actions.github.com"
)

func TestEphemeralRunnerReconciler(t *testing.T) {
	suite.Run(t, &EphemeralRunnerReconcilerTestSuite{})
}

type EphemeralRunnerReconcilerTestSuite struct {
	suite.Suite
	k8sClient              client.WithWatch
	runnerNamespacedName   types.NamespacedName
	workflowNamespacedName types.NamespacedName
}

func (ts *EphemeralRunnerReconcilerTestSuite) SetupTest() {
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

	_, err := createEphemeralRunner(ts.k8sClient, ts.runnerNamespacedName)
	ts.Require().NoError(err)
}

func (ts *EphemeralRunnerReconcilerTestSuite) requireNoPod(namespacedName types.NamespacedName) {
	pod := new(corev1.Pod)
	err := ts.k8sClient.Get(context.Background(), namespacedName, pod)
	ts.Require().Error(err)
}

func (ts *EphemeralRunnerReconcilerTestSuite) requirePod(namespacedName types.NamespacedName) {
	pod := new(corev1.Pod)
	err := ts.k8sClient.Get(context.Background(), namespacedName, pod)
	ts.Require().NoError(err)
}

func (ts *EphemeralRunnerReconcilerTestSuite) TestShouldDeleteWorkflowPodWhenRunnerPending() {
	_, err := createRunnerPod(ts.k8sClient, ts.runnerNamespacedName, corev1.PodPending)
	ts.Require().NoError(err)

	_, err = createWorkflowPod(ts.k8sClient, ts.runnerNamespacedName)
	ts.Require().NoError(err)

	controller := actionsgithubcom.NewEphemeralRunnerReconciler(ts.k8sClient)
	err = controller.Reconcile(context.Background(), ts.runnerNamespacedName)
	ts.Require().NoError(err)

	ts.requireNoPod(ts.workflowNamespacedName)
}

func (ts *EphemeralRunnerReconcilerTestSuite) TestShouldNotErrorWhenWorkflowPodNonExistent() {
	_, err := createRunnerPod(ts.k8sClient, ts.runnerNamespacedName, corev1.PodPending)
	ts.Require().NoError(err)

	controller := actionsgithubcom.NewEphemeralRunnerReconciler(ts.k8sClient)
	err = controller.Reconcile(context.Background(), ts.runnerNamespacedName)
	ts.Require().NoError(err)
}

func (ts *EphemeralRunnerReconcilerTestSuite) TestShouldDeleteWorkflowPodWhenRunnerNonExistent() {
	_, err := createWorkflowPod(ts.k8sClient, ts.runnerNamespacedName)
	ts.Require().NoError(err)

	controller := actionsgithubcom.NewEphemeralRunnerReconciler(ts.k8sClient)
	err = controller.Reconcile(context.Background(), ts.runnerNamespacedName)
	ts.Require().NoError(err)

	ts.requireNoPod(ts.workflowNamespacedName)
}

func (ts *EphemeralRunnerReconcilerTestSuite) TestShouldKeepWorkflowPodWhenRunnerRunning() {
	_, err := createRunnerPod(ts.k8sClient, ts.runnerNamespacedName, corev1.PodRunning)
	ts.Require().NoError(err)

	_, err = createWorkflowPod(ts.k8sClient, ts.runnerNamespacedName)
	ts.Require().NoError(err)

	controller := actionsgithubcom.NewEphemeralRunnerReconciler(ts.k8sClient)
	err = controller.Reconcile(context.Background(), ts.runnerNamespacedName)
	ts.Require().NoError(err)

	ts.requirePod(ts.workflowNamespacedName)
}
