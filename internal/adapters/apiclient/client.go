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
	// 允许 apiKey 为空：有些环境/接口可能不强制鉴权；若上游需要鉴权则会返回 401/403，由调用方通过 source_up 体现。
	if strings.TrimSpace(c.apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

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

	// 兼容两类响应：
	// 1) 标准 envelope：{"code":..., "message":..., "data":...}
	// 2) 直接返回 data（无 envelope），例如 stake 某些接口：{"validators":[...], ...}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(b, &top); err != nil {
		return err
	}

	// 无 envelope：直接把 body 当成 data
	if _, hasCode := top["code"]; !hasCode {
		if out == nil {
			return nil
		}
		return json.Unmarshal(b, out)
	}

	// 有 envelope：按 envelope 规则解包 data
	var env Envelope
	if err := json.Unmarshal(b, &env); err != nil {
		return err
	}
	// 不同环境返回的 code 口径可能不同：
	// - 部分接口使用 0 表示成功
	// - 部分接口使用 200 表示成功（同时 http status 也是 200）
	// 这里兼容两种口径，避免误判导致指标全部为 0。
	if env.Code != 0 && env.Code != 200 {
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
