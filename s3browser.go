package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/fastly/compute-sdk-go/fsthttp"
)

const (
	bucketName = "geonet-open-data"
	region     = "ap-southeast-2"
	bucketURL  = "https://geonet-open-data.s3-ap-southeast-2.amazonaws.com"
)

// ListBucketResult represents the XML response from S3 ListObjects API
type ListBucketResult struct {
	XMLName        xml.Name       `xml:"ListBucketResult"`
	CommonPrefixes []CommonPrefix `xml:"CommonPrefixes"`
	Contents       []Object       `xml:"Contents"`
}

// CommonPrefix represents a directory prefix in the S3 bucket
type CommonPrefix struct {
	Prefix string `xml:"Prefix"`
}

// Object represents a file in the S3 bucket
type Object struct {
	Key          string    `xml:"Key"`
	LastModified time.Time `xml:"LastModified"`
	Size         int64     `xml:"Size"`
}

func listObjects(prefix string) ([]S3Object, error) {
	// Ensure prefix ends with / if it's not empty
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	// Construct the URL for listing objects
	url := fmt.Sprintf("%s/?prefix=%s&delimiter=/", bucketURL, prefix)

	// Create a new request
	req, err := fsthttp.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Send the request
	resp, err := req.Send(context.Background(), "TheOrigin")
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != fsthttp.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the XML response
	var result ListBucketResult
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %v", err)
	}

	// Pre-allocate slice with estimated capacity
	objects := make([]S3Object, 0, len(result.CommonPrefixes)+len(result.Contents))

	// Add common prefixes (directories)
	for _, cp := range result.CommonPrefixes {
		key := cp.Prefix
		name := path.Base(strings.TrimSuffix(key, "/"))
		href := path.Join(prefix, name) + "/"
		objects = append(objects, S3Object{
			Key:         key,
			Name:        name,
			IsDirectory: true,
			Href:        href,
		})
	}

	// Add objects (files)
	for _, object := range result.Contents {
		key := object.Key
		// Skip the current directory marker (object with key == prefix)
		if prefix != "" && key == prefix {
			continue
		}
		// Skip objects that are actually directories (they end with /)
		if strings.HasSuffix(key, "/") {
			continue
		}
		objects = append(objects, S3Object{
			Key:          key,
			Name:         path.Base(key),
			LastModified: object.LastModified.Format("2006-01-02 15:04:05"),
			Size:         object.Size,
			IsDirectory:  false,
		})
	}

	return objects, nil
}

func formatSize(size int64) string {
	switch {
	case size < 1024:
		return fmt.Sprintf("%d B", size)
	case size < 1024*1024:
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	case size < 1024*1024*1024:
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	default:
		return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
	}
}
