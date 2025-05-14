package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"
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
