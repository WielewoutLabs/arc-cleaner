package system

import (
	"context"

	helmclient "github.com/mittwald/go-helm-client"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	"k8s.io/client-go/tools/clientcmd"
	kubeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	k3sImage = "docker.io/rancher/k3s:v1.31.0-k3s1"
)

func startKubernetesClusterInContainer(c *Config) {
	c.testingT.Helper()

	ctx := context.Background()

	k3sContainer, err := k3s.Run(ctx, k3sImage)
	require.NoError(c.testingT, err)

	c.testingT.Cleanup(func() {
		require.NoError(c.testingT, k3sContainer.Terminate(ctx))
	})

	kubeConfig, err := k3sContainer.GetKubeConfig(ctx)
	require.NoError(c.testingT, err)

	c.kubeConfig = kubeConfig
	c.LoadImages = k3sContainer.LoadImages

	setKubeClient(c)
	setHelmClient(c)
}

func setKubeClient(c *Config) {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(c.kubeConfig))
	require.NoError(c.testingT, err)

	kubeClient, err := kubeclient.New(config, kubeclient.Options{})
	require.NoError(c.testingT, err)

	c.KubeClient = kubeClient
}

func setHelmClient(c *Config) {
	opt := &helmclient.KubeConfClientOptions{
		Options: &helmclient.Options{
			Debug: true,
			DebugLog: func(format string, v ...interface{}) {
				c.testingT.Logf(format, v...)
			},
		},
		KubeConfig: c.kubeConfig,
	}

	helmClient, err := helmclient.NewClientFromKubeConf(opt)
	require.NoError(c.testingT, err)

	c.HelmClient = helmClient
}
