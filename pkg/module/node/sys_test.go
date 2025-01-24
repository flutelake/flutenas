package node

import (
	"fmt"
	"testing"
)

func TestGetOS(t *testing.T) {
	os, ver := GetOS("127.0.0.1")
	fmt.Println(os)
	fmt.Println(ver)
}

func TestGetHostname(t *testing.T) {
	hn := GetHostname("127.0.0.1")
	fmt.Println(hn)
}

func TestGetKernel(t *testing.T) {
	hn := GetKernelVersion("127.0.0.1")
	fmt.Println(hn)
}

func TestGetArch(t *testing.T) {
	hn := GetArch("127.0.0.1")
	fmt.Println(hn)
}
