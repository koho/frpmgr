package util

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"time"
)

// DownloadFile downloads a file from the given url
func DownloadFile(ctx context.Context, url string) (filename, mediaType string, data []byte, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", resp.Status)
		return
	}
	// Use the filename in header
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			filename = params["filename"]
		}
	}
	// Use the base filename part of the URL
	if filename == "" {
		filename = path.Base(resp.Request.URL.Path)
	}
	if mediaType, _, err = mime.ParseMediaType(resp.Header.Get("Content-Type")); err == nil {
		data, err = io.ReadAll(resp.Body)
		return filename, mediaType, data, err
	} else {
		return "", "", nil, err
	}
}
