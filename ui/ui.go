package ui

import (
	"frpmgr/config"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"os"
)

type FRPManager struct {
	window *walk.MainWindow
	tabs   *walk.TabWidget

	confPage *ConfPage
	logPage  *LogPage
}

var curDir string

func RunUI() {
	fm := new(FRPManager)
	var err error
	config.Configurations, err = config.LoadConfig()
	if err != nil {
		return
	}
	fm.confPage = NewConfPage()
	fm.logPage = NewLogPage()

	mw := MainWindow{
		AssignTo:   &fm.window,
		Title:      "FRP 管理器",
		Persistent: true,
		MinSize:    Size{500, 400},
		Size:       Size{800, 525},
		Layout:     VBox{Margins: Margins{5, 5, 5, 5}},
		Font:       Font{Family: "微软雅黑", PointSize: 9},
		Children: []Widget{
			TabWidget{
				AssignTo: &fm.tabs,
				Pages: []TabPage{
					fm.confPage.View(),
					fm.logPage.View(),
				},
			},
		},
	}
	if err := mw.Create(); err != nil {
		panic(err)
		return
	}
	fm.confPage.Initialize()
	fm.logPage.Initialize()
	curDir, _ = os.Getwd()
	fm.window.Run()
}
