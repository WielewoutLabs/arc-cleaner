package system

import (
	"testing"
)

func SetUp(t *testing.T) Config {
	c := NewConfig(t)
	startKubernetesClusterInContainer(&c)
	startWithHelmChart(&c)
	return c
}
