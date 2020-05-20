package ui

import (
	"frpmgr/config"
	"frpmgr/utils"
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

type LogModel struct {
	walk.ReflectTableModelBase
	items []*struct{ Text string }
}

func NewLogModel(path string) *LogModel {
	if path == "" {
		return nil
	}
	m := new(LogModel)
	lines, err := utils.ReadFileLines(path)
	if err != nil {
		return nil
	}
	m.items = make([]*struct{ Text string }, 0)
	for _, line := range lines {
		m.items = append(m.items, &struct{ Text string }{line})
	}
	return m
}

func (m *LogModel) Items() interface{} {
	return m.items
}
