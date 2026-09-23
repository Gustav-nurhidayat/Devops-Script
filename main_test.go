package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	expected := "Hello, DevOps! version=dev"

	if !strings.Contains(rec.Body.String(), expected) {
		t.Fatalf(
			"expected response to contain %q, got %q",
			expected,
			rec.Body.String(),
		)
	}
}
