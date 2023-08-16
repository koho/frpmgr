package util

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"path"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/miekg/dns"
	"golang.org/x/sys/windows"
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

// ResolveSVCB resolves the SVCB resource record of a given host. SVCB RR provides
// the information needed to connect to a service. The function returns a string
// for the target IP or domain name, an uint16 for the port, and an error.
func ResolveSVCB(ctx context.Context, host string, server string) (string, uint16, error) {
	if server == "" {
		if sysDns := GetSystemDnsServer(); sysDns != "" {
			server = net.JoinHostPort(sysDns, "53")
		} else {
			return "", 0, fmt.Errorf("no available dns server")
		}
	}
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(host), dns.TypeSVCB)
	m.RecursionDesired = true
	conn, err := c.DialContext(ctx, server)
	if err != nil {
		return "", 0, err
	}
	defer conn.Close()
	r, _, err := c.ExchangeWithConnContext(ctx, m, conn)
	if err != nil {
		return "", 0, err
	}

	var rr *dns.SVCB
	var ok bool
	if len(r.Answer) > 0 {
		rr, ok = r.Answer[0].(*dns.SVCB)
	}
	if !ok {
		return "", 0, fmt.Errorf("record not found")
	}
	if rr.Priority == 0 {
		return "", 0, fmt.Errorf("not a service mode record")
	}

	var ip string
	var port uint16
	for _, v := range rr.Value {
		switch v := v.(type) {
		case *dns.SVCBIPv4Hint:
			if len(v.Hint) > 0 {
				ip = v.Hint[0].String()
			}
		case *dns.SVCBPort:
			port = v.Port
		}
	}
	if ip == "" {
		if rr.Target == "." {
			return strings.TrimSuffix(rr.Hdr.Name, "."), port, nil
		}
		return strings.TrimSuffix(rr.Target, "."), port, nil
	}
	return ip, port, nil
}
