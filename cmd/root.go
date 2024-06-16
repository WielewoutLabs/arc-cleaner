package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	githubarcv1alpha1 "github.com/actions/actions-runner-controller/apis/actions.github.com/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/wielewout/arc-cleaner/internal/kubernetes"
	"github.com/wielewout/arc-cleaner/internal/logging"
)

var (
	version string
	commit  string
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "arc-cleaner",
	Short: "A cleaner for GitHub ARC",
	Long: `ARC cleaner is an application to clean up resources from the GitHub
Actions Runner Controller (ARC).

GitHub Actions Runners in kubernetes mode sometimes get stuck as ephemeral
volumes are used. These are tied to the lifetime of the runner pod.
When a runner pod exits or crashes while a workflow pod is still running,
then the runner gets stuck waiting indefinitely for storage.
By cleaning up the workflow pod and thus detaching the volume,
the runner can become available again.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		logger := slog.Default()
		logger.Info("started arc-cleaner", "version", version, "commit", commit)

		k8sClient, err := kubernetes.NewClient()
		if err != nil {
			logger.Error("failed to create kubernetes client", "error", err.Error())
		}

		namespace := viper.GetString("namespace")

		nsLogger := logger.With("namespace", namespace)
		nsCtx := logging.WithContext(ctx, nsLogger)

		ephemeralRunnerList := getEphemeralRunnerList(nsCtx, k8sClient, namespace)
		for _, ephemeralRunner := range ephemeralRunnerList.Items {
			erLogger := nsLogger.With("ephemeral-runner", ephemeralRunner.Name)
			erCtx := logging.WithContext(nsCtx, erLogger)
			cleanEphemeralRunnerPods(erCtx, k8sClient, ephemeralRunner)
		}
	},
}

func getEphemeralRunnerList(ctx context.Context, k8sClient *kubernetes.Client, namespace string) *githubarcv1alpha1.EphemeralRunnerList {
	logger := logging.FromContext(ctx)

	ephemeralRunnerList := new(githubarcv1alpha1.EphemeralRunnerList)
	err := k8sClient.List(
		ctx,
		ephemeralRunnerList,
		client.InNamespace(namespace),
	)

	if err != nil {
		logger.Error("failed to list ephemeral runners", "error", err.Error())
		return ephemeralRunnerList
	}

	logger.Debug("listed ephemeral runners", "length", len(ephemeralRunnerList.Items))
	return ephemeralRunnerList
}

func cleanEphemeralRunnerPods(ctx context.Context, k8sClient *kubernetes.Client, ephemeralRunner githubarcv1alpha1.EphemeralRunner) {
	logger := logging.FromContext(ctx)

	runnerPod, err := getRunnerPod(ctx, k8sClient, ephemeralRunner)
	if err != nil {
		logger.Debug("skipping", "reason", "runner pod does not exist", "error", err.Error())
		return
	}

	runnerPodStatus := getPodStatus(runnerPod)
	logger.Debug(fmt.Sprintf("runner pod status is %s", strings.ToLower(string(runnerPodStatus))))
	if runnerPodStatus != corev1.PodPending {
		logger.Debug("skipping", "reason", "runner pod is not pending")
		return
	}

	workflowPod, err := getWorkflowPod(ctx, k8sClient, ephemeralRunner)
	if err != nil {
		logger.Debug("skipping", "reason", "workflow pod does not exist", "error", err.Error())
		return
	}

	workflowPodStatus := getPodStatus(workflowPod)
	logger.Debug(fmt.Sprintf("workflow pod status is %s", strings.ToLower(string(workflowPodStatus))))

	opts := &client.DeleteOptions{}
	if viper.GetBool("dryrun") {
		opts.DryRun = []string{metav1.DryRunAll}
		logger.Debug("dry run to delete worflow pod")
	}
	err = k8sClient.Delete(ctx, workflowPod, opts)
	if err != nil {
		logger.Error("failed deleting workflow pod", "error", err.Error())
	}
	logger.Info("deleted workflow pod")
}

func getRunnerPod(ctx context.Context, k8sClient *kubernetes.Client, ephemeralRunner githubarcv1alpha1.EphemeralRunner) (*corev1.Pod, error) {
	logger := logging.FromContext(ctx)

	pod := new(corev1.Pod)
	err := k8sClient.Get(
		ctx,
		types.NamespacedName{
			Namespace: ephemeralRunner.Namespace,
			Name:      ephemeralRunner.Name,
		},
		pod,
	)
	if err != nil {
		logger.Debug("unable to get runner pod", "error", err.Error())
		return nil, err
	}

	return pod, nil
}

func getWorkflowPod(ctx context.Context, k8sClient *kubernetes.Client, ephemeralRunner githubarcv1alpha1.EphemeralRunner) (*corev1.Pod, error) {
	logger := logging.FromContext(ctx)

	pod := new(corev1.Pod)
	err := k8sClient.Get(
		ctx,
		types.NamespacedName{
			Namespace: ephemeralRunner.Namespace,
			Name:      fmt.Sprintf("%s-workflow", ephemeralRunner.Name),
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

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.arc-cleaner.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".arc-cleaner")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		level := viper.GetString("log.level")
		logging.SetLevel(level)

		slog.Debug("using config file", "path", viper.ConfigFileUsed())
	}
}
