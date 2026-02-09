package node

import (
	"fmt"
	"os"
	"testing"
)

func TestExec_Command1(t *testing.T) {
	if os.Getenv("FLUTENAS_ENABLE_SSH_TESTS") != "1" {
		t.Skip("ssh tests disabled")
	}
	client := NewExec()
	client.SetHost("10.0.1.10")
	if err := client.Connect(); err != nil {
		t.Error(err)
	}
	defer client.Close()
	bs, err := client.Command("ls -l")
	if err != nil {
		t.Error(err)
	}
	fmt.Print(string(bs))
}

func TestExec_Command2(t *testing.T) {
	if os.Getenv("FLUTENAS_ENABLE_SSH_TESTS") != "1" {
		t.Skip("ssh tests disabled")
	}
	client := NewExec()
	client.SetHost("127.0.0.1")
	if err := client.Connect(); err != nil {
		t.Error(err)
	}
	defer client.Close()
	bs, err := client.Command("ls -l /root")
	if err != nil {
		t.Error(err)
	}
	fmt.Print(string(bs))
}
