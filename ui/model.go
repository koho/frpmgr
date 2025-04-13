package ui

import (
	"net"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/lxn/walk"
	"github.com/samber/lo"

	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/util"
)

type ConfListModel struct {
	walk.ReflectTableModelBase
	sync.Mutex

	items              []*Conf
	rowEditedPublisher walk.IntEventPublisher
}

func NewConfListModel(items []*Conf) *ConfListModel {
	m := new(ConfListModel)
	m.items = items
	return m
}

func (m *ConfListModel) Value(row, col int) interface{} {
	return m.items[row].Name()
}

func (m *ConfListModel) SetStateByPath(path string, state consts.ConfigState) bool {
	if i := slices.IndexFunc(m.items, func(conf *Conf) bool {
		return conf.Path == path
	}); i >= 0 && m.items[i].State != state {
		m.items[i].State = state
		m.PublishRowChanged(i)
		return true
	}
	return false
}

func (m *ConfListModel) SetStateByConf(conf *Conf, state consts.ConfigState) bool {
	if i := slices.Index(m.items, conf); i >= 0 && m.items[i].State != state {
		m.items[i].State = state
		m.PublishRowChanged(i)
		return true
	}
	return false
}

func (m *ConfListModel) List() []*Conf {
	m.Lock()
	defer m.Unlock()
	cfgList := make([]*Conf, len(m.items))
	copy(cfgList, m.items)
	return cfgList
}

func (m *ConfListModel) Move(i, j int) {
	m.Lock()
	defer m.Unlock()
	util.MoveSlice(m.items, i, j)
	m.PublishRowsChanged(min(i, j), max(i, j))
	setConfOrder(m.items)
}

func (m *ConfListModel) RowCount() int {
	return len(m.items)
}

func (m *ConfListModel) Add(item ...*Conf) {
	m.Lock()
	defer m.Unlock()
	from := len(m.items)
	m.items = append(m.items, item...)
	m.PublishRowsInserted(from, from+len(item)-1)
	setConfOrder(m.items)
}

func (m *ConfListModel) Remove(index ...int) {
	if len(index) == 0 {
		return
	}
	m.Lock()
	defer m.Unlock()
	i := index[0]
	if len(index) == 1 {
		m.items = append(m.items[:i], m.items[i+1:]...)
		m.PublishRowsRemoved(i, i)
	} else {
		for i, idx := range index {
			// index must be sorted.
			j := idx - i
			m.items = append(m.items[:j], m.items[j+1:]...)
		}
		m.PublishRowsReset()
	}
	if i <= len(m.items)-1 {
		m.PublishRowsChanged(i, len(m.items)-1)
	}
	setConfOrder(m.items)
}

func (m *ConfListModel) RowEdited() *walk.IntEvent {
	return m.rowEditedPublisher.Event()
}

func (m *ConfListModel) PublishRowEdited(i int) {
	m.rowEditedPublisher.Publish(i)
}

type ProxyModel struct {
	walk.ReflectTableModelBase

	conf                  *Conf
	data                  *config.ClientConfig
	items                 []*ProxyRow
	beforeRemovePublisher walk.IntEventPublisher
	rowEditedPublisher    walk.IntEventPublisher
	rowRenamedPublisher   walk.IntEventPublisher
}

type ProxyRow struct {
	*config.Proxy
	// Domains is a list of domains bound to this proxy
	Domains string
	// DisplayLocalIP changes the local address shown in table
	DisplayLocalIP string
	// DisplayLocalPort changes the local port shown in table
	DisplayLocalPort string
	// DisplayRemotePort changes the remote port shown in table.
	DisplayRemotePort string
	// Running state.
	State consts.ProxyState
	// Error message.
	Error string
	// Name of the proxy reporting the status.
	StateSource string
	// Remote address got from server.
	RemoteAddr string
}

func NewProxyRow(p *config.Proxy) *ProxyRow {
	return fillProxyRow(p, new(ProxyRow))
}

// UpdateRemotePort attempts to display the remote port obtained from
// the server if the requested remote port is empty.
func (m *ProxyRow) UpdateRemotePort() {
	if m.RemoteAddr != "" && (m.RemotePort == "" || m.RemotePort == "0") {
		addr := strings.Split(m.RemoteAddr, ",")[0]
		if _, port, err := net.SplitHostPort(addr); err == nil && port != "" {
			m.DisplayRemotePort = "(" + port + ")"
			if m.Type == consts.ProxyTypeTCP || m.Type == consts.ProxyTypeUDP {
				m.DisplayRemotePort = "0 " + m.DisplayRemotePort
			}
			return
		}
	}
	m.DisplayRemotePort = m.RemotePort
}

