package metrics

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// 这是一个“离线可编译”的最小 Prometheus text exposition 实现：
// - 仅支持 gauge 与 histogram（足够覆盖 MVP）
// - 仅用于 exporter 自身输出 /metrics
// 后续如果你们恢复可用 Go Proxy，可再切回官方 prometheus/client_golang。

type Type string

const (
	TypeGauge     Type = "gauge"
	TypeCounter   Type = "counter"
	TypeHistogram Type = "histogram"
)

type Registry struct {
	mu sync.RWMutex

	help map[string]string
	typ  map[string]Type

	// metric -> seriesKey -> value
	gauges map[string]map[string]float64

	// metric -> seriesKey -> histogram state
	histograms map[string]map[string]*histState

	// 记录 label key 的顺序，保证输出稳定
	labelKeys map[string][]string
}

type histState struct {
	buckets []float64
	counts  []uint64
	sum     float64
	count   uint64
}

func NewRegistry() *Registry {
	return &Registry{
		help:       make(map[string]string),
		typ:        make(map[string]Type),
		gauges:     make(map[string]map[string]float64),
		histograms: make(map[string]map[string]*histState),
		labelKeys:  make(map[string][]string),
	}
}

func (r *Registry) MustDeclare(metric string, t Type, help string, labelKeys []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.typ[metric]; ok {
		return
	}
	r.typ[metric] = t
	r.help[metric] = help
	r.labelKeys[metric] = append([]string(nil), labelKeys...)
	if t == TypeGauge || t == TypeCounter {
		r.gauges[metric] = make(map[string]float64)
	}
	if t == TypeHistogram {
		r.histograms[metric] = make(map[string]*histState)
	}
}

func (r *Registry) SetGauge(metric string, labels map[string]string, v float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	seriesKey := r.seriesKeyLocked(metric, labels)
	if _, ok := r.gauges[metric]; !ok {
		r.gauges[metric] = make(map[string]float64)
	}
	r.gauges[metric][seriesKey] = v
}

func (r *Registry) ObserveHistogram(metric string, labels map[string]string, buckets []float64, v float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	seriesKey := r.seriesKeyLocked(metric, labels)
	m := r.histograms[metric]
	if m == nil {
		m = make(map[string]*histState)
		r.histograms[metric] = m
	}
	h := m[seriesKey]
	if h == nil {
		h = &histState{buckets: append([]float64(nil), buckets...), counts: make([]uint64, len(buckets))}
		m[seriesKey] = h
	}

	// update
	h.count++
	h.sum += v
	for i, b := range h.buckets {
		if v <= b {
			h.counts[i]++
		}
	}
}

func (r *Registry) RenderText() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := make([]string, 0, len(r.typ))
	for k := range r.typ {
		metrics = append(metrics, k)
	}
	sort.Strings(metrics)

	var buf bytes.Buffer
	for _, metric := range metrics {
		help := r.help[metric]
		t := r.typ[metric]
		if help != "" {
			buf.WriteString("# HELP ")
			buf.WriteString(metric)
			buf.WriteString(" ")
			buf.WriteString(escapeHelp(help))
			buf.WriteString("\n")
		}
		buf.WriteString("# TYPE ")
		buf.WriteString(metric)
		buf.WriteString(" ")
		buf.WriteString(string(t))
		buf.WriteString("\n")

		switch t {
		case TypeGauge:
			fallthrough
		case TypeCounter:
			series := r.gauges[metric]
			keys := sortedMapKeys(series)
			for _, sk := range keys {
				buf.WriteString(metric)
				if sk != "" {
					buf.WriteString("{")
					buf.WriteString(sk)
					buf.WriteString("}")
				}
				buf.WriteString(" ")
				buf.WriteString(formatFloat(series[sk]))
				buf.WriteString("\n")
			}
		case TypeHistogram:
			series := r.histograms[metric]
			keys := sortedMapKeys(series)
			for _, sk := range keys {
				h := series[sk]
				// bucket/count lines use suffix _bucket
				var cumulative uint64
				for i, b := range h.buckets {
					cumulative += h.counts[i]
					buf.WriteString(metric)
					buf.WriteString("_bucket")
					buf.WriteString("{")
					if sk != "" {
						buf.WriteString(sk)
						buf.WriteString(",")
					}
					buf.WriteString(`le="`)
					buf.WriteString(formatFloat(b))
					buf.WriteString(`"}`)
					buf.WriteString(" ")
					buf.WriteString(strconv.FormatUint(cumulative, 10))
					buf.WriteString("\n")
				}
				// +Inf bucket
				buf.WriteString(metric)
				buf.WriteString("_bucket")
				buf.WriteString("{")
				if sk != "" {
					buf.WriteString(sk)
					buf.WriteString(",")
				}
				buf.WriteString(`le="+Inf"}`)
				buf.WriteString(" ")
				buf.WriteString(strconv.FormatUint(h.count, 10))
				buf.WriteString("\n")

				// sum
				buf.WriteString(metric)
				buf.WriteString("_sum")
				if sk != "" {
					buf.WriteString("{")
					buf.WriteString(sk)
					buf.WriteString("}")
				}
				buf.WriteString(" ")
				buf.WriteString(formatFloat(h.sum))
				buf.WriteString("\n")

				// count
				buf.WriteString(metric)
				buf.WriteString("_count")
				if sk != "" {
					buf.WriteString("{")
					buf.WriteString(sk)
					buf.WriteString("}")
				}
				buf.WriteString(" ")
				buf.WriteString(strconv.FormatUint(h.count, 10))
				buf.WriteString("\n")
			}
		}
	}
	return buf.String()
}

func (r *Registry) seriesKeyLocked(metric string, labels map[string]string) string {
	if labels == nil || len(labels) == 0 {
		return ""
	}
	keys := r.labelKeys[metric]
	if len(keys) == 0 {
		// fallback stable ordering
		keys = make([]string, 0, len(labels))
		for k := range labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)
	}
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		if v, ok := labels[k]; ok {
			parts = append(parts, fmt.Sprintf(`%s=%q`, k, v))
		}
	}
	return strings.Join(parts, ",")
}

func escapeHelp(s string) string {
	// Prometheus HELP 允许任意 UTF-8；这里只做最小替换
	return strings.ReplaceAll(s, "\n", " ")
}

func sortedMapKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
