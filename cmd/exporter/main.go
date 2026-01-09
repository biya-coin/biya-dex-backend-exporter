package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/biya-coin/biya-dex-backend-exporter/internal/adapters/explorer"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/adapters/stake"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/adapters/tendermint"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/collectors"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/config"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/metrics"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/server"
)

// 版本信息（可在构建时通过 -ldflags 注入）
var (
	version = "dev"
	commit  = "none"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "", "config file path (.yaml/.yml/.json). optional; if empty uses defaults only")
	flag.Parse()

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load config failed:", err)
		os.Exit(1)
	}

	logger := config.NewLogger(cfg.Log.Level)
	slog.SetDefault(logger)
	logger.Info("exporter starting",
		"version", version,
		"commit", commit,
		"http_listen_addr", cfg.HTTP.ListenAddr,
		"chain_id", cfg.Chain.ChainID,
	)

	reg, m := metrics.New(cfg.Chain.ChainID, version, commit)

	// adapters
	stakeCli := stake.NewClient(cfg.Stake.BaseURL, cfg.Stake.APIKey, cfg.HTTPClient.Timeout)
	tmCli := tendermint.NewClient(cfg.Node.TendermintRPCBaseURL, cfg.HTTPClient.Timeout)
	explorerCli := explorer.NewClient(cfg.Explorer.BaseURL, cfg.Explorer.APIKey, cfg.HTTPClient.Timeout)

	// collectors（按类型分组：node / stake / explorer）
	// 注意：这里仅调整代码结构以便维护；不修改 job 名称与 interval，避免影响指标 source label。
	nodeJobs := []collectors.Job{
		collectors.NewJob("realtime_chain", cfg.ScrapeIntervals.Realtime, collectors.NewRealtimeChainCollector(logger, m, tmCli, cfg.Mock)),
		collectors.NewJob("minute_chain", cfg.ScrapeIntervals.Minute, collectors.NewMinuteChainCollector(logger, m, tmCli, cfg.Mock, cfg.Node.MempoolCapacity)),
	}

	stakeJobs := []collectors.Job{
		collectors.NewJob("realtime_stake", cfg.ScrapeIntervals.Realtime, collectors.NewRealtimeStakeCollector(logger, m, stakeCli)),
	}

	// explorer jobs：当前 explorer/indexer 指标仍以内置 mock 方式由 node collectors 兜底，
	// 后续接入真实 explorer client 后，可在这里新增独立 collector。
	explorerJobs := []collectors.Job{
		collectors.NewJob("realtime_explorer", cfg.ScrapeIntervals.Realtime, collectors.NewRealtimeExplorerCollector(logger, m, explorerCli, cfg.Mock)),
	}

	jobs := make([]collectors.Job, 0, len(nodeJobs)+len(stakeJobs)+len(explorerJobs))
	jobs = append(jobs, nodeJobs...)
	jobs = append(jobs, stakeJobs...)
	jobs = append(jobs, explorerJobs...)

	s := collectors.NewScheduler(logger, m, jobs)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := s.Run(ctx); err != nil {
			logger.Error("scheduler stopped with error", "err", err)
			stop()
		}
	}()

	httpSrv := server.New(cfg.HTTP.ListenAddr, reg, s.Ready)
	if err := httpSrv.Start(ctx); err != nil {
		logger.Error("http server stopped with error", "err", err)
		os.Exit(1)
	}

	// 让 scheduler 有机会优雅退出
	select {
	case <-ctx.Done():
		logger.Info("shutdown requested")
	case <-time.After(100 * time.Millisecond):
	}
}
