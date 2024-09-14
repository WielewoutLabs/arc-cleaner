package dsl

import (
	"net/http"
	"time"

	"github.com/wielewout/arc-cleaner/test/acceptance/driver"
	"github.com/wielewout/arc-cleaner/test/acceptance/system"
)

type Dsl struct {
	Health Health
}

func NewDsl(systemConfig system.Config) Dsl {
	httpClient := &http.Client{
		Timeout: time.Second,
	}
	healthDriver := driver.NewHealth(systemConfig, httpClient)

	return Dsl{
		Health: healthDriver,
	}
}
