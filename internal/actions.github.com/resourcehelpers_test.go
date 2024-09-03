package actionsgithubcom_test

import (
	"context"
	"fmt"
	"strings"

	githubv1alpha1 "github.com/actions/actions-runner-controller/apis/actions.github.com/v1alpha1"
	actionsgithubcom "github.com/actions/actions-runner-controller/controllers/actions.github.com"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	runnerImage        = "ghcr.io/actions/actions-runner:latest"
	repoUrl            = "https://github.com/owner/repo"
	defaultGitHubToken = "gh_token"
)

func getScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = githubv1alpha1.AddToScheme(scheme)
	return scheme
}

func createEphemeralRunner(k8sClient client.Client, namespacedName types.NamespacedName) (*githubv1alpha1.EphemeralRunner, error) {
	_, err := createConfigSecret(k8sClient, namespacedName)
	if err != nil {
		return nil, err
	}

	ephemeralRunner := newEphemeralRunner(namespacedName)
	err = k8sClient.Create(context.Background(), ephemeralRunner)
	if err != nil {
		return nil, err
	}

	return ephemeralRunner, nil
}

func createConfigSecret(k8sClient client.Client, namespacedName types.NamespacedName) (*corev1.Secret, error) {
	configSecret := newConfigSecret(namespacedName)
	err := k8sClient.Create(context.Background(), configSecret)
	if err != nil {
		return nil, err
	}

	return configSecret, nil
}

func createRunnerPod(k8sClient client.Client, namespacedName types.NamespacedName, status corev1.PodPhase) (*corev1.Pod, error) {
	runnerPod := newRunnerPod(namespacedName, status)
	err := k8sClient.Create(context.Background(), runnerPod)
	if err != nil {
		return nil, err
	}

	return runnerPod, nil
}

func createWorkflowPod(k8sClient client.Client, namespacedName types.NamespacedName) (*corev1.Pod, error) {
	workflowPod := newWorkflowPod(namespacedName)
	err := k8sClient.Create(context.Background(), workflowPod)
	if err != nil {
		return nil, err
	}

	return workflowPod, nil
}

func createSomePod(k8sClient client.Client, namespacedName types.NamespacedName) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "web",
					Image: "nginx:latest",
				},
			},
		},
	}
	err := k8sClient.Create(context.Background(), pod)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func newEphemeralRunner(namespacedName types.NamespacedName) *githubv1alpha1.EphemeralRunner {
	return &githubv1alpha1.EphemeralRunner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
		Spec: githubv1alpha1.EphemeralRunnerSpec{
			GitHubConfigUrl:    repoUrl,
			GitHubConfigSecret: fmt.Sprintf("%s-secret-config", namespacedName.Name),
			RunnerScaleSetId:   1,
			PodTemplateSpec:    *newRunnerPodTemplateSpec(),
		},
	}
}

func newConfigSecret(namespacedName types.NamespacedName) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-secret-config", namespacedName.Name),
			Namespace: namespacedName.Namespace,
		},
		Data: map[string][]byte{
			"github_token": []byte(defaultGitHubToken),
		},
	}
}

func newRunnerPod(namespacedName types.NamespacedName, status corev1.PodPhase) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
		Spec:   newRunnerPodTemplateSpec().Spec,
		Status: corev1.PodStatus{Phase: status},
	}
}

func newWorkflowPod(namespacedName types.NamespacedName) *corev1.Pod {
	labels := make(map[string]string)
	labels["runner-pod"] = strings.TrimSuffix(namespacedName.Name, "-workflow")

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
			Labels:    labels,
		},
		Spec:   newRunnerPodTemplateSpec().Spec,
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
}

func newRunnerPodTemplateSpec() *corev1.PodTemplateSpec {
	return &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    actionsgithubcom.EphemeralRunnerContainerName,
					Image:   runnerImage,
					Command: []string{"/runner/run.sh"},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "runner",
							MountPath: "/runner",
						},
					},
				},
			},
			InitContainers: []corev1.Container{
				{
					Name:    "setup",
					Image:   runnerImage,
					Command: []string{"sh", "-c", "cp -r /home/runner/* /runner/"},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "runner",
							MountPath: "/runner",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "runner",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}
}

func getWorkflowPod(k8sClient client.Client, namespacedName types.NamespacedName) (*corev1.Pod, error) {
	pod := new(corev1.Pod)
	err := k8sClient.Get(
		context.Background(),
		namespacedName,
		pod,
	)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func deleteWorkflowPod(k8sClient client.Client, namespacedName types.NamespacedName) error {
	workflowPod, _ := getWorkflowPod(k8sClient, namespacedName)
	return k8sClient.Delete(context.Background(), workflowPod)
}
