package main

import (
	"testing"
)

// TestGetCommit
func TestGetCommit(t *testing.T) {
	r := Request{Service: "imagebot", Namespace: "test", Tag: "00.00.01", Repository: "vkuznet/imagebot"}
	sha, err := getCommit(r)
	if err != nil {
		t.Errorf("Fail TestGetCommit, %v\n", err)
	}
	if sha != "1130ebd31beeb1a0a4b50e908896e918f2b9be7d" {
		t.Errorf("Fail to get sha, %s\n", sha)
	}
}
