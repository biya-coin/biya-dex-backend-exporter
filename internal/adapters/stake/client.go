package stake

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
	// 兼容两种配置写法：
	// - https://prv.stake.biya.io
	// - https://prv.stake.biya.io/stake
	//
	// 由于后续请求路径会拼接 "/stake/..."，这里将末尾的 "/stake" 归一化掉，避免出现 "/stake/stake"。
	baseURL = strings.TrimRight(baseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/stake")
	return &Client{
		api: apiclient.New(baseURL, apiKey, timeout),
	}
}

func (c *Client) GetValidators(ctx context.Context, page, pageSize int) (*GetValidatorsResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}

	q := url.Values{}
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("pageSize", fmt.Sprintf("%d", pageSize))

	var out GetValidatorsResponse
	if err := c.api.GetJSON(ctx, "/stake/validators", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetValidatorsAll(ctx context.Context, pageSize int, maxPages int) ([]Validator, error) {
	if pageSize <= 0 {
		pageSize = 100
	}
	if maxPages <= 0 {
		maxPages = 1000
	}
	var all []Validator
	for page := 1; page <= maxPages; page++ {
		resp, err := c.GetValidators(ctx, page, pageSize)
		if err != nil {
			return nil, err
		}
		all = append(all, resp.Validators...)
		if !resp.Pagination.HasNext {
			return all, nil
		}
	}
	return nil, fmt.Errorf("stake.GetValidatorsAll exceeded maxPages=%d", maxPages)
}

func (c *Client) CheckHealth(ctx context.Context, service string) (json.RawMessage, error) {
	q := url.Values{}
	if strings.TrimSpace(service) != "" {
		q.Set("service", service)
	}
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/stake/health", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetDelegation(ctx context.Context, delegatorAddress, validatorAddress string) (json.RawMessage, error) {
	q := url.Values{}
	q.Set("delegatorAddress", delegatorAddress)
	q.Set("validatorAddress", validatorAddress)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/stake/delegation", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetDelegationReward(ctx context.Context, delegatorAddress, validatorAddress string) (json.RawMessage, error) {
	q := url.Values{}
	q.Set("delegatorAddress", delegatorAddress)
	q.Set("validatorAddress", validatorAddress)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/stake/delegation/reward", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetDelegationTotalRewards(ctx context.Context, delegatorAddress string) (json.RawMessage, error) {
	q := url.Values{}
	q.Set("delegatorAddress", delegatorAddress)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/stake/delegation/rewards", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetDelegatorDelegations(ctx context.Context, delegatorAddress string, p NestedPagination) (json.RawMessage, error) {
	q := url.Values{}
	q.Set("delegatorAddress", delegatorAddress)
	addNestedPagination(q, p)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/stake/delegator/delegations", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetDelegatorValidators(ctx context.Context, delegatorAddress string) (json.RawMessage, error) {
	q := url.Values{}
	q.Set("delegatorAddress", delegatorAddress)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/stake/delegator/validators", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetDelegatorWithdrawAddress(ctx context.Context, delegatorAddress string) (json.RawMessage, error) {
	q := url.Values{}
	q.Set("delegatorAddress", delegatorAddress)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/stake/delegator/withdraw/address", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetProposals(ctx context.Context, status int, p NestedPagination) (json.RawMessage, error) {
	q := url.Values{}
	// status 为 enum；0/空值是否有意义由上游定义，这里仅当 >0 才传递，避免误筛选。
	if status > 0 {
		q.Set("status", fmt.Sprintf("%d", status))
	}
	addNestedPagination(q, p)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/stake/governance/proposals", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetProposalByID(ctx context.Context, proposalID string) (json.RawMessage, error) {
	q := url.Values{}
	q.Set("proposalId", proposalID)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/stake/governance/proposals/by-id", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetValidator(ctx context.Context, operatorAddress string) (json.RawMessage, error) {
	q := url.Values{}
	q.Set("operatorAddress", operatorAddress)
	var out json.RawMessage
	if err := c.api.GetJSON(ctx, "/stake/validator", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func addCursorPage(q url.Values, p CursorPage) {
	if p.Page > 0 {
		q.Set("page", fmt.Sprintf("%d", p.Page))
	}
	if p.PageSize > 0 {
		q.Set("pageSize", fmt.Sprintf("%d", p.PageSize))
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
