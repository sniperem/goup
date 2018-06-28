package util

import (
	"strconv"
)

func ToFloat64(v interface{}) float64 {
	if v == nil {
		return 0.0
	}

	switch v.(type) {
	case float64:
		return v.(float64)
	case string:
		s := v.(string)
		vF, _ := strconv.ParseFloat(s, 64)
		return vF
	default:
		panic("to float64 error.")
	}
}

func ToInt(v interface{}) int {
	if v == nil {
		return 0
	}

	switch v.(type) {
	case string:
		vStr := v.(string)
		vInt, _ := strconv.Atoi(vStr)
		return vInt
	case int:
		return v.(int)
	case float64:
		vF := v.(float64)
		return int(vF)
	default:
		panic("to int error.")
	}
}

func ToInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}

	switch v.(type) {
	case int:
		return int64(v.(int64))
	case float64:
		return int64((v.(float64)))
	case string:
		uV, _ := strconv.ParseInt(v.(string), 10, 64)
		return uV
	default:
		panic("to uint64 error.")
	}
}

func ToUint64(v interface{}) uint64 {
	if v == nil {
		return 0
	}

	switch v.(type) {
	case int:
		return uint64(v.(int))
	case float64:
		return uint64((v.(float64)))
	case string:
		uV, _ := strconv.ParseUint(v.(string), 10, 64)
		return uV
	default:
		panic("to uint64 error.")
	}
}