func fillProxyRow(p *config.Proxy, pr *ProxyRow) *ProxyRow {
	pr.Proxy = p
	pr.DisplayLocalIP = p.LocalIP
	pr.DisplayLocalPort = p.LocalPort
	// Combine subdomain and custom domains to form a list of domains
	pr.Domains = strings.Join(lo.Filter([]string{p.SubDomain, p.CustomDomains}, func(s string, i int) bool {
		return strings.TrimSpace(s) != ""
	}), ",")
	// Show bind address and server name for visitor
	if p.IsVisitor() {
		pr.Domains = p.ServerName
		pr.DisplayLocalIP = p.BindAddr
		if p.BindPort > 0 {
			pr.DisplayLocalPort = strconv.Itoa(p.BindPort)
		}
	} else if p.Plugin != "" && p.PluginLocalAddr != "" {
		if host, port, err := net.SplitHostPort(p.PluginLocalAddr); err == nil {
			pr.DisplayLocalIP = host
			pr.DisplayLocalPort = port
		} else {
			pr.DisplayLocalIP = p.PluginLocalAddr
		}
	}
	pr.UpdateRemotePort()
	return pr
}

func NewProxyModel(conf *Conf) *ProxyModel {
	m := new(ProxyModel)
	m.conf = conf
	m.data = conf.Data.(*config.ClientConfig)
	m.items = lo.Map(m.data.Proxies, func(p *config.Proxy, i int) *ProxyRow {
		return NewProxyRow(p)
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

func (m *ProxyModel) Add(proxy ...*config.Proxy) {
	from := len(m.items)
	for _, item := range proxy {
		m.items = append(m.items, NewProxyRow(item))
		m.data.AddItem(item)
	}
	m.PublishRowsInserted(from, from+len(proxy)-1)
}

func (m *ProxyModel) Remove(index ...int) {
	if len(index) == 0 {
		return
	}
	i := index[0]
	for i, idx := range index {
		// index must be sorted.
		j := idx - i
		m.beforeRemovePublisher.Publish(j)
		m.items = append(m.items[:j], m.items[j+1:]...)
		m.data.DeleteItem(j)
	}
	m.PublishRowsReset()
	if i <= len(m.items)-1 {
		m.PublishRowsChanged(i, len(m.items)-1)
	}
}

func (m *ProxyModel) Reset(row int) {
	fillProxyRow(m.items[row].Proxy, m.items[row])
	m.PublishRowChanged(row)
	m.PublishRowEdited(row)
}

func (m *ProxyModel) HasName(name string) bool {
	return slices.ContainsFunc(m.items, func(row *ProxyRow) bool {
		return row.Name == name
	})
}

func (m *ProxyModel) BeforeRemove() *walk.IntEvent {
	return m.beforeRemovePublisher.Event()
}

func (m *ProxyModel) RowRenamed() *walk.IntEvent {
	return m.rowRenamedPublisher.Event()
}

func (m *ProxyModel) PublishRowRenamed(row int) {
	m.rowRenamedPublisher.Publish(row)
}

func (m *ProxyModel) RowEdited() *walk.IntEvent {
	return m.rowEditedPublisher.Event()
}

func (m *ProxyModel) PublishRowEdited(row int) {
	m.rowEditedPublisher.Publish(row)
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

// ListEditModel is a list of strings, but supports editing.
type ListEditModel struct {
	walk.ReflectTableModelBase

	values []string
}

func NewListEditModel(values []string) *ListEditModel {
	return &ListEditModel{values: values}
}

func (m *ListEditModel) Value(row, col int) interface{} {
	return &m.values[row]
}

func (m *ListEditModel) RowCount() int {
	return len(m.values)
}

func (m *ListEditModel) Add(value string) {
	m.values = append(m.values, value)
	i := len(m.values) - 1
	m.PublishRowsInserted(i, i)
}

func (m *ListEditModel) Delete(i int) {
	m.values = append(m.values[:i], m.values[i+1:]...)
	m.PublishRowsRemoved(i, i)
}

func (m *ListEditModel) Clear() {
	m.values = nil
	m.PublishRowsReset()
}

func (m *ListEditModel) Move(i, j int) {
	util.MoveSlice(m.values, i, j)
	m.PublishRowsChanged(min(i, j), max(i, j))
}

func (m *ListEditModel) AsString() string {
	if len(m.values) == 0 {
		return ""
	}
	return strings.Join(lo.Filter(m.values, func(item string, index int) bool {
		return strings.TrimSpace(item) != ""
	}), ",")
}
