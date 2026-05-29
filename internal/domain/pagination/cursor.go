package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-echo-demo/internal/constants"
	"strconv"
	"time"
)

const cursorVersion = 1

// cursorPayload 是游标的序列化结构。
// SortValue 统一用 string 存储，解码时根据 SortValueType 还原为正确类型。
type cursorPayload struct {
	DocID          string `json:"id"`
	SortField      string `json:"sf"`
	SortValueRaw   string `json:"sv"`   // 序列化后的排序值
	SortValueType  string `json:"svt"`  // 值类型标识："time" | "string" | "int64" | "float64"
	Version        int    `json:"ver"`
}

// CursorData 游标携带的位置信息。
type CursorData struct {
	DocID     string // Firestore 文档 ID，用于 tiebreaker
	SortField string // 排序字段名
	SortValue any    // 排序字段的值，用于 Firestore StartAfter（已还原为正确类型）
}

// EncodeCursor 将游标位置信息编码为 opaque 字符串（base64 JSON）。
// 外部只感知字符串，内部结构对前端透明。
// todo 对 cursor 进行签名防篡改
func EncodeCursor(data CursorData) (string, error) {
	raw, typ, err := encodeSortValue(data.SortValue)
	if err != nil {
		return "", err
	}
	p := cursorPayload{
		DocID:         data.DocID,
		SortField:     data.SortField,
		SortValueRaw:  raw,
		SortValueType: typ,
		Version:       cursorVersion,
	}
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// DecodeCursor 解码游标字符串，返回 CursorData。
func DecodeCursor(cursor string) (CursorData, error) {
	if cursor == "" {
		return CursorData{}, nil
	}
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return CursorData{}, constants.InvalidCursor
	}
	var p cursorPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return CursorData{}, constants.InvalidCursor
	}
	if p.Version != cursorVersion {
		return CursorData{}, constants.InvalidCursor
	}
	sortValue, err := decodeSortValue(p.SortValueRaw, p.SortValueType)
	if err != nil {
		return CursorData{}, constants.InvalidCursor
	}
	return CursorData{
		DocID:     p.DocID,
		SortField: p.SortField,
		SortValue: sortValue,
	}, nil
}

// encodeSortValue 将排序值序列化为 string，并返回类型标识。
func encodeSortValue(v any) (raw string, typ string, err error) {
	switch val := v.(type) {
	case time.Time:
		return strconv.FormatInt(val.UnixNano(), 10), "time", nil
	case string:
		return val, "string", nil
	case int64:
		return strconv.FormatInt(val, 10), "int64", nil
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), "float64", nil
	case nil:
		return "", "nil", nil
	default:
		return "", "", fmt.Errorf("unsupported sort value type: %T", v)
	}
}

// decodeSortValue 根据类型标识将 string 还原为正确类型。
func decodeSortValue(raw, typ string) (any, error) {
	switch typ {
	case "time":
		ns, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, err
		}
		return time.Unix(0, ns).UTC(), nil
	case "string":
		return raw, nil
	case "int64":
		return strconv.ParseInt(raw, 10, 64)
	case "float64":
		return strconv.ParseFloat(raw, 64)
	case "nil", "":
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown sort value type: %s", typ)
	}
}
