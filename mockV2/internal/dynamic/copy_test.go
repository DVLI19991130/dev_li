package dynamic

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/bytedance/sonic"
)

func BenchmarkDeepCopy_MapStringAny(b *testing.B) {
	data := map[string]any{
		"name":    "test",
		"age":     25.0,
		"active":  true,
		"address": map[string]any{"city": "Beijing", "zip": "100000"},
		"scores":  []any{95.5, 88.0, 92.3},
		"tags":    []any{"go", "性能", "测试"},
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		DeepCopy(data)
	}
}

func BenchmarkDeepCopy_SliceNested(b *testing.B) {
	data := []any{
		map[string]any{"id": 1.0, "name": "item1"},
		map[string]any{"id": 2.0, "name": "item2"},
		map[string]any{"id": 3.0, "name": "item3"},
		[]any{1.0, 2.0, 3.0, 4.0, 5.0},
		"string value",
		123.45,
		true,
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		DeepCopy(data)
	}
}

func BenchmarkDeepCopy_LargeMap(b *testing.B) {
	data := make(map[string]any)
	for i := 0; i < 100; i++ {
		data[string(rune('a'+i%26))+string(rune('a'+(i+1)%26))] = map[string]any{
			"id":      float64(i),
			"name":    "test",
			"value":   123.45,
			"enabled": true,
			"meta":    map[string]any{"key": "value"},
		}
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		DeepCopy(data)
	}
}

func BenchmarkDeepCopy_PrimitiveTypes(b *testing.B) {
	primitives := []any{
		"string",
		123.45,
		true,
		nil,
		int(42),
		int32(42),
		int64(42),
		float32(3.14),
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, p := range primitives {
			DeepCopy(p)
		}
	}
}

func BenchmarkJSONFallback_Sonic(b *testing.B) {
	data := map[string]any{
		"custom": struct{ Name string }{Name: "test"},
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		JsonCopy(data)
	}
}

func BenchmarkJSONFallback_StdLib(b *testing.B) {
	data := map[string]any{
		"custom": struct{ Name string }{Name: "test"},
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		data, _ := json.Marshal(data)
		var dst any
		json.Unmarshal(data, &dst)
	}
}

func BenchmarkDeepCopy_Comparison(b *testing.B) {
	data := map[string]any{
		"name":    "performance test",
		"age":     30.0,
		"isValid": true,
		"address": map[string]any{
			"city":     "Shanghai",
			"district": "Pudong",
			"zip":      "200000",
		},
		"skills": []any{"Go", "Python", "Rust", "C++"},
		"metrics": map[string]any{
			"cpu":    45.5,
			"memory": 1024.0,
			"disk":   2048.0,
		},
	}

	b.Run("DeepCopy (optimized)", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			DeepCopy(data)
		}
	})

	b.Run("StdLib JSON Marshal/Unmarshal", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			bytes, _ := json.Marshal(data)
			var dst any
			json.Unmarshal(bytes, &dst)
		}
	})

	b.Run("Sonic Marshal + StdLib Unmarshal", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			bytes, _ := sonic.Marshal(data)
			var dst any
			json.Unmarshal(bytes, &dst)
		}
	})
}

func TestDeepCopy_DataIntegrity(t *testing.T) {
	original := map[string]any{
		"name":   "test",
		"age":    25.0,
		"active": true,
		"address": map[string]any{
			"city": "Beijing",
			"zip":  "100000",
		},
		"scores": []any{95.5, 88.0, 92.3},
		"nilVal": nil,
		"zero":   0.0,
	}

	copied := DeepCopy(original)

	if copied == nil {
		t.Fatal("DeepCopy returned nil")
	}
	if reflect.ValueOf(copied).Pointer() == reflect.ValueOf(original).Pointer() {
		t.Fatal("DeepCopy did not create a new object")
	}

	copiedMap := copied.(map[string]any)
	origMap := original

	if copiedMap["name"] != origMap["name"] {
		t.Errorf("name mismatch: got %v, want %v", copiedMap["name"], origMap["name"])
	}
	if copiedMap["age"] != origMap["age"] {
		t.Errorf("age mismatch: got %v, want %v", copiedMap["age"], origMap["age"])
	}
	if copiedMap["active"] != origMap["active"] {
		t.Errorf("active mismatch: got %v, want %v", copiedMap["active"], origMap["active"])
	}

	copiedAddr := copiedMap["address"].(map[string]any)
	origAddr := origMap["address"].(map[string]any)
	if reflect.ValueOf(copiedAddr).Pointer() == reflect.ValueOf(origAddr).Pointer() {
		t.Error("nested map was not deep copied")
	}
	if copiedAddr["city"] != origAddr["city"] {
		t.Errorf("nested city mismatch: got %v, want %v", copiedAddr["city"], origAddr["city"])
	}

	copiedScores := copiedMap["scores"].([]any)
	origScores := origMap["scores"].([]any)
	if reflect.ValueOf(copiedScores).Pointer() == reflect.ValueOf(origScores).Pointer() {
		t.Error("slice was not deep copied")
	}
	if copiedScores[0] != origScores[0] {
		t.Errorf("slice value mismatch: got %v, want %v", copiedScores[0], origScores[0])
	}

	if copiedMap["nilVal"] != nil {
		t.Errorf("nil value not preserved: got %v", copiedMap["nilVal"])
	}
}

func TestDeepCopy_PrimitiveTypes(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{"string", "hello"},
		{"float64", 3.14},
		{"bool true", true},
		{"bool false", false},
		{"nil", nil},
		{"int", int(42)},
		{"int32", int32(42)},
		{"int64", int64(42)},
		{"float32", float32(3.14)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copied := DeepCopy(tt.value)
			if copied != tt.value && tt.value != nil {
				t.Errorf("DeepCopy(%v) = %v, want %v", tt.name, copied, tt.value)
			}
		})
	}
}

func TestSonicFallbackCopy(t *testing.T) {
	data := map[string]any{
		"custom": struct{ Name string }{Name: "test"},
	}

	copied := JsonCopy(data)
	if copied == nil {
		t.Fatal("JsonCopy returned nil")
	}

	copiedMap := copied.(map[string]any)
	if copiedMap["custom"] == nil {
		t.Error("custom field should be copied")
	}
}
