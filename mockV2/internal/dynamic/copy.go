package dynamic

import (
	"github.com/bytedance/sonic"
)

// DeepCopy implements high-performance deep copy (avoids modifying original data)
func DeepCopy(src any) any {
	if src == nil {
		return nil
	}
	return deepCopyValue(src)
}

func DeepCopyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}

	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = deepCopyValue(v)
	}
	return dst
}

func DeepCopySlice(src []any) []any {
	if src == nil {
		return nil
	}

	dst := make([]any, len(src))
	for i, v := range src {
		dst[i] = deepCopyValue(v)
	}
	return dst
}

func deepCopyValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return DeepCopyMap(val)
	case []any:
		return DeepCopySlice(val)
	case string, float64, bool:
		return v
	case nil:
		return nil
	case int:
		return val
	case int32:
		return val
	case int64:
		return val
	case float32:
		return val
	default:
		return JsonCopy(val)
	}
}

// JsonCopy uses sonic for high-performance JSON serialization/deserialization fallback
func JsonCopy(val any) any {
	data, _ := sonic.Marshal(val)

	var dst any
	_ = sonic.Unmarshal(data, &dst)

	return dst
}
