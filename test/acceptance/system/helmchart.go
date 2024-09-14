package system

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/api/types/image"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"golang.org/x/exp/rand"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultBinaryName  = "arc-cleaner"
	binaryBuildContext = "../.."

	containerfileName     = "build/container/Containerfile"
	containerBuildContext = "../.."
	defaultContainerImage = "ghcr.io/wielewout/arc-cleaner:acceptance-test"

	releaseName    = "arc-cleaner"
	localChartPath = "../../deploy/chart"
)

func startWithHelmChart(c *Config) {
	c.testingT.Helper()

	ctx := context.Background()

	containerImageName := build(ctx, c)
	installHelmChart(ctx, c, containerImageName)
}

func build(ctx context.Context, c *Config) string {
	containerImageName := containerImageName()
	if !usePrebuiltContainerImage() {
		binaryName := binaryName()
		if !usePrebuiltBinary() {
			binaryName = buildBinary(ctx, c)
		}
		containerImageName = buildImageFromContainerfile(ctx, c, binaryName)
	} else {
		pullImageFromRegistry(ctx, c, containerImageName)
	}

	loadImageInKubernetesCluster(ctx, c, containerImageName)

	return containerImageName
}

func usePrebuiltBinary() bool {
	usePrebuiltBinary := os.Getenv("USE_PREBUILT_BINARY")
	return strings.ToLower(usePrebuiltBinary) == "true"
}

func buildBinary(ctx context.Context, c *Config) string {
	binaryName := fmt.Sprintf("%s-%s", binaryName(), strings.ToLower(randomString(8)))

	build := exec.CommandContext(ctx, "/bin/sh", "-c", fmt.Sprintf(`cd "%s" && make build BINARY_NAME="%s"`, binaryBuildContext, binaryName))

	err := build.Run()
	require.NoError(c.testingT, err)

	return binaryName
}

func usePrebuiltContainerImage() bool {
	usePrebuiltContainerImage := os.Getenv("USE_PREBUILT_CONTAINER_IMAGE")
	return strings.ToLower(usePrebuiltContainerImage) == "true"
}

func buildImageFromContainerfile(ctx context.Context, c *Config, binaryName string) string {
	provider, err := testcontainers.NewDockerProvider()
	require.NoError(c.testingT, err)

	containerImage := fmt.Sprintf("%s-%s", containerImageName(), strings.ToLower(randomString(8)))
	containerRequest := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    containerBuildContext,
			Dockerfile: containerfileName,
			Repo:       containerImageRepo(containerImage),
			Tag:        containerImageTag(containerImage),
			BuildArgs: map[string]*string{
				"BINARY_NAME": &binaryName,
			},
			PrintBuildLog: true,
		},
	}
	containerImageName, err := provider.BuildImage(ctx, &containerRequest)
	require.NoError(c.testingT, err)

	c.testingT.Cleanup(func() {
		_, err := provider.Client().ImageRemove(ctx, containerImageName, image.RemoveOptions{
			Force:         true,
			PruneChildren: true,
		})
		require.NoError(c.testingT, err)
	})

	return containerImageName
}

func pullImageFromRegistry(ctx context.Context, c *Config, containerImageName string) {
	provider, err := testcontainers.NewDockerProvider()
	require.NoError(c.testingT, err)

	err = provider.PullImage(ctx, containerImageName)
	require.NoError(c.testingT, err)
}

func loadImageInKubernetesCluster(ctx context.Context, c *Config, containerImageName string) {
	err := c.LoadImages(ctx, containerImageName)
	require.NoError(c.testingT, err)
}

func binaryName() string {
	containerImage := os.Getenv("BINARY_NAME")
	if containerImage == "" {
		containerImage = defaultBinaryName
	}

	return containerImage
}

func containerImageName() string {
	containerImage := os.Getenv("CONTAINER_IMAGE")
	if containerImage == "" {
		containerImage = defaultContainerImage
	}

	return containerImage
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func containerImageRepo(containerImage string) string {
	lastIndex := strings.LastIndex(containerImage, ":")
	return containerImage[:lastIndex]
}

func containerImageTag(containerImage string) string {
	lastIndex := strings.LastIndex(containerImage, ":")
	return containerImage[lastIndex+1:]
}

func installHelmChart(ctx context.Context, c *Config, containerImageName string) {
	release, err := c.HelmClient.InstallOrUpgradeChart(ctx, chartSpec(containerImageName), &helmclient.GenericHelmOptions{})
	require.NoError(c.testingT, err)

	c.testingT.Cleanup(func() {
		err := c.HelmClient.UninstallReleaseByName(release.Name)
		require.NoError(c.testingT, err)
	})

	podList := new(corev1.PodList)
	err = c.KubeClient.List(
		ctx,
		podList,
		client.MatchingLabels{
			"app.kubernetes.io/name":     "arc-cleaner",
			"app.kubernetes.io/instance": release.Name,
		},
	)
	require.NoError(c.testingT, err)

	require.Equal(c.testingT, 1, len(podList.Items))
	pod := podList.Items[0]

	port := forwardPodPort(c, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, 8080)
	c.BaseURL = fmt.Sprintf("http://localhost:%d", port)
}

func chartSpec(containerImageName string) *helmclient.ChartSpec {
	return &helmclient.ChartSpec{
		ChartName:    localChartPath,
		GenerateName: true,
		NameTemplate: fmt.Sprintf("%s-{{randAlphaNum 6 | lower}}", releaseName),
		Atomic:       true,
		Timeout:      30 * time.Second,
		ValuesOptions: values.Options{
			Values: []string{fmt.Sprintf("image.tag=%s", containerImageTag(containerImageName))},
		},
	}
}

func forwardPodPort(c *Config, namespacedPodName types.NamespacedName, port int) int {
	restKubeConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(c.kubeConfig))
	require.NoError(c.testingT, err)

	hostIP := strings.TrimLeft(restKubeConfig.Host, "htps:/")
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		namespacedPodName.Namespace, namespacedPodName.Name)

	transport, upgrader, err := spdy.RoundTripperFor(restKubeConfig)
	require.NoError(c.testingT, err)

	localPort := freeLocalPort(c)

	stopChannel := make(chan struct{})
	readyChannel := make(chan struct{})

	var errorBuffer, dataBuffer bytes.Buffer
	errorBufferWriter := bufio.NewWriter(&errorBuffer)
	dataBufferWriter := bufio.NewWriter(&dataBuffer)

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", localPort, port)}, stopChannel, readyChannel, dataBufferWriter, errorBufferWriter)
	require.NoError(c.testingT, err)

	c.testingT.Cleanup(func() {
		close(stopChannel)
	})

	go func() {
		err := fw.ForwardPorts()
		require.NoError(c.testingT, err)
	}()

	<-readyChannel

	return localPort
}

func freeLocalPort(c *Config) int {
	address, err := net.ResolveTCPAddr("tcp", "localhost:0")
	require.NoError(c.testingT, err)

	listener, err := net.ListenTCP("tcp", address)
	require.NoError(c.testingT, err)
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}
