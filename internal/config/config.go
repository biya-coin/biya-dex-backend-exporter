package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Chain    ChainConfig    `json:"chain"`
	Node     NodeConfig     `json:"node"`
	Explorer ExplorerConfig `json:"explorer"`
	Stake    StakeConfig    `json:"stake"`
	HTTP     HTTPConfig     `json:"http"`
	Log      LogConfig      `json:"log"`

	ScrapeIntervals ScrapeIntervalsConfig `json:"scrape_intervals"`
	HTTPClient      HTTPClientConfig      `json:"http_client"`
	Mock            MockConfig            `json:"mock"`
}

type ChainConfig struct {
	ChainID string `json:"chain_id"`
}

type NodeConfig struct {
	// Tendermint/CometBFT RPC，例如：https://rpc.xxx:26657
	TendermintRPCBaseURL string `json:"tendermint_rpc_base_url"`
	// Cosmos LCD REST，例如：https://api.xxx:1317
	LCDBaseURL string `json:"lcd_base_url"`
	// Mempool 容量（pending tx 上限）。若未配置，默认 5000（见 provide.md）
	MempoolCapacity int `json:"mempool_capacity"`
}

type ExplorerConfig struct {
	// Explorer/Indexer API base url，例如：https://prv.explorer.biya.io/demo
	// 注意：当前版本仍可能使用 mock 值兜底；该字段用于后续接入真实 explorer 数据源。
	BaseURL string `json:"base_url"`
	// Bearer token（即文档中的 API Key，不要带 "Bearer " 前缀）
	APIKey string `json:"api_key"`
}

type StakeConfig struct {
	// Stake API base url，例如：https://prv.stake.biya.io 或 https://prv.stake.biya.io/stake
	// 说明：内部会做兼容归一化，避免出现 /stake/stake 的重复路径。
	BaseURL string `json:"base_url"`
	// Bearer token（即文档中的 API Key，不要带 "Bearer " 前缀）
	APIKey string `json:"api_key"`
}

type HTTPConfig struct {
	ListenAddr string `json:"listen_addr"`
}

type LogConfig struct {
	// debug|info|warn|error
	Level string `json:"level"`
}

type HTTPClientConfig struct {
	Timeout time.Duration `json:"timeout"`
}

type ScrapeIntervalsConfig struct {
	Realtime time.Duration `json:"realtime"`
	Minute   time.Duration `json:"minute"`
	Hourly   time.Duration `json:"hourly"`
}

type MockConfig struct {
	Enabled bool `json:"enabled"`
	Values  struct {
		GasUtilizationRatio  float64 `json:"gas_utilization_ratio"`
		CongestionRatio      float64 `json:"congestion_ratio"`
		TPSWindow            float64 `json:"tps_window"`
		MempoolPendingTxs    float64 `json:"mempool_pending_txs"`
		TxConfirmTimeSeconds float64 `json:"tx_confirm_time_seconds"`
	} `json:"values"`
}

func Default() Config {
	var c Config
	c.Chain.ChainID = "biya"
	c.Node.TendermintRPCBaseURL = ""
	c.Node.LCDBaseURL = ""
	c.Node.MempoolCapacity = 5000
	c.Explorer.BaseURL = "https://prv.explorer.biya.io/demo"
	c.Explorer.APIKey = ""
	c.Stake.BaseURL = "https://prv.stake.biya.io/stake"
	c.Stake.APIKey = ""
	c.HTTP.ListenAddr = ":9100"
	c.Log.Level = "info"
	c.HTTPClient.Timeout = 5 * time.Second
	c.ScrapeIntervals.Realtime = 10 * time.Second
	c.ScrapeIntervals.Minute = 1 * time.Minute
	c.ScrapeIntervals.Hourly = 1 * time.Hour
	c.Mock.Enabled = true
	c.Mock.Values.GasUtilizationRatio = 0
	c.Mock.Values.CongestionRatio = 0
	c.Mock.Values.TPSWindow = 0
	c.Mock.Values.MempoolPendingTxs = 0
	c.Mock.Values.TxConfirmTimeSeconds = 0
	return c
}

// Load 仅做最小能力：如果提供 configPath 则读取 YAML 覆盖默认值。
// 说明：后续若需要支持 env 覆盖/多环境，可再扩展，但 MVP 先保证联调可跑通。
func Load(configPath string) (Config, error) {
	cfg := Default()
	if configPath == "" {
		return cfg, nil
	}
	b, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	ext := strings.ToLower(filepath.Ext(configPath))
	switch ext {
	case ".json":
		if err := json.Unmarshal(b, &cfg); err != nil {
			return Config{}, fmt.Errorf("unmarshal json: %w", err)
		}
	case ".yaml", ".yml":
		if err := unmarshalYAMLMinimal(b, &cfg); err != nil {
			return Config{}, fmt.Errorf("unmarshal yaml: %w", err)
		}
	default:
		return Config{}, fmt.Errorf("unsupported config extension: %s (supported: .yaml/.yml/.json)", ext)
	}
	if cfg.Chain.ChainID == "" {
		return Config{}, errors.New("chain.chain_id is required")
	}
	return cfg, nil
}

func NewLogger(level string) *slog.Logger {
	lvl := slog.LevelInfo
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(h)
}
