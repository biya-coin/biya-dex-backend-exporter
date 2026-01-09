package alertmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client 用于查询 Alertmanager API
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient 创建新的 Alertmanager 客户端
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: timeout,
		},
	}
}

// Alert 表示 Alertmanager 中的告警
type Alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      *time.Time        `json:"endsAt,omitempty"`
	Status      struct {
		State       string            `json:"state"` // "active", "suppressed", "unprocessed"
		SilencedBy  []string          `json:"silencedBy"`
		InhibitedBy []string          `json:"inhibitedBy"`
	} `json:"status"`
	Receivers []string `json:"receivers"`
	Fingerprint string `json:"fingerprint"`
}

// GetAlertsResponse 表示获取告警的响应
type GetAlertsResponse []Alert

// GetAlerts 获取所有告警（支持过滤）
func (c *Client) GetAlerts(ctx context.Context, active, silenced, inhibited bool) (GetAlertsResponse, error) {
	u := fmt.Sprintf("%s/api/v2/alerts", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if active {
		q.Set("active", "true")
	}
	if silenced {
		q.Set("silenced", "true")
	}
	if inhibited {
		q.Set("inhibited", "true")
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("alertmanager query failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var alerts GetAlertsResponse
	if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		return nil, err
	}

	return alerts, nil
}
