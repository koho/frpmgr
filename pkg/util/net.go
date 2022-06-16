package util

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"golang.org/x/sys/windows"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"path"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	modIPHelp            = syscall.NewLazyDLL("iphlpapi.dll")
	procGetNetworkParams = modIPHelp.NewProc("GetNetworkParams")
)

// LookupIP lookups the IP address of the given host name with the given dns server
func LookupIP(addr string, server string) (string, error) {
	if net.ParseIP(addr) != nil {
		return addr, nil
	}
	c := dns.Client{}
	m := dns.Msg{}
	if !strings.HasSuffix(addr, ".") {
		addr += "."
	}
	if !strings.Contains(server, ":") {
		server += ":53"
	}
	m.SetQuestion(addr, dns.TypeA)
	r, _, err := c.Exchange(&m, server)
	if err != nil {
		return "", err
	}
	if len(r.Answer) == 0 {
		m.SetQuestion(addr, dns.TypeAAAA)
		if r, _, err = c.Exchange(&m, server); err != nil {
			return "", err
		}
		if len(r.Answer) == 0 {
			return "", errors.New(fmt.Sprintf("no record for host '%s' with '%s'", addr, server))
		}
	}
	switch v := r.Answer[0].(type) {
	case *dns.A:
		return v.A.String(), nil
	case *dns.AAAA:
		return v.AAAA.String(), nil
	case *dns.CNAME:
		return LookupIP(v.Target, server)
	default:
		return "", errors.New(fmt.Sprintf("host '%s' lookup failed with '%s'", addr, server))
	}
}

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
		data, err = ioutil.ReadAll(resp.Body)
		return filename, mediaType, data, err
	} else {
		return "", "", nil, err
	}
}
