package cmd

import (
	"context"
	"log/slog"
	"os"

	"github.com/actions/actions-runner-controller/apis/actions.github.com/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
Actions Runner Controller (ARC).`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("started arc-cleaner", "version", version, "commit", commit)

		k8sClient, err := kubernetes.NewClient()
		if err != nil {
			slog.Error("failed to create kubernetes client", "error", err.Error())
		}

		namespace := viper.GetString("namespace")

		ephemeralRunnerSetList := new(v1alpha1.EphemeralRunnerSetList)
		if err := k8sClient.List(context.Background(), ephemeralRunnerSetList, client.InNamespace(namespace)); err != nil {
			slog.Error("failed to list ephemeral runner sets", "namespace", namespace, "error", err.Error())
		}

		slog.Debug("listed ephemeral runner set", "namespace", namespace, "length", len(ephemeralRunnerSetList.Items))
		for index, ephemeralRunnerSet := range ephemeralRunnerSetList.Items {
			slog.Debug("ephemeral runner set", "index", index, "name", ephemeralRunnerSet.Name)
		}
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
