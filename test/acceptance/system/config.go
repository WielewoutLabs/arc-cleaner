package system

import (
	"context"
	"errors"
	"testing"

	helmclient "github.com/mittwald/go-helm-client"
	kubeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	testingT   *testing.T
	BaseURL    string
	kubeConfig []byte
	KubeClient kubeclient.Client
	HelmClient helmclient.Client
	LoadImages func(ctx context.Context, images ...string) error
}

func NewConfig(t *testing.T) Config {
	return Config{
		testingT: t,
		LoadImages: func(_ context.Context, _ ...string) error {
			return errors.New("LoadImages is not implemented")
		},
	}
}
