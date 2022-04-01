package ui

import (
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/util"
	"github.com/lxn/walk"
	"os"
)

type SortedListModel struct {
	walk.TableModelBase

	items []*Conf
}

func NewSortedListModel(items []*Conf) *SortedListModel {
	m := new(SortedListModel)
	m.items = items
	return m
}

func (m *SortedListModel) Items() interface{} {
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

	conf *Conf
	data *config.ClientConfig
}

func NewProxyModel(conf *Conf) *ProxyModel {
	m := new(ProxyModel)
	m.conf = conf
	m.data = conf.Data.(*config.ClientConfig)
	return m
}

func (m *ProxyModel) Items() interface{} {
	return m.data.Proxies
}

// DefaultListModel has a default item at the top of the model
type DefaultListModel struct {
	Name        string
	DisplayName string
}

func NewDefaultListModel(items []string, defaultKey string, defaultName string) []*DefaultListModel {
	listItems := make([]*DefaultListModel, 0, len(items)+1)
	listItems = append(listItems, &DefaultListModel{Name: defaultKey, DisplayName: defaultName})
	for _, item := range items {
		listItems = append(listItems, &DefaultListModel{Name: item, DisplayName: item})
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
