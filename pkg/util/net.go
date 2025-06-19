package util

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modIPHelp               = syscall.NewLazyDLL("iphlpapi.dll")
	procGetExtendedTcpTable = modIPHelp.NewProc("GetExtendedTcpTable")
	procGetExtendedUdpTable = modIPHelp.NewProc("GetExtendedUdpTable")
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

//nolint:unused
type mibTCPRowOwnerPid struct {
	dwState      uint32
	dwLocalAddr  uint32
	dwLocalPort  uint32
	dwRemoteAddr uint32
	dwRemotePort uint32
	dwOwningPid  uint32
}

//nolint:unused
type mibTCP6RowOwnerPid struct {
	ucLocalAddr     [16]byte
	dwLocalScopeId  uint32
	dwLocalPort     uint32
	ucRemoteAddr    [16]byte
	dwRemoteScopeId uint32
	dwRemotePort    uint32
	dwState         uint32
	dwOwningPid     uint32
}

//nolint:unused
type mibUDPRowOwnerPid struct {
	dwLocalAddr uint32
	dwLocalPort uint32
	dwOwningPid uint32
}

//nolint:unused
type mibUDP6RowOwnerPid struct {
	ucLocalAddr    [16]byte
	dwLocalScopeId uint32
	dwLocalPort    uint32
	dwOwningPid    uint32
}

type mibTableOwnerPid[T any] struct {
	dwNumEntries uint32
	table        [1]T
}

// countConnections returns the number of IPv4 and IPv6 connections that match the given filter.
//   - https://learn.microsoft.com/en-us/windows/win32/api/iphlpapi/nf-iphlpapi-getextendedtcptable
//   - https://learn.microsoft.com/en-us/windows/win32/api/iphlpapi/nf-iphlpapi-getextendedudptable
func countConnections[R4, R6 any](proc *syscall.LazyProc, tableClass uintptr, filter4 func(R4) bool, filter6 func(R6) bool) (count int) {
	var size uint32
	var buf []byte
	getTable := func(af uintptr) bool {
		for {
			var pTable *byte
			if len(buf) > 0 {
				pTable = &buf[0]
			}
			ret, _, _ := proc.Call(uintptr(unsafe.Pointer(pTable)), uintptr(unsafe.Pointer(&size)), 0, af, tableClass, 0)
			if ret != 0 {
				if errors.Is(syscall.Errno(ret), syscall.ERROR_INSUFFICIENT_BUFFER) {
					buf = make([]byte, int(size))
					continue
				}
				return false
			}
			return true
		}
	}
	if getTable(windows.AF_INET) {
		table := (*mibTableOwnerPid[R4])(unsafe.Pointer(&buf[0]))
		for _, conn := range unsafe.Slice(&table.table[0], table.dwNumEntries) {
			if filter4(conn) {
				count++
			}
		}
	}
	if getTable(windows.AF_INET6) {
		table := (*mibTableOwnerPid[R6])(unsafe.Pointer(&buf[0]))
		for _, conn := range unsafe.Slice(&table.table[0], table.dwNumEntries) {
			if filter6(conn) {
				count++
			}
		}
	}
	return
}

// CountTCPConnections returns the number of connected TCP endpoints for a given process.
func CountTCPConnections(pid uint32) int {
	return countConnections(procGetExtendedTcpTable, 4, func(r4 mibTCPRowOwnerPid) bool {
		return r4.dwOwningPid == pid
	}, func(r6 mibTCP6RowOwnerPid) bool {
		return r6.dwOwningPid == pid
	})
}

// CountUDPConnections returns the number of UDP endpoints for a given process.
func CountUDPConnections(pid uint32) int {
	return countConnections(procGetExtendedUdpTable, 1, func(r4 mibUDPRowOwnerPid) bool {
		return r4.dwOwningPid == pid
	}, func(r6 mibUDP6RowOwnerPid) bool {
		return r6.dwOwningPid == pid
	})
}
