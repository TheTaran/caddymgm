package main

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONBodyRejectsOversizedPayload(t *testing.T) {
	request := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"username":"`+strings.Repeat("x", loginJSONBodyLimit)+`"}`))
	response := httptest.NewRecorder()
	var payload map[string]string

	if err := decodeJSONBody(response, request, &payload, loginJSONBodyLimit); err == nil {
		t.Fatal("oversized JSON body was accepted")
	}
}

func TestDecodeJSONBodyRejectsMultipleValues(t *testing.T) {
	request := httptest.NewRequest("POST", "/api/sites", strings.NewReader(`{"mode":"proxy"} {"mode":"static"}`))
	response := httptest.NewRecorder()
	var payload map[string]string

	if err := decodeJSONBody(response, request, &payload, adminJSONBodyLimit); err == nil {
		t.Fatal("multiple JSON values were accepted")
	}
}

func TestDecodeJSONBodyAcceptsWhitespaceAfterDocument(t *testing.T) {
	request := httptest.NewRequest("POST", "/api/sites", bytes.NewBufferString("{\"mode\":\"proxy\"}\n  \t"))
	response := httptest.NewRecorder()
	var payload map[string]string

	if err := decodeJSONBody(response, request, &payload, adminJSONBodyLimit); err != nil {
		t.Fatalf("valid JSON body was rejected: %v", err)
	}
}
