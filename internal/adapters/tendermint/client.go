package tendermint

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Status(ctx context.Context) (*StatusResponse, error) {
	var out StatusResponse
	if err := c.getJSON(ctx, "/status", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Block(ctx context.Context, height int64) (*BlockResponse, error) {
	q := url.Values{}
	if height > 0 {
		q.Set("height", strconv.FormatInt(height, 10))
	}
	var out BlockResponse
	if err := c.getJSON(ctx, "/block", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) NumUnconfirmedTxs(ctx context.Context) (*NumUnconfirmedTxsResponse, error) {
	var out NumUnconfirmedTxsResponse
	if err := c.getJSON(ctx, "/num_unconfirmed_txs", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) getJSON(ctx context.Context, path string, q url.Values, out any) error {
	if c.baseURL == "" {
		return fmt.Errorf("tendermint rpc base url is empty")
	}
	u := c.baseURL + path
	if q != nil && len(q) > 0 {
		u += "?" + q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

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
	// 兼容性优先：不做 DisallowUnknownFields，避免上游加字段导致 exporter 直接挂掉。
	if err := json.Unmarshal(b, out); err != nil {
		return err
	}
	return nil
}
