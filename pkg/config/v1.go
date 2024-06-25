package config

import (
	"encoding/json"

	"github.com/fatedier/frp/pkg/config/v1"
)

type ClientConfigV1 struct {
	v1.ClientCommonConfig

	Proxies  []TypedProxyConfig   `json:"proxies,omitempty"`
	Visitors []TypedVisitorConfig `json:"visitors,omitempty"`

	Mgr Mgr `json:"frpmgr,omitempty"`
}

type Mgr struct {
	ManualStart bool       `json:"manualStart,omitempty"`
	SVCBEnable  bool       `json:"svcbEnable,omitempty"`
	AutoDelete  AutoDelete `json:"autoDelete,omitempty"`
}

type TypedProxyConfig struct {
	v1.TypedProxyConfig
	Mgr ProxyMgr `json:"frpmgr,omitempty"`
}

type TypedVisitorConfig struct {
	v1.TypedVisitorConfig
	Mgr ProxyMgr `json:"frpmgr,omitempty"`
}

type ProxyMgr struct {
	Range RangePort `json:"range,omitempty"`
	Sort  int       `json:"sort,omitempty"`
}

type RangePort struct {
	Local  string `json:"local"`
	Remote string `json:"remote"`
}

func (c *TypedProxyConfig) UnmarshalJSON(b []byte) error {
	err := c.TypedProxyConfig.UnmarshalJSON(b)
	if err != nil {
		return err
	}
	c.Mgr, err = unmarshalProxyMgr(b)
	return err
}

func (c *TypedVisitorConfig) UnmarshalJSON(b []byte) error {
	err := c.TypedVisitorConfig.UnmarshalJSON(b)
	if err != nil {
		return err
	}
	c.Mgr, err = unmarshalProxyMgr(b)
	return err
}

func unmarshalProxyMgr(b []byte) (c ProxyMgr, err error) {
	s := struct {
		Mgr ProxyMgr `json:"frpmgr"`
	}{}
	if err = json.Unmarshal(b, &s); err != nil {
		return
	}
	c = s.Mgr
	return
}
