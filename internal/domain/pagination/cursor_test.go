package pagination

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"go-echo-demo/internal/constants"
)

// ---------------------------------------------------------------------------
// EncodeCursor / DecodeCursor 往返测试
// ---------------------------------------------------------------------------

func TestCursor_RoundTrip_Time(t *testing.T) {
	original := CursorData{
		DocID:     "abc123",
		SortField: "updated_at",
		SortValue: time.Date(2026, 5, 29, 8, 25, 31, 0, time.UTC),
	}
	encoded, err := EncodeCursor(original)
	if err != nil {
		t.Fatalf("EncodeCursor error: %v", err)
	}
	decoded, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeCursor error: %v", err)
	}
	if decoded.DocID != original.DocID {
		t.Errorf("DocID: got %q, want %q", decoded.DocID, original.DocID)
	}
	if decoded.SortField != original.SortField {
		t.Errorf("SortField: got %q, want %q", decoded.SortField, original.SortField)
	}
	oriTime := original.SortValue.(time.Time)
	decTime := decoded.SortValue.(time.Time)
	if !oriTime.Equal(decTime) {
		t.Errorf("SortValue(time): got %v, want %v", decTime, oriTime)
	}
}

func TestCursor_RoundTrip_String(t *testing.T) {
	original := CursorData{
		DocID:     "doc-str",
		SortField: "title",
		SortValue: "hello world",
	}
	encoded, err := EncodeCursor(original)
	if err != nil {
		t.Fatalf("EncodeCursor error: %v", err)
	}
	decoded, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeCursor error: %v", err)
	}
	if decoded.SortValue.(string) != "hello world" {
		t.Errorf("SortValue(string): got %q, want hello world", decoded.SortValue)
	}
}

func TestCursor_RoundTrip_Int64(t *testing.T) {
	original := CursorData{
		DocID:     "doc-int",
		SortField: "score",
		SortValue: int64(9876543210),
	}
	encoded, err := EncodeCursor(original)
	if err != nil {
		t.Fatalf("EncodeCursor error: %v", err)
	}
	decoded, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeCursor error: %v", err)
	}
	if decoded.SortValue.(int64) != int64(9876543210) {
		t.Errorf("SortValue(int64): got %v, want 9876543210", decoded.SortValue)
	}
}

func TestCursor_RoundTrip_Float64(t *testing.T) {
	original := CursorData{
		DocID:     "doc-float",
		SortField: "rating",
		SortValue: float64(3.14159),
	}
	encoded, err := EncodeCursor(original)
	if err != nil {
		t.Fatalf("EncodeCursor error: %v", err)
	}
	decoded, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeCursor error: %v", err)
	}
	if decoded.SortValue.(float64) != float64(3.14159) {
		t.Errorf("SortValue(float64): got %v, want 3.14159", decoded.SortValue)
	}
}

func TestCursor_RoundTrip_Nil(t *testing.T) {
	original := CursorData{
		DocID:     "doc-nil",
		SortField: "optional_field",
		SortValue: nil,
	}
	encoded, err := EncodeCursor(original)
	if err != nil {
		t.Fatalf("EncodeCursor error: %v", err)
	}
	decoded, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeCursor error: %v", err)
	}
	if decoded.SortValue != nil {
		t.Errorf("SortValue(nil): got %v, want nil", decoded.SortValue)
	}
}

func TestEncodeCursor_UnsupportedType(t *testing.T) {
	type unsupported struct{ x int }
	_, err := EncodeCursor(CursorData{
		DocID:     "doc",
		SortField: "field",
		SortValue: unsupported{x: 1},
	})
	if err == nil {
		t.Fatal("expected error for unsupported SortValue type, got nil")
	}
}

// ---------------------------------------------------------------------------
// DecodeCursor 错误路径
// ---------------------------------------------------------------------------

func TestDecodeCursor_EmptyString(t *testing.T) {
	data, err := DecodeCursor("")
	if err != nil {
		t.Fatalf("DecodeCursor(\"\") should return no error, got: %v", err)
	}
	if data.DocID != "" || data.SortField != "" || data.SortValue != nil {
		t.Errorf("DecodeCursor(\"\") should return zero CursorData, got: %+v", data)
	}
}

func TestDecodeCursor_InvalidBase64(t *testing.T) {
	_, err := DecodeCursor("not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != constants.InvalidCursor {
		t.Errorf("expected constants.InvalidCursor, got %v", err)
	}
}

func TestDecodeCursor_InvalidJSON(t *testing.T) {
	// 合法 base64，但内容不是 JSON
	encoded := base64.URLEncoding.EncodeToString([]byte("this is not json"))
	_, err := DecodeCursor(encoded)
	if err != constants.InvalidCursor {
		t.Errorf("expected constants.InvalidCursor, got %v", err)
	}
}

func TestDecodeCursor_VersionMismatch(t *testing.T) {
	// 手动构造一个 version=99 的合法 JSON payload
	payload := map[string]any{
		"id":  "doc1",
		"sf":  "updated_at",
		"sv":  "0",
		"svt": "int64",
		"ver": 99, // 与 cursorVersion=1 不符
	}
	b, _ := json.Marshal(payload)
	encoded := base64.URLEncoding.EncodeToString(b)

	_, err := DecodeCursor(encoded)
	if err != constants.InvalidCursor {
		t.Errorf("expected constants.InvalidCursor for version mismatch, got %v", err)
	}
}
