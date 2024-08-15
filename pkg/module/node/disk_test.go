package node

import (
	"fmt"
	"testing"
)

func TestDescribeDisk(t *testing.T) {
	got, err := DescribeDisk()
	if err != nil {
		t.Errorf("DescribeDisk() error = %v", err)
		return
	}
	fmt.Println(len(got))
}
