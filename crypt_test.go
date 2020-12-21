package main

import (
	"testing"
)

// TestCrypt
func TestCrypt(t *testing.T) {
	msg := "test"
	data, _ := encrypt([]byte(msg), "salt")
	res, _ := decrypt(data, "salt")
	if string(res) != msg {
		t.Errorf("Fail TestCrypt, %s!=%s\n", msg, string(res))
	}
}
