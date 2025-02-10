package ui

import (
	"net"
	"slices"
	"strconv"
	"strings"

	"github.com/lxn/walk"
	"github.com/samber/lo"

	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/util"
)

type ConfListModel struct {
	walk.ReflectTableModelBase

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

func (m *ConfListModel) Move(i, j int) {
	util.MoveSlice(m.items, i, j)
	m.PublishRowsChanged(min(i, j), max(i, j))
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
		} else if p.Plugin != "" && p.PluginLocalAddr != "" {
			if host, port, err := net.SplitHostPort(p.PluginLocalAddr); err == nil {
				pi.DisplayLocalIP = host
				pi.DisplayLocalPort = port
			} else {
				pi.DisplayLocalIP = p.PluginLocalAddr
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

type ListItem struct {
	Title any
	Value string
}

type ListModel []*ListItem

func NewListModel(values []string, titles ...any) ListModel {
	var items []*ListItem
	for i, value := range values {
		var title any = value
		if i < len(titles) {
			title = titles[i]
		}
		items = append(items, &ListItem{
			Title: title,
			Value: value,
		})
	}
	return items
}

type LogModel struct {
	walk.TableModelBase

	path     string
	offset   int64
	maxLines int
	lines    []string
}

func NewLogModel(paths []string, maxLines int) (*LogModel, bool) {
	m := &LogModel{
		path:     paths[0],
		maxLines: maxLines,
		lines:    make([]string, 0),
	}
	ok := false
	for i, path := range paths {
		lines, k, offset, err := util.ReadFileLines(path, 0, maxLines)
		if err != nil {
			continue
		}
		ok = true
		if i == 0 {
			m.offset = offset
		}
		if k >= 0 {
			for n, j := len(lines), k-1; (j+n)%n != k; j-- {
				m.lines = append(m.lines, lines[(j+n)%n])
			}
			m.lines = append(m.lines, lines[k])
		} else {
			for j := len(lines) - 1; j >= 0; j-- {
				m.lines = append(m.lines, lines[j])
			}
		}
		maxLines -= len(lines)
		if maxLines <= 0 {
			break
		}
	}
	if len(m.lines) > 0 {
		slices.Reverse(m.lines)
	}
	return m, ok
}

func (m *LogModel) write(lines []string, i int) {
	if len(lines) == 0 {
		return
	}
	if m.maxLines > 0 && len(m.lines) >= m.maxLines {
		copy(m.lines[:], m.lines[len(lines):])
		m.lines = m.lines[:len(m.lines)-len(lines)]
		m.PublishRowsRemoved(0, len(lines)-1)
		m.PublishRowsChanged(0, len(m.lines)-1)
	}
	from := len(m.lines)
	if i >= 0 {
		m.lines = append(m.lines, lines[i:]...)
		m.lines = append(m.lines, lines[:i]...)
	} else {
		m.lines = append(m.lines, lines...)
	}
	to := from + len(lines) - 1
	m.PublishRowsInserted(from, to)
}

func (m *LogModel) Value(row, col int) any {
	return m.lines[row]
}

func (m *LogModel) RowCount() int {
	return len(m.lines)
}

func (m *LogModel) Reset() {
	m.offset = 0
}

func (m *LogModel) ReadMore() (int, error) {
	lines, k, offset, err := util.ReadFileLines(m.path, m.offset, m.maxLines)
	if err != nil {
		return 0, err
	}
	m.write(lines, k)
	m.offset = offset
	return len(lines), nil
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
