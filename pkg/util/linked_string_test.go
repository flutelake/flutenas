package util

import (
	"testing"
)

func TestLinkedRune_LinkedRune(t *testing.T) {
	message := "hello fluteNAS"

	r := NewLinkedRune(message)
	if r.String() != message {
		t.Fail()
	}
}
