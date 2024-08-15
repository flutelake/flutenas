package util

import "fmt"

var storageUnit = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}

func FormatStorageSize(size uint64) string {
	unit := 0
	for size > 1024 {
		size = size / 1024
		unit = unit + 1
	}
	return fmt.Sprintf("%d%s", size, storageUnit[unit])
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
