package util

import "testing"

func TestFormatStorageSize(t *testing.T) {
	if got := FormatStorageSize(1024); got != "1024B" {
		t.Errorf("FormatStorageSize() = %v, want %v", got, "1024B")
	}
	if got := FormatStorageSize(1024 * 1024); got != "1024KiB" {
		t.Errorf("FormatStorageSize() = %v, want %v", got, "1024KiB")
	}
	if got := FormatStorageSize(1024 * 1024 * 512); got != "512MiB" {
		t.Errorf("FormatStorageSize() = %v, want %v", got, "512MiB")
	}
}
