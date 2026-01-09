package explorer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/biya-coin/biya-dex-backend-exporter/internal/adapters/apiclient"
)

type Client struct {
	api *apiclient.Client
}

// CursorPage 为文档中常见的分页参数组合（page/pageSize/cursor）。
type CursorPage struct {
	Page     int
	PageSize int
	Cursor   string
}

// NestedPagination 为文档中另一类分页参数组合（pagination.page/pagination.pageSize/pagination.cursor）。
type NestedPagination struct {
	Page     int
	PageSize int
	Cursor   string
}

func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	// 移除末尾的斜杠，确保 baseURL 格式正确
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		api: apiclient.New(baseURL, apiKey, timeout),
	}
}

func (c *Client) CheckHealth(ctx context.Context, service string) (json.RawMessage, error) {
	q := url.Values{}
	if strings.TrimSpace(service) != "" {
		q.Set("service", service)
	}
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/api/v1/health", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetAccountBalances(ctx context.Context, address string) (json.RawMessage, error) {
	q := url.Values{}
	q.Set("address", address)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/api/v1/account/balances", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetAccountInfo(ctx context.Context, address string) (json.RawMessage, error) {
	q := url.Values{}
	q.Set("address", address)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/api/v1/account/info", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetAccountTransactions(ctx context.Context, address string, p NestedPagination) (json.RawMessage, error) {
	q := url.Values{}
	q.Set("address", address)
	addNestedPagination(q, p)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/api/v1/account/transactions", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetAccountTransactionsAllPages(ctx context.Context, address string, p NestedPagination, maxPages int) ([]json.RawMessage, error) {
	if maxPages <= 0 {
		maxPages = 1000
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	var pages []json.RawMessage
	for i := 0; i < maxPages; i++ {
		pageData, err := c.GetAccountTransactions(ctx, address, p)
		if err != nil {
			return nil, err
		}
		pages = append(pages, pageData)

		stop, nextCursor := paginationStopAndNextCursor(pageData)
		if stop {
			return pages, nil
		}
		// 优先 cursor，其次 page++（兼容两种分页模式）
		if strings.TrimSpace(nextCursor) != "" {
			p.Cursor = nextCursor
		} else {
			p.Page++
		}
	}
	return nil, fmt.Errorf("explorer.GetAccountTransactionsAllPages exceeded maxPages=%d", maxPages)
}

func (c *Client) GetBlockByHeight(ctx context.Context, height string) (json.RawMessage, error) {
	q := url.Values{}
	q.Set("height", height)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/api/v1/block/by-height", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetLatestBlockHeight(ctx context.Context) (json.RawMessage, error) {
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/api/v1/block/latest-height", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetLatestBlocks(ctx context.Context, p CursorPage) (json.RawMessage, error) {
	q := url.Values{}
	addCursorPage(q, p)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/api/v1/block/latest", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetLatestTransactions(ctx context.Context, p CursorPage) (json.RawMessage, error) {
	q := url.Values{}
	addCursorPage(q, p)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/api/v1/transaction/latest", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetTransactionByHash(ctx context.Context, hash string) (json.RawMessage, error) {
	q := url.Values{}
	q.Set("hash", hash)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/api/v1/transaction/by-hash", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetTransactionStats(ctx context.Context) (json.RawMessage, error) {
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/api/v1/transaction/stats", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetBlockGasUtilization(ctx context.Context) (json.RawMessage, error) {
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/api/v1/block/gas-utilization", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetFailedTransactions24H(ctx context.Context, p NestedPagination) (json.RawMessage, error) {
	q := url.Values{}
	addNestedPagination(q, p)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/api/v1/transaction/failed-24h", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func addCursorPage(q url.Values, p CursorPage) {
	if p.Page > 0 {
		q.Set("page", fmt.Sprintf("%d", p.Page))
	}
	if p.PageSize > 0 {
		q.Set("page_size", fmt.Sprintf("%d", p.PageSize))
	}
	if strings.TrimSpace(p.Cursor) != "" {
		q.Set("cursor", p.Cursor)
	}
}

func addNestedPagination(q url.Values, p NestedPagination) {
	if p.Page > 0 {
		q.Set("pagination.page", fmt.Sprintf("%d", p.Page))
	}
	if p.PageSize > 0 {
		q.Set("pagination.pageSize", fmt.Sprintf("%d", p.PageSize))
	}
	if strings.TrimSpace(p.Cursor) != "" {
		q.Set("pagination.cursor", p.Cursor)
	}
}

// paginationStopAndNextCursor 尝试从 data 中解析分页信息。
// 由于当前阶段主要目标是“把请求打通”，因此这里采用启发式解析：
// - 优先识别 data.pagination.hasNext / totalPages
// - 若存在 data.pagination.cursor，则返回为 nextCursor
func paginationStopAndNextCursor(data json.RawMessage) (stop bool, nextCursor string) {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return false, ""
	}
	pg, ok := m["pagination"].(map[string]any)
	if !ok || pg == nil {
		return false, ""
	}
	// hasNext: bool
	if v, ok := pg["hasNext"].(bool); ok && !v {
		return true, ""
	}
	// totalPages: number, page: number
	if tp, ok := pg["totalPages"].(float64); ok && tp > 0 {
		if p, ok := pg["page"].(float64); ok && p >= tp {
			return true, ""
		}
	}
	// cursor: string
	if c, ok := pg["cursor"].(string); ok && strings.TrimSpace(c) != "" {
		return false, c
	}
	return false, ""
}
