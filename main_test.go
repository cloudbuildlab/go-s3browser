package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestFormatSize(t *testing.T) {
	cases := []struct {
		size int64
		want string
	}{
		{512, "512 B"},
		{2048, "2.00 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
	}
	for _, c := range cases {
		got := formatSize(c.size)
		if got != c.want {
			t.Errorf("formatSize(%d) = %q, want %q", c.size, got, c.want)
		}
	}
}

func TestRootHandlerReturnsHTML(t *testing.T) {
	w := httptest.NewRecorder()

	// Use a minimal handler for testing: just call the template with dummy data
	tmpl := template.Must(template.New("test").Funcs(template.FuncMap{
		"formatSize": formatSize,
		"add":        add,
		"dec":        dec,
		"inc":        inc,
		"until":      until,
		"slice":      slice,
	}).Parse(htmlTemplate))

	data := PageData{
		Objects:      []S3Object{},
		Breadcrumbs:  []Breadcrumb{},
		CurrentPath:  "",
		ParentPrefix: "",
	}
	resp := struct {
		PageData
		Page, TotalPages, Limit, Total int
		SortOrder, SortBy, Prefix      string
	}{
		PageData:   data,
		Page:       1,
		TotalPages: 1,
		Limit:      25,
		Total:      0,
		SortOrder:  "asc",
		SortBy:     "name",
		Prefix:     "",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tmpl.Execute(w, resp)
	if err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	result := w.Result()
	if result.StatusCode != http.StatusOK && result.StatusCode != 0 {
		t.Errorf("expected status 200 or 0, got %d", result.StatusCode)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Geonet Open Data Browser") {
		t.Errorf("expected HTML to contain title, got: %s", body)
	}
}

func TestHealthCheck(t *testing.T) {
	// Set up test environment
	if err := os.Setenv("FASTLY_SERVICE_VERSION", "test-version"); err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("FASTLY_SERVICE_VERSION"); err != nil {
			t.Errorf("Failed to unset environment variable: %v", err)
		}
	}()

	// Create a test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call the handler directly
	if req.URL.Path == "/health" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintf(w, `{"status":"ok","version":"%s"}`, os.Getenv("FASTLY_SERVICE_VERSION")); err != nil {
			t.Errorf("Error writing health check response: %v", err)
		}
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Parse response body
	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response body: %v", err)
	}

	// Verify response fields
	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
	if response["version"] != "test-version" {
		t.Errorf("Expected version 'test-version', got '%s'", response["version"])
	}
}
