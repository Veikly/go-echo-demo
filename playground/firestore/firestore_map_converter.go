// Package fsconv 提供高性能的 Firestore DTO → map[string]interface{} 转换器。
//
// 设计要点：
//   - 首次对某 struct 类型调用时，通过反射解析 `firestore` tag 并将字段元数据
//     缓存到 sync.Map（全局唯一，并发安全）。
//   - 后续调用直接按缓存的字段 Index 读取 reflect.Value，无需再次反射类型信息。
//   - 支持嵌套 struct、指针、slice、map、time.Time 等常见 Firestore 类型。
//   - 支持 `omitempty`：零值字段自动跳过。
package main

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

// ----------------------------------------
// 元数据结构
// ----------------------------------------

// fieldMeta 存储单个字段的缓存信息，仅在首次解析时填充。
type fieldMeta struct {
	index     []int  // reflect.FieldByIndex 路径（支持嵌套）
	key       string // Firestore 字段名（来自 tag 或字段名小写）
	omitempty bool
	isTime    bool // 快捷标记，避免类型断言
}

// structMeta 缓存一个 struct 类型的全部字段元数据。
type structMeta struct {
	fields []fieldMeta
}

// typeRegistry 是全局反射缓存，key 为 reflect.Type。
var typeRegistry sync.Map // map[reflect.Type]*structMeta

// ----------------------------------------
// 元数据解析（仅首次）
// ----------------------------------------

var timeType = reflect.TypeOf(time.Time{})

// getOrBuildMeta 从缓存获取或（首次）构建指定类型的元数据。
func getOrBuildMeta(t reflect.Type) (*structMeta, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("fsconv: 仅支持 struct 类型，得到 %s", t.Kind())
	}

	if v, ok := typeRegistry.Load(t); ok {
		return v.(*structMeta), nil
	}

	meta, err := buildMeta(t, nil)
	if err != nil {
		return nil, err
	}

	// LoadOrStore 处理并发竞争：若已有其他 goroutine 先存入，直接使用已有值。
	actual, _ := typeRegistry.LoadOrStore(t, meta)
	return actual.(*structMeta), nil
}

// buildMeta 递归解析 struct 字段，收集 fieldMeta。
// indexPrefix 用于内联嵌套 struct（anonymous embed）。
func buildMeta(t reflect.Type, indexPrefix []int) (*structMeta, error) {
	meta := &structMeta{}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		// 跳过未导出字段
		if !f.IsExported() {
			continue
		}

		tag := f.Tag.Get("firestore")
		if tag == "-" {
			continue
		}

		var key string
		var omitempty bool

		if tag != "" {
			parts := strings.Split(tag, ",")
			key = parts[0]
			for _, opt := range parts[1:] {
				if opt == "omitempty" {
					omitempty = true
				}
			}
		}
		if key == "" {
			key = strings.ToLower(f.Name)
		}

		// 构建字段索引路径
		idx := make([]int, len(indexPrefix)+1)
		copy(idx, indexPrefix)
		idx[len(indexPrefix)] = i

		ft := f.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		// 内联匿名 struct（不带 tag）
		if f.Anonymous && ft.Kind() == reflect.Struct && tag == "" {
			nested, err := buildMeta(ft, idx)
			if err != nil {
				return nil, err
			}
			meta.fields = append(meta.fields, nested.fields...)
			continue
		}

		meta.fields = append(meta.fields, fieldMeta{
			index:     idx,
			key:       key,
			omitempty: omitempty,
			isTime:    ft == timeType,
		})
	}

	return meta, nil
}

// ----------------------------------------
// 主转换函数
// ----------------------------------------

// ToMap 将任意 Firestore DTO struct（或指向它的指针）转换为 map[string]interface{}。
//
//	task := PlayTask{Title: "hello", Priority: 1}
//	m, err := fsconv.ToMap(task)
func ToMap(v interface{}) (map[string]interface{}, error) {
	if v == nil {
		return nil, fmt.Errorf("fsconv: 输入为 nil")
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, fmt.Errorf("fsconv: 输入指针为 nil")
		}
		rv = rv.Elem()
	}

	meta, err := getOrBuildMeta(rv.Type())
	if err != nil {
		return nil, err
	}

	return buildMap(rv, meta)
}

