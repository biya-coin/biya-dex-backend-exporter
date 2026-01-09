package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/biya-coin/biya-dex-backend-exporter/internal/adapters/alertmanager"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/adapters/prometheus"
)

// AlertTrendService 处理告警趋势查询
type AlertTrendService struct {
	prometheusClient   *prometheus.Client
	alertmanagerClient *alertmanager.Client
	logger             *slog.Logger
}

// NewAlertTrendService 创建告警趋势服务
func NewAlertTrendService(prometheusClient *prometheus.Client, alertmanagerClient *alertmanager.Client, logger *slog.Logger) *AlertTrendService {
	return &AlertTrendService{
		prometheusClient:   prometheusClient,
		alertmanagerClient: alertmanagerClient,
		logger:             logger,
	}
}

// AlertTrendPoint 表示趋势图中的一个数据点
type AlertTrendPoint struct {
	Timestamp int64 `json:"timestamp"` // Unix 时间戳（秒）
	Critical  int   `json:"critical"`  // 严重告警数
	Warning   int   `json:"warning"`   // 警告告警数
	Total     int   `json:"total"`     // 总告警数
}

// AlertTrendResponse 告警趋势响应
type AlertTrendResponse struct {
	Success bool              `json:"success"`
	Data    []AlertTrendPoint `json:"data"`
	Message string            `json:"message,omitempty"`
}

// GetAlertTrend 获取最近 N 天的告警趋势
func (s *AlertTrendService) GetAlertTrend(ctx context.Context, days int) (*AlertTrendResponse, error) {
	if days <= 0 || days > 30 {
		days = 7 // 默认7天
	}

	end := time.Now()
	start := end.AddDate(0, 0, -days)
	// 使用1小时作为步长，减少数据点数量
	step := 1 * time.Hour

	// 查询 Prometheus 的 ALERTS 指标
	// ALERTS{alertstate="firing"} 表示正在触发的告警
	// 使用 sum_over_time 来统计每个时间点的告警数量
	criticalQuery := `sum(ALERTS{alertstate="firing",severity="critical"})`
	warningQuery := `sum(ALERTS{alertstate="firing",severity="warning"})`
	totalQuery := `count(ALERTS{alertstate="firing"})`

	// 查询严重告警
	criticalResult, err := s.prometheusClient.QueryRange(ctx, criticalQuery, start, end, step)
	if err != nil {
		s.logger.Warn("failed to query critical alerts", "error", err)
		// 如果查询失败，尝试使用 Alertmanager API 作为备选方案
		return s.getTrendFromAlertmanager(ctx, days)
	}

	// 查询警告告警
	warningResult, err := s.prometheusClient.QueryRange(ctx, warningQuery, start, end, step)
	if err != nil {
		s.logger.Warn("failed to query warning alerts", "error", err)
		return s.getTrendFromAlertmanager(ctx, days)
	}

	// 查询总告警数
	totalResult, err := s.prometheusClient.QueryRange(ctx, totalQuery, start, end, step)
	if err != nil {
		s.logger.Warn("failed to query total alerts", "error", err)
		return s.getTrendFromAlertmanager(ctx, days)
	}

	// 合并数据点
	points := s.mergeQueryResults(criticalResult, warningResult, totalResult, step)

	return &AlertTrendResponse{
		Success: true,
		Data:    points,
	}, nil
}

