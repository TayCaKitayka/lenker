package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrPanelURLRequired        = errors.New("panel url is required")
	ErrNodeTokenRequired       = errors.New("node token is required")
	ErrPendingRevisionAuth     = errors.New("pending config revision auth failed")
	ErrUnexpectedPanelResponse = errors.New("unexpected panel response")
)

type PendingConfigRevisionClient interface {
	FetchPendingConfigRevision(ctx context.Context, nodeID string, nodeToken string) (ConfigRevision, bool, error)
}

type PanelClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func (c PanelClient) FetchPendingConfigRevision(ctx context.Context, nodeID string, nodeToken string) (ConfigRevision, bool, error) {
	if strings.TrimSpace(nodeID) == "" {
		return ConfigRevision{}, false, ErrNodeIDRequired
	}
	if strings.TrimSpace(nodeToken) == "" {
		return ConfigRevision{}, false, ErrNodeTokenRequired
	}

	baseURL := strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	if baseURL == "" {
		return ConfigRevision{}, false, ErrPanelURLRequired
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/nodes/%s/config-revisions/pending", baseURL, url.PathEscape(nodeID)),
		nil,
	)
	if err != nil {
		return ConfigRevision{}, false, err
	}
	request.Header.Set("Authorization", "Bearer "+nodeToken)

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return ConfigRevision{}, false, err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return ConfigRevision{}, false, nil
	case http.StatusUnauthorized:
		return ConfigRevision{}, false, ErrPendingRevisionAuth
	default:
		return ConfigRevision{}, false, fmt.Errorf("%w: status %d", ErrUnexpectedPanelResponse, response.StatusCode)
	}

	var envelope struct {
		Data *ConfigRevision `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		return ConfigRevision{}, false, fmt.Errorf("%w: %v", ErrUnexpectedPanelResponse, err)
	}
	if envelope.Data == nil {
		return ConfigRevision{}, false, ErrUnexpectedPanelResponse
	}

	return *envelope.Data, true, nil
}
