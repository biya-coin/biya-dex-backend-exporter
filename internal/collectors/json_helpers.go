package collectors

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

// jsonUnmarshal 为 collector 侧提供一个“尽量不丢精度”的 JSON 反序列化：
// - 对 interface{} / any 使用 UseNumber，避免大整数被转成 float64 后精度丢失
// - 对上游偶发的“数字字符串”场景，配合 toFloat64 做兼容解析
func jsonUnmarshal(raw json.RawMessage, out any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	return dec.Decode(out)
}

// toFloat64 将常见的 JSON 标量（number / string-number）转为 float64。
func toFloat64(v any) (float64, bool) {
	switch x := v.(type) {
	case nil:
		return 0, false
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case uint64:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(s, 64)
		return f, err == nil
	default:
		return 0, false
	}
}