// buildMap 按已缓存的 fieldMeta 遍历字段，构建结果 map。
func buildMap(rv reflect.Value, meta *structMeta) (map[string]interface{}, error) {
	result := make(map[string]interface{}, len(meta.fields))

	for _, f := range meta.fields {
		fv := rv.FieldByIndex(f.index)

		// 处理指针：取实际值或跳过 nil
		if fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				if !f.omitempty {
					result[f.key] = nil
				}
				continue
			}
			fv = fv.Elem()
		}

		// omitempty：跳过零值
		if f.omitempty && fv.IsZero() {
			continue
		}

		val, err := convertValue(fv, f.isTime)
		if err != nil {
			return nil, fmt.Errorf("fsconv: 字段 %q 转换失败: %w", f.key, err)
		}
		result[f.key] = val
	}

	return result, nil
}

// convertValue 将单个 reflect.Value 转换为 interface{}。
// 对嵌套 struct 递归调用 ToMap，其余类型直接 Interface()。
func convertValue(fv reflect.Value, isTime bool) (interface{}, error) {
	// time.Time 直接返回，Firestore SDK 原生支持
	if isTime {
		return fv.Interface(), nil
	}

	switch fv.Kind() {
	case reflect.Struct:
		// 非 time.Time 的 struct → 递归转换
		return ToMap(fv.Interface())

	case reflect.Slice:
		if fv.IsNil() {
			return nil, nil
		}
		result := make([]interface{}, fv.Len())
		for i := 0; i < fv.Len(); i++ {
			elem := fv.Index(i)
			if elem.Kind() == reflect.Struct && elem.Type() != timeType {
				m, err := ToMap(elem.Interface())
				if err != nil {
					return nil, err
				}
				result[i] = m
			} else {
				result[i] = elem.Interface()
			}
		}
		return result, nil

	case reflect.Map:
		if fv.IsNil() {
			return nil, nil
		}
		result := make(map[string]interface{}, fv.Len())
		for _, k := range fv.MapKeys() {
			mv := fv.MapIndex(k)
			var err error
			result[fmt.Sprint(k.Interface())], err = convertValue(mv, mv.Type() == timeType)
			if err != nil {
				return nil, err
			}
		}
		return result, nil

	default:
		return fv.Interface(), nil
	}
}

// ----------------------------------------
// 批量转换（减少函数调用开销）
// ----------------------------------------

// ToMaps 批量转换 DTO 切片，共享同一份 structMeta，比逐条调用 ToMap 更高效。
func ToMaps(slice interface{}) ([]map[string]interface{}, error) {
	sv := reflect.ValueOf(slice)
	if sv.Kind() != reflect.Slice {
		return nil, fmt.Errorf("fsconv: ToMaps 需要 slice 类型，得到 %s", sv.Kind())
	}
	if sv.Len() == 0 {
		return []map[string]interface{}{}, nil
	}

	// 取第一个元素的类型解析元数据（只做一次）
	elemType := sv.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	meta, err := getOrBuildMeta(elemType)
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, sv.Len())
	for i := 0; i < sv.Len(); i++ {
		elem := sv.Index(i)
		if elem.Kind() == reflect.Ptr {
			if elem.IsNil() {
				results[i] = nil
				continue
			}
			elem = elem.Elem()
		}
		m, err := buildMap(elem, meta)
		if err != nil {
			return nil, fmt.Errorf("fsconv: 索引 %d 转换失败: %w", i, err)
		}
		results[i] = m
	}
	return results, nil
}

// ----------------------------------------
// 缓存管理
// ----------------------------------------

// ClearCache 清空反射缓存（测试/热更新场景使用）。
func ClearCache() {
	typeRegistry.Range(func(k, _ interface{}) bool {
		typeRegistry.Delete(k)
		return true
	})
}
