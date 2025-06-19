package ui

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.org/x/sys/windows"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/res"
	"github.com/koho/frpmgr/pkg/util"
	"github.com/koho/frpmgr/services"
)

type PropertiesDialog struct {
	*walk.Dialog

	table *walk.TableView
	conf  *Conf
}

func NewPropertiesDialog(conf *Conf) *PropertiesDialog {
	return &PropertiesDialog{conf: conf}
}

func (pd *PropertiesDialog) logFileStat() (count int, size int64) {
	if logs, _, err := util.FindLogFiles(pd.conf.Data.GetLogFile()); err == nil {
		for _, logFile := range logs {
			if fileInfo, err := os.Stat(logFile); err == nil {
				count++
				size += fileInfo.Size()
			}
		}
	}
	return
}

func (pd *PropertiesDialog) Run(owner walk.Form) (int, error) {
	logFileCount, logFileSize := pd.logFileStat()
	logSizeDesc := util.ByteCountIEC(logFileSize)

	var startTypeDesc string
	startType, pid, _ := services.QueryStartInfo(pd.conf.Path)
	switch startType {
	case windows.SERVICE_AUTO_START:
		startTypeDesc = i18n.Sprintf("Auto")
	case windows.SERVICE_DEMAND_START:
		startTypeDesc = i18n.Sprintf("Manual")
	default:
		startTypeDesc = i18n.Sprintf("None")
	}
	items := []*ListItem{
		{Title: i18n.Sprintf("Name"), Value: pd.conf.Name()},
		{Title: i18n.Sprintf("Identifier"), Value: util.FileNameWithoutExt(pd.conf.Path)},
		{Title: i18n.Sprintf("Service Name"), Value: services.ServiceNameOfClient(pd.conf.Path)},
		{Title: i18n.Sprintf("File Format"), Value: strings.ToUpper(pd.conf.Data.Ext()[1:])},
		{Title: i18n.Sprintf("Number of Proxies"), Value: strconv.Itoa(reflect.ValueOf(pd.conf.Data.Items()).Len())},
		{Title: i18n.Sprintf("Start Type"), Value: startTypeDesc},
		{Title: i18n.Sprintf("Log"), Value: i18n.Sprintf("%d Files, %s", logFileCount, logSizeDesc)},
	}
	if info, err := os.Stat(pd.conf.Path); err == nil {
		created := time.Unix(0, info.Sys().(*syscall.Win32FileAttributeData).CreationTime.Nanoseconds())
		modified := info.ModTime()
		items = append(items, &ListItem{
			Title: i18n.Sprintf("Created"),
			Value: created.Format(time.DateTime),
		}, &ListItem{
			Title: i18n.Sprintf("Modified"),
			Value: modified.Format(time.DateTime),
		})
	}
	if pid > 0 {
		if process, err := syscall.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid); err == nil {
			var creationTime, unusedTime syscall.Filetime
			if err = syscall.GetProcessTimes(process, &creationTime, &unusedTime, &unusedTime, &unusedTime); err == nil {
				items = append(items, &ListItem{
					Title: i18n.Sprintf("Started"),
					Value: time.Unix(0, creationTime.Nanoseconds()).Format(time.DateTime),
				})
			}
			syscall.CloseHandle(process)
		}
		items = append(items, &ListItem{
			Title: i18n.Sprintf("Number of TCP Connections"),
			Value: strconv.Itoa(util.CountTCPConnections(pid)),
		}, &ListItem{
			Title: i18n.Sprintf("Number of UDP Connections"),
			Value: strconv.Itoa(util.CountUDPConnections(pid)),
		})
	}
	dlg := NewBasicDialog(&pd.Dialog, i18n.Sprintf("%s Properties", pd.conf.Name()),
		loadIcon(res.IconFile, 32),
		DataBinder{}, nil,
		TableView{
			AssignTo: &pd.table,
			Name:     "properties",
			Columns: []TableViewColumn{
				{Title: i18n.Sprintf("Item"), DataMember: "Title"},
				{Title: i18n.Sprintf("Value"), DataMember: "Value", Width: 180},
			},
			ColumnsOrderable: false,
			Model:            NewNonSortedModel(items),
			OnBoundsChanged: func() {
				pd.table.FitColumn(0, 140)
			},
			ContextMenuItems: []MenuItem{
				Action{
					Text:    i18n.Sprintf("Copy Value"),
					Enabled: Bind("properties.CurrentIndex >= 0"),
					Visible: Bind("properties.CurrentIndex >= 0"),
					OnTriggered: func() {
						if idx := pd.table.CurrentIndex(); idx >= 0 && idx < len(items) {
							walk.Clipboard().SetText(items[idx].Value)
						}
					},
				},
			},
		},
	)
	dlg.MinSize = Size{Width: 400, Height: 350}
	return dlg.Run(owner)
}
