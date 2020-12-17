package main

import (
	"strings"
	"testing"
)

// TestChangeTag
func TestChangeTag(t *testing.T) {
	vals := []string{"bla-bla", "image: repo/srv:tag", "goo-goo"}
	r := Request{Name: "srv", Namespace: "test", Tag: "123", Repo: "repo"}
	res := changeTag(strings.Join(vals, "\n"), r)
	if !strings.Contains(res, "123") {
		t.Errorf("Fail TestChangeTag, %s\n", res)
	}
}
