package tools

import "strconv"

func UInt32ToString(value uint32) string {
	return strconv.FormatUint(uint64(value), 10)
}

func UInt64ToString(value uint64) string {
	return strconv.FormatUint(value, 10)
}

func Int64ToString(value int64) string {
	return strconv.FormatInt(value, 10)
}

func Int32ToString(value int32) string {
	return strconv.FormatInt(int64(value), 10)
}

func IntToString(value int) string {
	return strconv.FormatInt(int64(value), 10)
}
