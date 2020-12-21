package main

import (
	"testing"
)

// TestInList
func TestInList(t *testing.T) {
	vals := []string{"bla-bla", "image: repo/srv:tag", "goo-goo"}
	if !InList("goo-goo", vals) {
		t.Errorf("Fail TestInList\n")
	}
}