// mergeQueryResults 合并多个查询结果
func (s *AlertTrendService) mergeQueryResults(critical, warning, total *prometheus.QueryResult, step time.Duration) []AlertTrendPoint {
	// 创建一个时间戳到数据点的映射
	pointMap := make(map[int64]*AlertTrendPoint)

	// 辅助函数：解析 Prometheus 返回的值
	parseValue := func(v interface{}) (float64, bool) {
		switch val := v.(type) {
		case string:
			f, err := strconv.ParseFloat(val, 64)
			return f, err == nil
		case float64:
			return val, true
		case int:
			return float64(val), true
		default:
			return 0, false
		}
	}

	// 处理严重告警数据
	for _, series := range critical.Data.Result {
		for _, value := range series.Values {
			if len(value) < 2 {
				continue
			}
			ts, ok := value[0].(float64)
			if !ok {
				continue
			}
			count, ok := parseValue(value[1])
			if !ok {
				continue
			}

			timestamp := int64(ts)
			if pointMap[timestamp] == nil {
				pointMap[timestamp] = &AlertTrendPoint{Timestamp: timestamp}
			}
			pointMap[timestamp].Critical += int(count)
		}
	}

	// 处理警告告警数据
	for _, series := range warning.Data.Result {
		for _, value := range series.Values {
			if len(value) < 2 {
				continue
			}
			ts, ok := value[0].(float64)
			if !ok {
				continue
			}
			count, ok := parseValue(value[1])
			if !ok {
				continue
			}

			timestamp := int64(ts)
			if pointMap[timestamp] == nil {
				pointMap[timestamp] = &AlertTrendPoint{Timestamp: timestamp}
			}
			pointMap[timestamp].Warning += int(count)
		}
	}

	// 处理总告警数（如果 total 查询有数据，使用它；否则使用 critical + warning）
	for _, series := range total.Data.Result {
		for _, value := range series.Values {
			if len(value) < 2 {
				continue
			}
			ts, ok := value[0].(float64)
			if !ok {
				continue
			}
			count, ok := parseValue(value[1])
			if !ok {
				continue
			}

			timestamp := int64(ts)
			if pointMap[timestamp] == nil {
				pointMap[timestamp] = &AlertTrendPoint{Timestamp: timestamp}
			}
			// 如果 total 查询有值，使用它；否则保持 critical + warning 的和
			if pointMap[timestamp].Total == 0 {
				pointMap[timestamp].Total = int(count)
			}
		}
	}

	// 对于没有 total 数据的时间点，使用 critical + warning 的和
	for timestamp, point := range pointMap {
		if point.Total == 0 {
			point.Total = point.Critical + point.Warning
		}
		pointMap[timestamp] = point
	}

	// 转换为切片并排序
	points := make([]AlertTrendPoint, 0, len(pointMap))
	for _, point := range pointMap {
		points = append(points, *point)
	}

	// 按时间戳排序（使用简单的冒泡排序，数据点不会太多）
	for i := 0; i < len(points)-1; i++ {
		for j := i + 1; j < len(points); j++ {
			if points[i].Timestamp > points[j].Timestamp {
				points[i], points[j] = points[j], points[i]
			}
		}
	}

	return points
}

// getTrendFromAlertmanager 从 Alertmanager 获取趋势（备选方案，仅返回当前状态）
func (s *AlertTrendService) getTrendFromAlertmanager(ctx context.Context, days int) (*AlertTrendResponse, error) {
	alerts, err := s.alertmanagerClient.GetAlerts(ctx, true, false, false)
	if err != nil {
		return &AlertTrendResponse{
			Success: false,
			Message: fmt.Sprintf("failed to query alerts: %v", err),
		}, nil
	}

	// 统计当前告警
	critical := 0
	warning := 0
	for _, alert := range alerts {
		severity := alert.Labels["severity"]
		if severity == "critical" {
			critical++
		} else if severity == "warning" {
			warning++
		}
	}

	// 由于 Alertmanager 不存储历史，我们只能返回当前时间点的数据
	now := time.Now().Unix()
	points := []AlertTrendPoint{
		{
			Timestamp: now,
			Critical:  critical,
			Warning:   warning,
			Total:     len(alerts),
		},
	}

	return &AlertTrendResponse{
		Success: true,
		Data:    points,
		Message: "Note: Alertmanager does not store historical data. Only current alerts are shown.",
	}, nil
}

// HandleAlertTrend 处理告警趋势查询请求
func (s *AlertTrendService) HandleAlertTrend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取查询参数
	daysStr := r.URL.Query().Get("days")
	days := 7 // 默认7天
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 30 {
			days = d
		}
	}

	// 查询趋势数据
	response, err := s.GetAlertTrend(r.Context(), days)
	if err != nil {
		s.logger.Error("failed to get alert trend", "error", err)
		response = &AlertTrendResponse{
			Success: false,
			Message: fmt.Sprintf("Internal error: %v", err),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if !response.Success {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(response)
}
