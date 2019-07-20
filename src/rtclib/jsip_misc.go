// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP MISC

package rtclib

const (
	Unknown = iota
)

func copySlice(src []interface{}) []interface{} {
	dst := make([]interface{}, 0, len(src))

	for _, v := range src {
		if d, ok := v.(map[string]interface{}); ok {
			dst = append(dst, copyMap(d))
		} else if d, ok := v.([]interface{}); ok {
			dst = append(dst, copySlice(d))
		} else {
			dst = append(dst, v)
		}
	}

	return dst
}

func copyMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{})

	for k, v := range src {
		if d, ok := v.(map[string]interface{}); ok {
			dst[k] = copyMap(d)
		} else if d, ok := v.([]interface{}); ok {
			dst[k] = copySlice(d)
		} else {
			dst[k] = v
		}
	}

	return dst
}

func copyBody(src interface{}) interface{} {
	if s, ok := src.(string); ok {
		return s
	}

	if s, ok := src.(map[string]interface{}); ok {
		return copyMap(s)
	}

	if s, ok := src.([]interface{}); ok {
		return copySlice(s)
	}

	return nil
}

func getJsonInt(j map[string]interface{}, key string) (int, bool) {
	value, ok := j[key].(float64)
	if !ok {
		return 0, false
	}

	return int(value), true
}

func getJsonInt64(j map[string]interface{}, key string) (int64, bool) {
	value, ok := j[key].(float64)
	if !ok {
		return 0, false
	}

	return int64(value), true
}

func getJsonUint(j map[string]interface{}, key string) (uint, bool) {
	value, ok := j[key].(float64)
	if !ok {
		return 0, false
	}

	return uint(value), true
}

func getJsonUint64(j map[string]interface{}, key string) (uint64, bool) {
	value, ok := j[key].(float64)
	if !ok {
		return 0, false
	}

	return uint64(value), true
}
