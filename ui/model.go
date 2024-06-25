package ui

import (
	"os"
	"strconv"
	"strings"

	"github.com/lxn/walk"
	"github.com/samber/lo"

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
	m.items = lo.Map(m.data.Proxies, func(p *config.Proxy, i int) ProxyItem {
		pi := ProxyItem{Proxy: p, DisplayLocalIP: p.LocalIP, DisplayLocalPort: p.LocalPort}
		// Combine subdomain and custom domains to form a list of domains
		pi.Domains = strings.Join(lo.Filter([]string{p.SubDomain, p.CustomDomains}, func(s string, i int) bool {
			return strings.TrimSpace(s) != ""
		}), ",")
		// Show bind address and server name for visitor
		if p.IsVisitor() {
			pi.Domains = p.ServerName
			pi.DisplayLocalIP = p.BindAddr
			if p.BindPort > 0 {
				pi.DisplayLocalPort = strconv.Itoa(p.BindPort)
			}
		}
		return pi
	})
	return m
}

func (m *ProxyModel) Items() interface{} {
	return m.items
}

func (m *ProxyModel) Move(i, j int) {
	util.MoveSlice(m.items, i, j)
	util.MoveSlice(m.data.Proxies, i, j)
	m.PublishRowsChanged(min(i, j), max(i, j))
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

// NonSortedModel preserves the original order of items
// in the slice.
type NonSortedModel[T any] struct {
	walk.ReflectTableModelBase

	items []*T
}

func NewNonSortedModel[T any](items []*T) *NonSortedModel[T] {
	return &NonSortedModel[T]{items: items}
}

func (m *NonSortedModel[T]) Items() interface{} {
	return m.items
}

// AttributeModel is a list of name-value pairs.
type AttributeModel struct {
	walk.TableModelBase
	data [][2]string
}

func NewAttributeModel(attrs map[string]string) *AttributeModel {
	m := &AttributeModel{data: make([][2]string, 0, len(attrs))}
	for k, v := range attrs {
		m.data = append(m.data, [2]string{k, v})
	}
	return m
}

func (a *AttributeModel) Value(row, col int) interface{} {
	var empty string
	if row >= 0 && row < len(a.data) && col >= 0 && col < 2 {
		return &a.data[row][col]
	}
	return &empty
}

func (a *AttributeModel) RowCount() int {
	return len(a.data)
}

func (a *AttributeModel) Add(k, v string) {
	a.data = append(a.data, [2]string{k, v})
	i := len(a.data) - 1
	a.PublishRowsInserted(i, i)
}

func (a *AttributeModel) Delete(i int) {
	if i >= 0 && i < len(a.data) {
		a.data = append(a.data[:i], a.data[i+1:]...)
		a.PublishRowsRemoved(i, i)
	}
}

func (a *AttributeModel) Clear() {
	a.data = nil
	a.PublishRowsReset()
}

func (a *AttributeModel) AsMap() map[string]string {
	if len(a.data) == 0 {
		return nil
	}
	var m map[string]string
	for _, pair := range a.data {
		if k := strings.TrimSpace(pair[0]); k != "" {
			if m == nil {
				m = make(map[string]string)
			}
			m[k] = strings.TrimSpace(pair[1])
		}
	}
	return m
}
