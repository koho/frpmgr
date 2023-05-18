package ui

import (
	"os"
	"strings"

	"github.com/lxn/walk"
	"github.com/thoas/go-funk"

	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/util"
)

type ConfListModel struct {
	walk.TableModelBase

	items []*Conf
}

func NewConfListModel(items []*Conf) *ConfListModel {
	m := new(ConfListModel)
	m.items = items
	return m
}

func (m *ConfListModel) Items() interface{} {
	for _, x := range m.items {
		if x.Data.Expiry() {
			x.DisplayName = x.Name + "🕓"
		} else {
			x.DisplayName = x.Name
		}
	}
	return m.items
}

type ListModel struct {
	walk.ListModelBase

	items []*Conf
}

func NewListModel(items []*Conf) *ListModel {
	m := new(ListModel)
	m.items = items
	return m
}

func (m *ListModel) Value(index int) interface{} {
	return m.items[index].Name
}

func (m *ListModel) ItemCount() int {
	return len(m.items)
}

type ProxyModel struct {
	walk.ReflectTableModelBase

	conf  *Conf
	data  *config.ClientConfig
	items []ProxyItem
}

type ProxyItem struct {
	*config.Proxy
	// Domains is a list of domains bound to this proxy
	Domains string
	// DisplayLocalIP changes the local address shown in table
	DisplayLocalIP string
	// DisplayLocalPort changes the local port shown in table
	DisplayLocalPort string
}

func NewProxyModel(conf *Conf) *ProxyModel {
	m := new(ProxyModel)
	m.conf = conf
	m.data = conf.Data.(*config.ClientConfig)
	m.items = funk.Map(m.data.Proxies, func(p *config.Proxy) ProxyItem {
		pi := ProxyItem{Proxy: p, DisplayLocalIP: p.LocalIP, DisplayLocalPort: p.LocalPort}
		// Combine subdomain and custom domains to form a list of domains
		pi.Domains = strings.Join(funk.FilterString([]string{p.SubDomain, p.CustomDomains}, func(s string) bool {
			return strings.TrimSpace(s) != ""
		}), ",")
		// Show bind address and server name for visitor
		if p.IsVisitor() {
			pi.Domains = p.ServerName
			pi.DisplayLocalIP = p.BindAddr
			pi.DisplayLocalPort = p.BindPort
		}
		return pi
	}).([]ProxyItem)
	return m
}

func (m *ProxyModel) Items() interface{} {
	return m.items
}

// StringPair is a simple struct to hold a pair of strings.
type StringPair struct {
	Name        string
	DisplayName string
}

// NewDefaultListModel creates a default item at the top of the model.
func NewDefaultListModel(items []string, defaultKey string, defaultName string) []*StringPair {
	listItems := make([]*StringPair, 0, len(items)+1)
	listItems = append(listItems, &StringPair{Name: defaultKey, DisplayName: defaultName})
	for _, item := range items {
		listItems = append(listItems, &StringPair{Name: item, DisplayName: item})
	}
	return listItems
}

// NewStringPairModel creates a slice of string pair from two string slices.
func NewStringPairModel(keys []string, values []string, defaultValue string) []*StringPair {
	listItems := make([]*StringPair, 0, len(keys))
	for i, k := range keys {
		pair := &StringPair{Name: k, DisplayName: values[i]}
		if pair.DisplayName == "" {
			pair.DisplayName = defaultValue
		}
		listItems = append(listItems, pair)
	}
	return listItems
}

type TextLine struct {
	Text string
}

type LogModel struct {
	walk.ReflectTableModelBase

	path  string
	lines []*TextLine
}

func NewLogModel(path string) *LogModel {
	m := new(LogModel)
	m.path = path
	m.lines = make([]*TextLine, 0)
	return m
}

func (m *LogModel) Items() interface{} {
	return m.lines
}

// Reset reload the whole log file from disk
func (m *LogModel) Reset() error {
	if m.path == "" {
		return os.ErrInvalid
	}
	textLines, err := util.ReadFileLines(m.path)
	if err != nil {
		return err
	}
	lines := make([]*TextLine, 0)
	for _, line := range textLines {
		lines = append(lines, &TextLine{Text: line})
	}
	m.lines = lines
	return nil
}
