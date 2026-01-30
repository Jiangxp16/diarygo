package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func Trim(s string) string {
	return strings.TrimSpace(string([]rune(s)))
}

func YMD2Int(y, m, d int) int {
	return y*10000 + m*100 + d
}

func Date2Int(t time.Time) int {
	year := t.Year()
	month := int(t.Month())
	day := t.Day()
	return YMD2Int(year, month, day)
}

func GetCurrentDateInt() int {
	return Date2Int(time.Now())
}

func GetCurrentYYYYMM() int {
	date := time.Now()
	return date.Year()*100 + int(date.Month())
}

func GetNowYYMMDD() string {
	now := time.Now()
	return now.Format("060102")
}

func GetTypeName[T any]() string {
	var t T
	typ := reflect.TypeOf(t)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	name := typ.Name()
	return strings.ToLower(name)
}

func MakePlaceholders(n int) []string {
	s := make([]string, n)
	for i := range s {
		s[i] = "?"
	}
	return s
}

func NormalizeID(v any) (int, error) {
	switch x := v.(type) {
	case float64:
		return int(x), nil
	case int:
		return x, nil
	case int64:
		return int(x), nil
	case string:
		return strconv.Atoi(x)
	default:
		return 0, fmt.Errorf("invalid id type %T", v)
	}
}

func ConvertByStruct[T any](params map[string]any) (map[string]any, error) {
	var zero T
	typ := reflect.TypeOf(zero)

	out := make(map[string]any, len(params))

	for key, val := range params {
		field, ok := findStructField(typ, key)
		if !ok {
			continue // 不存在的字段直接忽略
		}

		converted, err := convertValue(val, field.Type)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", key, err)
		}

		out[key] = converted
	}

	return out, nil
}

func findStructField(t reflect.Type, name string) (reflect.StructField, bool) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if strings.EqualFold(f.Name, name) {
			return f, true
		}

		if tag := f.Tag.Get("json"); tag != "" {
			if strings.Split(tag, ",")[0] == name {
				return f, true
			}
		}
	}
	return reflect.StructField{}, false
}

func convertValue(v any, t reflect.Type) (any, error) {
	// 先处理 nil
	if v == nil {
		return nil, nil
	}

	// 如果本来就能直接用
	rv := reflect.ValueOf(v)
	if rv.Type().AssignableTo(t) {
		return v, nil
	}

	s := fmt.Sprint(v)

	switch t.Kind() {
	case reflect.String:
		return s, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		return reflect.ValueOf(i).Convert(t).Interface(), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
		return reflect.ValueOf(u).Convert(t).Interface(), nil

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
		return reflect.ValueOf(f).Convert(t).Interface(), nil

	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return nil, err
		}
		return b, nil
	}

	return nil, fmt.Errorf("unsupported type %s", t)
}
