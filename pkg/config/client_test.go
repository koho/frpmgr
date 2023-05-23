package config

import (
	"reflect"
	"testing"
	"time"
)

func TestUnmarshalClientConfFromIni(t *testing.T) {
	input := `
		[common]
		server_addr = example.com
		server_port = 7001
		token = 123456
		frpmgr_manual_start = true
		frpmgr_delete_method = absolute
		frpmgr_delete_after_date = 2023-03-23T00:00:00Z
		meta_1 = value
		
		[ssh]
		type = tcp
		local_ip = 192.168.1.1
		local_port = 22
		remote_port = 6000
		meta_2 = value
	`
	expected := NewDefaultClientConfig()
	expected.ServerAddress = "example.com"
	expected.ServerPort = "7001"
	expected.Token = "123456"
	expected.ManualStart = true
	expected.Custom = map[string]string{"meta_1": "value"}
	expected.DeleteMethod = "absolute"
	expected.DeleteAfterDate = time.Date(2023, 3, 23, 0, 0, 0, 0, time.UTC)
	expected.Proxies = append(expected.Proxies, &Proxy{
		BaseProxyConf: BaseProxyConf{
			Name:      "ssh",
			Type:      "tcp",
			LocalIP:   "192.168.1.1",
			LocalPort: "22",
			Custom:    map[string]string{"meta_2": "value"},
		},
		RemotePort: "6000",
	})
	cc, err := UnmarshalClientConfFromIni([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cc, expected) {
		t.Errorf("Expected: %v, got: %v", expected, cc)
	}
}

func TestProxyGetAlias(t *testing.T) {
	input := `
		[range:test_tcp]
		type = tcp
		local_ip = 127.0.0.1
		local_port = 6000-6006,6007
		remote_port = 6000-6006,6007
	`
	expected := []string{"test_tcp_0", "test_tcp_1", "test_tcp_2", "test_tcp_3",
		"test_tcp_4", "test_tcp_5", "test_tcp_6", "test_tcp_7"}
	proxy, err := UnmarshalProxyFromIni([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	output := proxy.GetAlias()
	if !reflect.DeepEqual(output, expected) {
		t.Errorf("Expected: %v, got: %v", expected, output)
	}
}
