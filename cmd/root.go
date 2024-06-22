package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/wielewout/arc-cleaner/internal/app"
	"github.com/wielewout/arc-cleaner/internal/kubernetes"
	"github.com/wielewout/arc-cleaner/internal/logging"
	"github.com/wielewout/arc-cleaner/internal/signals"
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
		ctx := signals.NewContextWithSignals()

		logger := slog.Default()
		logger.Info("started arc-cleaner", "version", version, "commit", commit)

		k8sClient, err := kubernetes.NewClient()
		if err != nil {
			logger.Error("failed to create kubernetes client", "error", err.Error())
			return
		}

		appOptions := []app.Option{}
		if viper.IsSet("namespace") {
			appOptions = append(appOptions, app.WithNamespace(viper.GetString("namespace")))
		}
		if viper.IsSet("dryrun") {
			appOptions = append(appOptions, app.WithDryRun(viper.GetBool("dryrun")))
		}

		app := app.New(k8sClient, appOptions...)
		app.Start(ctx)

		logger.Debug("exiting arc-cleaner gracefully")
	},
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
