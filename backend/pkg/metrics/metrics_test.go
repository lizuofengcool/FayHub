package metrics

import (
	"testing"
	"time"
)

func TestRecordRequest(t *testing.T) {
	RecordRequest("GET", "/api/test", 100*time.Millisecond, 200)
	RecordRequest("POST", "/api/test", 50*time.Millisecond, 201)
	RecordRequest("GET", "/api/test", 200*time.Millisecond, 500)

	m := GetMetrics()

	totalReqs, ok := m["total_requests"].(int64)
	if !ok || totalReqs < 3 {
		t.Errorf("Expected total_requests >= 3, got %v", m["total_requests"])
	}

	errorReqs, ok := m["error_requests"].(int64)
	if !ok || errorReqs < 1 {
		t.Errorf("Expected error_requests >= 1, got %v", m["error_requests"])
	}
}

func TestActiveRequests(t *testing.T) {
	IncrementActiveRequests()
	IncrementActiveRequests()

	m := GetMetrics()
	active, ok := m["active_requests"].(int64)
	if !ok || active < 2 {
		t.Errorf("Expected active_requests >= 2, got %v", m["active_requests"])
	}

	DecrementActiveRequests()
}

func TestGetPrometheusFormat(t *testing.T) {
	RecordRequest("GET", "/api/prom", 10*time.Millisecond, 200)

	result := GetPrometheusFormat()
	if len(result) == 0 {
		t.Fatal("GetPrometheusFormat returned empty string")
	}

	if !containsStr(result, "fayhub_requests_total") {
		t.Error("Prometheus output missing fayhub_requests_total")
	}

	if !containsStr(result, "fayhub_goroutines") {
		t.Error("Prometheus output missing fayhub_goroutines")
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
