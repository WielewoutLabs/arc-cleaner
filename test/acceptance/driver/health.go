package driver

import (
	"fmt"
	"net/http"

	"github.com/wielewoutlabs/arc-cleaner/test/acceptance/system"
)

type Health struct {
	systemConfig system.Config
	client       *http.Client
}

func NewHealth(systemConfig system.Config, client *http.Client) Health {
	return Health{
		systemConfig: systemConfig,
		client:       client,
	}
}

func (h Health) Live() error {
	response, err := h.client.Get(fmt.Sprintf("%s/livez", h.systemConfig.BaseURL))
	if err != nil {
		return err
	}

	if http.StatusOK != response.StatusCode {
		return fmt.Errorf("expected live response with status code %d, but got %d", http.StatusOK, response.StatusCode)
	}

	return nil
}

func (h Health) Ready() error {
	response, err := h.client.Get(fmt.Sprintf("%s/readyz", h.systemConfig.BaseURL))
	if err != nil {
		return err
	}

	if http.StatusOK != response.StatusCode {
		return fmt.Errorf("expected ready response with status code %d, but got %d", http.StatusOK, response.StatusCode)
	}

	return nil
}
