package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetCall provides test of GET method for our service
func TestGetCall(t *testing.T) {
	// test GET request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(RequestHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

// TestPostCall provides test of GET method for our service
func TestPostCall(t *testing.T) {
	// test POST request
	r := Request{Service: "srv", Namespace: "test", Tag: "123", Repository: "repo/srv"}
	token, err := genToken(r)
	if err != nil {
		t.Errorf("Fail in genToken, error %v\n", err)
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Errorf("Fail json.Marshal, error %v\n", err)
	}
	req, err := http.NewRequest("POST", "/", bytes.NewBufferString(string(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(RequestHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}

}
