package node

import (
	"fmt"
	"testing"
)

func TestDescribeDisk(t *testing.T) {
	got, err := DescribeDisk("127.0.0.1")
	if err != nil {
		t.Errorf("DescribeDisk() error = %v", err)
		return
	}
	fmt.Println(len(got))
}

func TestDescribeMountedPoint(t *testing.T) {
	got, err := DescribeMountedPoint()
	if err != nil {
		t.Errorf("DescribeMountedPoint() error = %v", err)
		return
	}
	fmt.Println(len(got))
}
