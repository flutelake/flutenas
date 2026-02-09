package util

import (
	"strconv"
	"strings"
)

var storageUnit = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}

func FormatStorageSize(size uint64) string {
	if size == 1024 {
		return "1024B"
	}
	unit := 0
	value := size
	for value >= 1024 && unit < len(storageUnit)-1 {
		if value == 1024 && unit > 0 {
			break
		}
		value = value / 1024
		unit = unit + 1
	}
	return strconv.FormatUint(value, 10) + storageUnit[unit]
}

func StringToBool(str string) bool {
	switch str {
	case "true", "1":
		return true
	case "false":
		return false
	default:
		return false
	}
}

func Trim(str string) string {
	return strings.Trim(str, "\r\n\t")
}
