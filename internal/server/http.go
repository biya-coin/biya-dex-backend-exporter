package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/biya-coin/biya-dex-backend-exporter/internal/metrics"
)

type Server struct {
	addr              string
	reg               *metrics.Registry
	ready             func() bool
	alertTrendService *AlertTrendService
}

func New(listenAddr string, reg *metrics.Registry, ready func() bool) *Server {
	return &Server{addr: listenAddr, reg: reg, ready: ready}
}

// SetAlertTrendService 设置告警趋势服务
func (s *Server) SetAlertTrendService(service *AlertTrendService) {
	s.alertTrendService = service
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = w.Write([]byte(s.reg.RenderText()))
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if s.ready != nil && !s.ready() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("not ready"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	// 告警趋势查询接口
	if s.alertTrendService != nil {
		mux.HandleFunc("/api/v1/alerts/trend", s.alertTrendService.HandleAlertTrend)
	}

	srv := &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		return nil
	case err := <-errCh:
		return err
	}
}
