package util

import (
	"fmt"
	"strings"
)

var storageUnit = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}

func FormatStorageSize(size uint64) string {
	unit := 0
	sizeFloat := float64(size)
	for sizeFloat > 1024 {
		sizeFloat = sizeFloat / 1024
		unit = unit + 1
	}
	return fmt.Sprintf("%.2f%s", sizeFloat, storageUnit[unit])
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
