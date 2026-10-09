package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestResponse(t *testing.T) {
	h := New(Config{AppName: "response", Message: "hello"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "test-request")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), "hello") {
		t.Fatalf("unexpected response: %d %s", res.Code, res.Body.String())
	}
}
func TestControlledError(t *testing.T) {
	h := New(Config{})
	res := httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/error", nil))
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", res.Code)
	}
}

func TestControlledErrorLogsStatusAndSeverity(t *testing.T) {
	var logs bytes.Buffer
	var errorLogs bytes.Buffer
	requestLogger.SetOutput(&logs)
	errorLogger.SetOutput(&errorLogs)
	t.Cleanup(func() {
		requestLogger.SetOutput(os.Stdout)
		errorLogger.SetOutput(os.Stderr)
	})

	h := New(Config{})
	res := httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/error", nil))

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(logs.Bytes()), &entry); err != nil {
		t.Fatalf("request log is not JSON: %v", err)
	}
	if entry["severity"] != "ERROR" {
		t.Fatalf("severity = %v", entry["severity"])
	}
	if entry["status_code"] != float64(http.StatusInternalServerError) {
		t.Fatalf("status_code = %v", entry["status_code"])
	}

	var errorEntry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(errorLogs.Bytes()), &errorEntry); err != nil {
		t.Fatalf("error report is not JSON: %v", err)
	}
	if errorEntry["severity"] != "ERROR" {
		t.Fatalf("error severity = %v", errorEntry["severity"])
	}
	if !strings.Contains(errorEntry["message"].(string), "goroutine") {
		t.Fatalf("error report does not contain a Go stack trace")
	}
	serviceContext := errorEntry["serviceContext"].(map[string]any)
	if _, ok := serviceContext["service"]; !ok {
		t.Fatalf("error report has no service context")
	}
}
func TestDelayRejectsRange(t *testing.T) {
	h := New(Config{})
	res := httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/delay?ms=-1", nil))
	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", res.Code)
	}
}
