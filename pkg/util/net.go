package util

import (
	"bytes"
	"context"
	"fmt"
	"golang.org/x/sys/windows"
	"io"
	"mime"
	"net/http"
	"path"
	"syscall"
	"time"
	"unsafe"
)

var (
	modIPHelp            = syscall.NewLazyDLL("iphlpapi.dll")
	procGetNetworkParams = modIPHelp.NewProc("GetNetworkParams")
)

// GetSystemDnsServer returns the dns server used by local system
func GetSystemDnsServer() string {
	type ipAddress struct {
		next      *ipAddress
		ipAddress [4 * 4]byte
		ipMask    [4 * 4]byte
		context   uint32
	}
	type networkInfo struct {
		hostName         [windows.MAX_ADAPTER_DESCRIPTION_LENGTH + 4]byte
		domainName       [windows.MAX_ADAPTER_DESCRIPTION_LENGTH + 4]byte
		currentDnsServer *ipAddress
		dnsServerList    ipAddress
		// We only care about dns, remaining fields can be ignored
		// ...
	}
	info := &networkInfo{}
	size := uint32(unsafe.Sizeof(info))
	if r1, _, _ := procGetNetworkParams.Call(uintptr(unsafe.Pointer(info)), uintptr(unsafe.Pointer(&size))); syscall.Errno(r1) == windows.ERROR_BUFFER_OVERFLOW {
		newBuffer := make([]byte, size)
		info = (*networkInfo)(unsafe.Pointer(&newBuffer[0]))
	}
	if r1, _, _ := procGetNetworkParams.Call(uintptr(unsafe.Pointer(info)), uintptr(unsafe.Pointer(&size))); r1 == 0 {
		length := bytes.IndexByte(info.dnsServerList.ipAddress[:], 0)
		return string(info.dnsServerList.ipAddress[:length:length])
	}
	return ""
}

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
