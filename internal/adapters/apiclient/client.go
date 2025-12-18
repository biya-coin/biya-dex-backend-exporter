package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Envelope 为 Biya API 的通用响应包装：
//
//	{
//	  "code": 0,
//	  "message": "success",
//	  "data": { ... }
//	}
//
// 参考文档：https://prv.docs.biya.io/api-reference/introduction
type Envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func New(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  strings.TrimSpace(apiKey),
		http: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) GetJSON(ctx context.Context, path string, q url.Values, out any) error {
	return c.doJSON(ctx, http.MethodGet, path, q, out)
}

func (c *Client) doJSON(ctx context.Context, method, path string, q url.Values, out any) error {
	if c.baseURL == "" {
		return fmt.Errorf("api base url is empty")
	}
	if c.apiKey == "" {
		// 文档约定为必填；避免“悄悄打无鉴权请求”导致排查困难。
		return fmt.Errorf("api key is empty")
	}
	if path == "" || !strings.HasPrefix(path, "/") {
		return fmt.Errorf("invalid api path: %q", path)
	}

	u := c.baseURL + path
	if q != nil && len(q) > 0 {
		u += "?" + q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http %d from %s", resp.StatusCode, u)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var env Envelope
	if err := json.Unmarshal(b, &env); err != nil {
		return err
	}
	if env.Code != 0 {
		// message 由上游返回；这里保留 code，便于排查。
		return fmt.Errorf("api error code=%d message=%q", env.Code, env.Message)
	}
	if out == nil {
		return nil
	}
	if len(env.Data) == 0 {
		return fmt.Errorf("api response data is empty")
	}
	// 兼容性优先：不做 DisallowUnknownFields，避免上游加字段导致 exporter 直接挂掉。
	return json.Unmarshal(env.Data, out)
}
