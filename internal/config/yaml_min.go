package config

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type frame struct {
	indent int
	key    string
}

// unmarshalYAMLMinimal 是一个“离线可编译”的最小 YAML 解析器：
// - 仅支持缩进式 map（key: value / key: 作为父级）
// - 仅支持 string/bool/number/duration（duration 支持 5s/1m/1h）
// - 不支持数组、anchor、复杂类型
//
// 目的：当前环境无法拉取 gopkg.in/yaml.v3，先保证联调流程不被阻塞。
func unmarshalYAMLMinimal(b []byte, cfg *Config) error {
	// path -> setter
	setters := map[string]func(string) error{
		"chain.chain_id": func(v string) error { cfg.Chain.ChainID = v; return nil },

		"node.tendermint_rpc_base_url": func(v string) error { cfg.Node.TendermintRPCBaseURL = v; return nil },
		"node.lcd_base_url":            func(v string) error { cfg.Node.LCDBaseURL = v; return nil },
		"node.mempool_capacity": func(v string) error {
			n, err := strconv.Atoi(strings.TrimSpace(v))
			if err != nil {
				return fmt.Errorf("node.mempool_capacity: %w", err)
			}
			cfg.Node.MempoolCapacity = n
			return nil
		},

		"explorer.base_url": func(v string) error { cfg.Explorer.BaseURL = v; return nil },
		"explorer.api_key":  func(v string) error { cfg.Explorer.APIKey = v; return nil },

		"stake.base_url": func(v string) error { cfg.Stake.BaseURL = v; return nil },
		"stake.api_key":  func(v string) error { cfg.Stake.APIKey = v; return nil },

		"http.listen_addr": func(v string) error { cfg.HTTP.ListenAddr = v; return nil },
		"log.level":        func(v string) error { cfg.Log.Level = v; return nil },

		"http_client.timeout": func(v string) error {
			d, err := parseDurationOrNanos(v)
			if err != nil {
				return err
			}
			cfg.HTTPClient.Timeout = d
			return nil
		},

		"scrape_intervals.realtime": func(v string) error {
			d, err := parseDurationOrNanos(v)
			if err != nil {
				return err
			}
			cfg.ScrapeIntervals.Realtime = d
			return nil
		},
		"scrape_intervals.minute": func(v string) error {
			d, err := parseDurationOrNanos(v)
			if err != nil {
				return err
			}
			cfg.ScrapeIntervals.Minute = d
			return nil
		},
		"scrape_intervals.hourly": func(v string) error {
			d, err := parseDurationOrNanos(v)
			if err != nil {
				return err
			}
			cfg.ScrapeIntervals.Hourly = d
			return nil
		},

		"mock.enabled": func(v string) error {
			bv, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("mock.enabled: %w", err)
			}
			cfg.Mock.Enabled = bv
			return nil
		},

		"mock.values.gas_utilization_ratio": func(v string) error {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return fmt.Errorf("mock.values.gas_utilization_ratio: %w", err)
			}
			cfg.Mock.Values.GasUtilizationRatio = f
			return nil
		},
		"mock.values.congestion_ratio": func(v string) error {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return fmt.Errorf("mock.values.congestion_ratio: %w", err)
			}
			cfg.Mock.Values.CongestionRatio = f
			return nil
		},
		"mock.values.tps_window": func(v string) error {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return fmt.Errorf("mock.values.tps_window: %w", err)
			}
			cfg.Mock.Values.TPSWindow = f
			return nil
		},
		"mock.values.mempool_pending_txs": func(v string) error {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return fmt.Errorf("mock.values.mempool_pending_txs: %w", err)
			}
			cfg.Mock.Values.MempoolPendingTxs = f
			return nil
		},
		"mock.values.tx_confirm_time_seconds": func(v string) error {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return fmt.Errorf("mock.values.tx_confirm_time_seconds: %w", err)
			}
			cfg.Mock.Values.TxConfirmTimeSeconds = f
			return nil
		},
	}

	var stack []frame
	sc := bufio.NewScanner(bytes.NewReader(b))
	lineNo := 0
	for sc.Scan() {
		lineNo++
		raw := sc.Text()
		line := stripComment(raw)
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := leadingSpaces(line)
		trim := strings.TrimSpace(line)

		// key: value or key:
		colon := strings.IndexByte(trim, ':')
		if colon <= 0 {
			return fmt.Errorf("yaml line %d: invalid (expect key: value): %q", lineNo, raw)
		}
		key := strings.TrimSpace(trim[:colon])
		rest := strings.TrimSpace(trim[colon+1:])

		// pop to parent level
		for len(stack) > 0 && indent <= stack[len(stack)-1].indent {
			stack = stack[:len(stack)-1]
		}

		if rest == "" {
			// start nested map
			stack = append(stack, frame{indent: indent, key: key})
			continue
		}

		val, err := parseYAMLScalar(rest)
		if err != nil {
			return fmt.Errorf("yaml line %d: %w", lineNo, err)
		}

		full := buildPath(stack, key)
		setter := setters[full]
		if setter == nil {
			// 未声明的字段直接忽略，便于未来扩展与兼容
			continue
		}
		if err := setter(val); err != nil {
			return fmt.Errorf("yaml line %d (%s): %w", lineNo, full, err)
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return nil
}

func buildPath(stack []frame, leaf string) string {
	if len(stack) == 0 {
		return leaf
	}
	parts := make([]string, 0, len(stack)+1)
	for _, f := range stack {
		parts = append(parts, f.key)
	}
	parts = append(parts, leaf)
	return strings.Join(parts, ".")
}

func leadingSpaces(s string) int {
	n := 0
	for n < len(s) && s[n] == ' ' {
		n++
	}
	return n
}

func stripComment(s string) string {
	// 最小实现：遇到第一个 # 就截断；如果你们需要 # 作为值的一部分，请用引号包起来。
	if i := strings.IndexByte(s, '#'); i >= 0 {
		return s[:i]
	}
	return s
}

func parseYAMLScalar(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}
	// 双引号：支持转义
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		v, err := strconv.Unquote(s)
		if err != nil {
			return "", err
		}
		return v, nil
	}
	// 单引号：不处理转义，去掉外层
	if strings.HasPrefix(s, `'`) && strings.HasSuffix(s, `'`) && len(s) >= 2 {
		return s[1 : len(s)-1], nil
	}
	return s, nil
}

func parseDurationOrNanos(v string) (time.Duration, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, nil
	}
	// 优先按 5s/1m/1h 解析
	if d, err := time.ParseDuration(v); err == nil {
		return d, nil
	}
	// 兼容旧 JSON 的“纳秒整数”
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid duration: %q (expect e.g. 5s or nanoseconds int)", v)
	}
	return time.Duration(n), nil
}
