package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/fastly/compute-sdk-go/fsthttp"
)

// S3Object represents a file or directory in the S3 bucket
type S3Object struct {
	Key          string
	LastModified string
	Size         int64
	IsDirectory  bool
	Name         string
	Href         string
	Type         string // file extension/type
	S3URL        string // direct S3 URL
}

// Breadcrumb represents a navigation path element
type Breadcrumb struct {
	Name string
	Path string
}

// PageData contains all the data needed to render the browser page
type PageData struct {
	Objects      []S3Object
	Breadcrumbs  []Breadcrumb
	CurrentPath  string
	ParentPrefix string
}

const (
	htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Geonet Open Data Browser</title>
    <style>
        :root {
            --bg: #f6f8fa;
            --fg: #222;
            --card: #fff;
            --border: #e1e4e8;
            --primary: #2d7ff9;
            --hover: #f0f4fa;
            --icon: #6a737d;
            --accent: #eaf5ff;
        }
        [data-theme="dark"] {
            --bg: #181a1b;
            --fg: #eaeaea;
            --card: #23272e;
            --border: #30363d;
            --primary: #58a6ff;
            --hover: #23272e;
            --icon: #8b949e;
            --accent: #1a2a3a;
        }
        html, body { background: var(--bg); color: var(--fg); margin: 0; padding: 0; }
        body { font-family: 'Segoe UI', 'Roboto', Arial, sans-serif; min-height: 100vh; }
        .container { max-width: 900px; margin: 2rem auto; background: var(--card); border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.04); padding: 2rem 2.5vw; border: 1px solid var(--border); }
        h1 { font-size: 1.7rem; margin-bottom: 1.2rem; letter-spacing: -1px; }
        .breadcrumb { margin-bottom: 18px; font-size: 1.05em; }
        .breadcrumb a { color: var(--primary); text-decoration: none; }
        .breadcrumb a:hover { text-decoration: underline; }
        .breadcrumb .current { font-weight: bold; color: var(--fg); }
        .theme-toggle { float: right; margin-top: -2.5rem; margin-bottom: 1rem; }
        .theme-toggle button { background: var(--accent); color: var(--primary); border: none; border-radius: 6px; padding: 0.4em 1em; font-size: 1em; cursor: pointer; transition: background 0.2s; }
        .theme-toggle button:hover { background: var(--primary); color: #fff; }
        .controls { display: flex; align-items: center; gap: 2em; margin-bottom: 1em; }
        .sort-toggle, .limit-toggle { font-size: 1em; }
        .sort-toggle a, .limit-toggle a { color: var(--primary); text-decoration: none; margin-right: 0.5em; }
        .sort-toggle a.active, .limit-toggle a.active { font-weight: bold; text-decoration: underline; }
        .pagination { margin: 1.5em 0 1em 0; text-align: center; }
        .pagination a { color: var(--primary); text-decoration: none; margin: 0 0.3em; padding: 0.2em 0.7em; border-radius: 5px; }
        .pagination a.active { background: var(--primary); color: #fff; font-weight: bold; }
        .pagination a:hover { background: var(--accent); }
        table { width: 100%; border-collapse: separate; border-spacing: 0; background: var(--card); border-radius: 10px; overflow: hidden; }
        th, td { padding: 12px 10px; text-align: left; border-bottom: 1px solid var(--border); }
        th { background: var(--bg); font-weight: 600; }
        tr:last-child td { border-bottom: none; }
        tr:hover { background: var(--hover); }
        .folder, .file { font-weight: 500; }
        .icon { color: var(--icon); margin-right: 0.5em; font-size: 1.2em; vertical-align: middle; }
        .download-btn, .copy-btn { background: none; border: none; cursor: pointer; color: var(--icon); font-size: 1.1em; margin-left: 0.5em; }
        .download-btn:hover, .copy-btn:hover { color: var(--primary); }
        @media (max-width: 700px) {
            .container { padding: 0.5rem; }
            table, thead, tbody, th, td, tr { display: block; width: 100%; }
            th, td { box-sizing: border-box; width: 100%; }
            tr { margin-bottom: 1em; }
        }
        a { color: var(--primary); text-decoration: none; }
        a:hover { text-decoration: underline; }
        a:visited { color: var(--primary); }
        [data-theme="dark"] a { color: #58a6ff; }
        [data-theme="dark"] a:visited { color: #a5d6ff; }
    </style>
    <script>
    // Dark mode toggle logic
    function setTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem('theme', theme);
    }
    function toggleTheme() {
        const current = document.documentElement.getAttribute('data-theme');
        setTheme(current === 'dark' ? 'light' : 'dark');
    }
    (function() {
        // Sort order memory
        var url = new URL(window.location.href);
        var sort = url.searchParams.get('sort');
        if (!sort) {
            var savedSort = localStorage.getItem('sortOrder');
            if (savedSort === 'asc' || savedSort === 'desc') {
                url.searchParams.set('sort', savedSort);
                window.location.replace(url.toString());
            }
        } else {
            localStorage.setItem('sortOrder', sort);
        }
        // Theme memory
        const saved = localStorage.getItem('theme');
        if (saved) {
            setTheme(saved);
        } else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
            setTheme('dark');
        } else {
            setTheme('light');
        }
    })();
    function copyToClipboard(text) {
        navigator.clipboard.writeText(text).then(function() {
            alert('Copied to clipboard!');
        }, function() {
            alert('Failed to copy.');
        });
    }
    </script>
</head>
<body>
    <div class="container">
        <div class="theme-toggle">
            <button onclick="toggleTheme()" aria-label="Toggle dark mode">🌓 Theme</button>
        </div>
        <h1>Geonet Open Data Browser</h1>
        <div class="breadcrumb" aria-label="Breadcrumb">
            <a href="/?prefix=&page=1&sort={{.SortOrder}}&limit={{.Limit}}">Home</a>
            {{range $i, $b := .Breadcrumbs}}
                / {{if eq (add $i 1) (len $.Breadcrumbs)}}<span class="current">{{$b.Name}}</span>{{else}}<a href="{{$b.Path}}">{{$b.Name}}</a>{{end}}
            {{end}}
        </div>
        {{if .ParentPrefix}}
        <p><a href="?prefix={{.ParentPrefix}}&page=1&sort={{.SortOrder}}&limit={{.Limit}}" aria-label="Parent folder">⬅️ Parent folder</a></p>
        {{end}}
        <div class="controls">
            <span class="sort-toggle">
                <a href="?prefix={{.Prefix}}&page=1&sort={{if eq .SortOrder "asc"}}desc{{else}}asc{{end}}&limit={{.Limit}}">
                    Sort: {{if eq .SortOrder "asc"}}⬆️{{else}}⬇️{{end}}
                </a>
            </span>
            <span class="limit-toggle">
                Show:
                {{range $v := (slice 25 50 75 100)}}
                    <a href="?prefix={{$.Prefix}}&page=1&sort={{$.SortOrder}}&limit={{$v}}"{{if eq $.Limit $v}} class="active"{{end}}>{{$v}}</a>
                {{end}}
            </span>
        </div>
        {{if gt .TotalPages 1}}
        <div class="pagination" aria-label="Pagination">
            {{if gt .Page 1}}
                <a href="?prefix={{.Prefix}}&page={{dec .Page}}&sort={{$.SortOrder}}&limit={{$.Limit}}">⬅️ Prev</a>
            {{end}}
            {{range $i := until .TotalPages}}
                <a href="?prefix={{$.Prefix}}&page={{add $i 1}}&sort={{$.SortOrder}}&limit={{$.Limit}}"{{if eq $.Page (add $i 1)}} class="active"{{end}}>{{add $i 1}}</a>
            {{end}}
            {{if lt .Page .TotalPages}}
                <a href="?prefix={{.Prefix}}&page={{inc .Page}}&sort={{$.SortOrder}}&limit={{$.Limit}}">Next ➡️</a>
            {{end}}
        </div>
        {{end}}
        <table aria-label="File and folder list">
            <thead>
                <tr>
                    <th><a href="?prefix={{.Prefix}}&page=1&sortby=name&sort={{if and (eq .SortBy "name") (eq .SortOrder "asc")}}desc{{else}}asc{{end}}&limit={{.Limit}}">Name{{if eq .SortBy "name"}} {{if eq .SortOrder "asc"}}⬆️{{else}}⬇️{{end}}{{end}}</a></th>
                    <th class="date"><a href="?prefix={{.Prefix}}&page=1&sortby=date&sort={{if and (eq .SortBy "date") (eq .SortOrder "asc")}}desc{{else}}asc{{end}}&limit={{.Limit}}">Last Modified{{if eq .SortBy "date"}} {{if eq .SortOrder "asc"}}⬆️{{else}}⬇️{{end}}{{end}}</a></th>
                    <th class="size"><a href="?prefix={{.Prefix}}&page=1&sortby=size&sort={{if and (eq .SortBy "size") (eq .SortOrder "asc")}}desc{{else}}asc{{end}}&limit={{.Limit}}">Size{{if eq .SortBy "size"}} {{if eq .SortOrder "asc"}}⬆️{{else}}⬇️{{end}}{{end}}</a></th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                {{if eq (len .Objects) 0}}
                <tr>
                    <td colspan="4" style="text-align:center; color:#888;">This folder is empty.</td>
                </tr>
                {{end}}
                {{range .Objects}}
                <tr>
                    <td>
                        {{if .IsDirectory}}
                        <span class="icon" aria-label="Folder">📁</span> <a href="?prefix={{.Key}}&page=1&sortby={{$.SortBy}}&sort={{$.SortOrder}}&limit={{$.Limit}}" class="folder">{{.Name}}</a>
                        {{else}}
                        <span class="icon" aria-label="File">{{if eq .Type "pdf"}}📄{{else if eq .Type "jpg"}}🖼️{{else if eq .Type "jpeg"}}🖼️{{else if eq .Type "png"}}🖼️{{else if eq .Type "txt"}}📄{{else if eq .Type "csv"}}📑{{else if eq .Type "zip"}}🗜️{{else if eq .Type "json"}}📝{{else}}📄{{end}}</span> <a href="/{{.Key}}" class="file">{{.Name}}</a>
                        {{end}}
                    </td>
                    <td class="date">{{.LastModified}}</td>
                    <td class="size">{{if .IsDirectory}}-{{else}}{{formatSize .Size}}{{end}}</td>
                    <td>
                        {{if not .IsDirectory}}
                        <a href="/{{.Key}}" download class="download-btn" aria-label="Download">⬇️</a>
                        <button class="copy-btn" aria-label="Copy S3 URL" onclick="copyToClipboard('{{.S3URL}}')">🔗</button>
                        {{end}}
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
</body>
</html>`
	sortOrderAsc  = "asc"
	sortOrderDesc = "desc"
)

// The entry point for your application.
//
// Use this function to define your main request handling logic. It could be
// used to route based on the request properties (such as method or path), send
// the request to a backend, make completely new requests, and/or generate
// synthetic responses.

// handleFileRequest handles requests for individual files
func handleFileRequest(ctx context.Context, w fsthttp.ResponseWriter, fileKey string) error {
	s3URL := fmt.Sprintf("https://geonet-open-data.s3-ap-southeast-2.amazonaws.com/%s", fileKey)
	req, err := fsthttp.NewRequest("GET", s3URL, nil)
	if err != nil {
		w.WriteHeader(fsthttp.StatusInternalServerError)
		if _, err := fmt.Fprintf(w, "Error creating S3 request: %v\n", err); err != nil {
			fmt.Printf("Error writing response: %v\n", err)
		}
		return err
	}
	resp, err := req.Send(ctx, "TheOrigin")
	if err != nil {
		w.WriteHeader(fsthttp.StatusBadGateway)
		if _, err := fmt.Fprintf(w, "Error fetching from S3: %v\n", err); err != nil {
			fmt.Printf("Error writing response: %v\n", err)
		}
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		fmt.Printf("Error copying response body: %v\n", err)
		return err
	}
	return nil
}

// sortObjects sorts folders and files based on the given criteria
func sortObjects(folders, files []S3Object, sortBy, sortOrder string) []S3Object {
	// Sort folders by name only
	sort.Slice(folders, func(i, j int) bool {
		if sortOrder == sortOrderDesc {
			return folders[i].Name > folders[j].Name
		}
		return folders[i].Name < folders[j].Name
	})

	// Sort files by selected column
	sort.Slice(files, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "date":
			less = files[i].LastModified < files[j].LastModified
		case "size":
			less = files[i].Size < files[j].Size
		default:
			less = files[i].Name < files[j].Name
		}
		if sortOrder == sortOrderDesc {
			return !less
		}
		return less
	})

	// Combine for pagination: folders always first
	return append(append([]S3Object{}, folders...), files...)
}

// paginateObjects handles pagination of the sorted objects
func paginateObjects(allItems []S3Object, page, limit int) ([]S3Object, int, int) {
	total := len(allItems)
	totalPages := (total + limit - 1) / limit
	if page > totalPages {
		page = totalPages
	}
	if page < 1 {
		page = 1
	}
	start := (page - 1) * limit
	end := start + limit
	if end > total {
		end = total
	}
	if total == 0 {
		return allItems, totalPages, total
	}
	return allItems[start:end], totalPages, total
}

// generateBreadcrumbs creates the breadcrumb navigation
func generateBreadcrumbs(prefix, sortOrder string, limit int) []Breadcrumb {
	if prefix == "" {
		return nil
	}

	parts := strings.Split(strings.TrimSuffix(prefix, "/"), "/")
	breadcrumbs := make([]Breadcrumb, 0, len(parts))
	accum := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		if accum != "" {
			accum += "/"
		}
		accum += part
		breadcrumbs = append(breadcrumbs, Breadcrumb{
			Name: part,
			Path: fmt.Sprintf("?prefix=%s&page=1&sort=%s&limit=%d", url.QueryEscape(accum+"/"), sortOrder, limit),
		})
	}
	return breadcrumbs
}

// getParentPrefix computes the parent prefix path
func getParentPrefix(prefix string) string {
	if prefix == "" {
		return ""
	}
	parts := strings.Split(strings.TrimSuffix(prefix, "/"), "/")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], "/") + "/"
	}
	return ""
}

// addFileMetadata adds type and S3 URL to file objects
func addFileMetadata(items []S3Object) {
	for i := range items {
		if !items[i].IsDirectory {
			if idx := strings.LastIndex(items[i].Name, "."); idx != -1 {
				items[i].Type = strings.ToLower(items[i].Name[idx+1:])
			} else {
				items[i].Type = "file"
			}
			items[i].S3URL = "https://geonet-open-data.s3-ap-southeast-2.amazonaws.com/" + items[i].Key
		}
	}
}

// handleBrowserUI handles the browser UI rendering
func handleBrowserUI(w fsthttp.ResponseWriter, prefix string, page int, sortBy string, sortOrder string, limit int, tmpl *template.Template) error {
	objects, err := listObjects(prefix)
	if err != nil {
		w.WriteHeader(fsthttp.StatusInternalServerError)
		if _, err := fmt.Fprintf(w, "Error listing objects: %v\n", err); err != nil {
			fmt.Printf("Error writing response: %v\n", err)
		}
		return err
	}

	// Separate folders and files
	var folders, files []S3Object
	for _, obj := range objects {
		if obj.IsDirectory {
			folders = append(folders, obj)
		} else {
			files = append(files, obj)
		}
	}

	// Sort and paginate objects
	allItems := sortObjects(folders, files, sortBy, sortOrder)
	pageItems, totalPages, total := paginateObjects(allItems, page, limit)

	// Generate navigation elements
	breadcrumbs := generateBreadcrumbs(prefix, sortOrder, limit)
	parentPrefix := getParentPrefix(prefix)

	// Add metadata to files
	addFileMetadata(pageItems)

	// Set content type and render template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := PageData{
		Objects:      pageItems,
		Breadcrumbs:  breadcrumbs,
		CurrentPath:  prefix,
		ParentPrefix: parentPrefix,
	}
	if err := tmpl.Execute(w, struct {
		PageData
		Page, TotalPages, Limit, Total int
		SortOrder, SortBy, Prefix      string
	}{
		PageData:   data,
		Page:       page,
		TotalPages: totalPages,
		Limit:      limit,
		Total:      total,
		SortOrder:  sortOrder,
		SortBy:     sortBy,
		Prefix:     prefix,
	}); err != nil {
		w.WriteHeader(fsthttp.StatusInternalServerError)
		if _, err := fmt.Fprintf(w, "Error rendering template: %v\n", err); err != nil {
			fmt.Printf("Error writing response: %v\n", err)
		}
		return err
	}
	return nil
}

func main() {
	// Log service version
	fmt.Println("FASTLY_SERVICE_VERSION:", os.Getenv("FASTLY_SERVICE_VERSION"))

	tmpl := template.Must(template.New("browser").Funcs(template.FuncMap{
		"formatSize": formatSize,
		"add":        add,
		"dec":        dec,
		"inc":        inc,
		"until":      until,
		"slice":      slice,
	}).Parse(htmlTemplate))

	fsthttp.ServeFunc(func(ctx context.Context, w fsthttp.ResponseWriter, r *fsthttp.Request) {
		if r.Method != "GET" {
			w.WriteHeader(fsthttp.StatusMethodNotAllowed)
			if _, err := fmt.Fprintf(w, "Method not allowed\n"); err != nil {
				fmt.Printf("Error writing response: %v\n", err)
			}
			return
		}

		// Parse query params
		u, _ := url.Parse(r.URL.String())
		q := u.Query()
		prefix := q.Get("prefix")

		// If this is a file request (no prefix param, path does not end with / and is not "/"), proxy the file
		if prefix == "" && r.URL.Path != "/" && !strings.HasSuffix(r.URL.Path, "/") {
			fileKey := strings.TrimPrefix(r.URL.Path, "/")
			if err := handleFileRequest(ctx, w, fileKey); err != nil {
				return
			}
			return
		}

		// Otherwise, render the browser UI for the given prefix or folder
		if prefix == "/" {
			prefix = ""
		}
		page, _ := strconv.Atoi(q.Get("page"))
		if page < 1 {
			page = 1
		}
		sortBy := q.Get("sortby")
		if sortBy != "date" && sortBy != "size" {
			sortBy = "name"
		}
		sortOrder := q.Get("sort")
		if sortOrder != sortOrderAsc && sortOrder != sortOrderDesc {
			sortOrder = sortOrderAsc
		}
		limit, _ := strconv.Atoi(q.Get("limit"))
		if limit < 1 {
			limit = 25
		}

		if err := handleBrowserUI(w, prefix, page, sortBy, sortOrder, limit, tmpl); err != nil {
			return
		}
	})
}

// Add template helpers for slice, until, add, dec, inc
func add(x, y int) int { return x + y }
func dec(x int) int    { return x - 1 }
func inc(x int) int    { return x + 1 }
func until(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = i
	}
	return s
}
func slice(vals ...int) []int { return vals }
