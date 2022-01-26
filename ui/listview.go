package ui

import (
	"github.com/koho/frpmgr/config"
	"github.com/koho/frpmgr/utils"
	"github.com/lxn/walk"
)

type ConfListModel struct {
	walk.TableModelBase
	walk.SorterBase

	items []*config.Config
}

func NewConfListModel(confList []*config.Config) *ConfListModel {
	m := new(ConfListModel)
	m.items = confList
	return m
}

func (m *ConfListModel) Items() interface{} {
	return m.items
}

type ConfSectionModel struct {
	walk.ReflectTableModelBase
	conf *config.Config
}

func NewConfSectionModel(conf *config.Config) *ConfSectionModel {
	m := new(ConfSectionModel)
	m.conf = conf
	return m
}

func (m *ConfSectionModel) Items() interface{} {
	return m.conf.Items
}

func (m *ConfSectionModel) Count() int {
	return len(m.conf.Items)
}

type LogModel struct {
	walk.ReflectTableModelBase
	items []*struct{ Text string }
	path  string
}

func NewLogModel(path string) *LogModel {
	if path == "" {
		return nil
	}
	m := new(LogModel)
	m.items = make([]*struct{ Text string }, 0)
	m.path = path
	m.Reset()
	return m
}

func (m *LogModel) Items() interface{} {
	return m.items
}

func (m *LogModel) Reset() error {
	lines, err := utils.ReadFileLines(m.path)
	if err != nil {
		return err
	}
	items := make([]*struct{ Text string }, 0)
	for _, line := range lines {
		items = append(items, &struct{ Text string }{line})
	}
	m.items = items
	return nil
}
