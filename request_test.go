package main

import (
	"fmt"
	"strings"
	"testing"
)

// TestChangeTag
func TestChangeTag(t *testing.T) {
	vals := []string{"bla-bla", "image: repo/srv:tag", "goo-goo"}
	r := Request{Name: "srv", Namespace: "test", Tag: "123", Repository: "repo"}
	fmt.Println("input", strings.Join(vals, "\n"))
	res := changeTag(strings.Join(vals, "\n"), r)
	fmt.Println("output", res)
	if !strings.Contains(res, "123") {
		t.Errorf("Fail TestChangeTag, %s\n", res)
	}
}
