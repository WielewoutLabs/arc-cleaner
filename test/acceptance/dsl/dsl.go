package dsl

import (
	"net/http"
	"time"

	"github.com/wielewoutlabs/arc-cleaner/test/acceptance/driver"
	"github.com/wielewoutlabs/arc-cleaner/test/acceptance/system"
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
