package ui

import (
	"errors"
	"fmt"
	"frpmgr/config"
	"frpmgr/services"
	"frpmgr/utils"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.org/x/sys/windows"
	"path/filepath"
	"time"
)

type DetailView struct {
	*ConfStatusView
	*ConfSectionView
}

func NewDetailView() *DetailView {
	v := new(DetailView)
	v.ConfStatusView = NewConfStatusView()
	v.ConfSectionView = NewConfSectionView()
	return v
}

func (t *DetailView) SetConf(conf *config.Config) {
	if conf == nil {
		t.ConfSectionView.SetModel(nil)
		t.ConfStatusView.SetConf(conf)
	} else {
		t.ConfSectionView.SetModel(conf)
		t.ConfStatusView.SetConf(conf)
		if lastRunningState {
			t.toggleService(true)
			lastRunningState = false
		}
	}
}

func (t *DetailView) toggleService(restart bool) {
	var err error
	t.toggle.SetEnabled(false)
	confPath, err := config.PathFromName(t.conf.Name)
	if err != nil {
		return
	}
	utils.EnsurePath(t.conf.LogFile)
	if t.running {
		t.statusImage.SetImage(iconForState(config.StateStopping, 14))
		t.status.SetText("正在停止")
		err = services.UninstallService(config.NameFromPath(confPath))
		if restart {
			t.statusImage.SetImage(iconForState(config.StateStarting, 14))
			t.status.SetText("正在启动")
			err = services.InstallService(confPath, t.conf.ManualStart)
		}
	} else {
		t.statusImage.SetImage(iconForState(config.StateStarting, 14))
		t.status.SetText("正在启动")
		err = services.InstallService(confPath, t.conf.ManualStart)
	}
	if err != nil {
		walk.MsgBox(t.view.Form(), "错误", "操作服务失败\n\n"+err.Error(), walk.MsgBoxOK|walk.MsgBoxIconError)
	}
}

func (t *DetailView) Initialize() {
	t.ConfStatusView.Initialize()
	t.toggle.Clicked().Attach(func() {
		t.toggleService(false)
	})
	t.ConfSectionView.toggleService = t.toggleService
	t.ConfSectionView.editToolBar.ApplyDPI(t.view.DPI())
}

type ConfSectionView struct {
	model *ConfSectionModel

	editToolBar   *walk.ToolBar
	sectionView   *walk.TableView
	newAction     *walk.Action
	rdAction      *walk.Action
	sshAction     *walk.Action
	webAction     *walk.Action
	editAction    *walk.Action
	deleteAction  *walk.Action
	toggleService func(bool)
}

var lastEditSection = -1

func NewConfSectionView() *ConfSectionView {
	csv := new(ConfSectionView)
	return csv
}

func (t *ConfSectionView) SetModel(conf *config.Config) {
	if conf == nil {
		t.model = nil
		t.sectionView.SetModel(nil)
	} else {
		t.model = NewConfSectionModel(conf)
		t.sectionView.SetModel(t.model)
	}
}

func (t *ConfSectionView) ResetModel() {
	t.sectionView.SetModel(t.model)
	if lastEditSection >= 0 && t.model != nil && t.model.Count() > 0 {
		t.sectionView.SetCurrentIndex(lastEditSection)
		lastEditSection = -1
	}
}

func (t *ConfSectionView) mustSelectConf() bool {
	return t.model != nil && t.model.conf != nil
}

func (t *ConfSectionView) onQuickAddSection(ssd *SimpleSection) {
	if !t.mustSelectConf() {
		return
	}
	if res, _ := ssd.Run(t.sectionView.Form()); res == walk.DlgCmdOK {
		t.model.conf.Items = append(t.model.conf.Items, ssd.sections...)
		t.ResetModel()
		if err := t.model.conf.Save(); err != nil {
			walk.MsgBox(t.sectionView.Form(), "错误", "写入失败", walk.MsgBoxOK|walk.MsgBoxIconError)
			return
		}
		if running, _ := services.QueryService(t.model.conf.Name); running {
			t.toggleService(true)
		}
	}
}

func (t *ConfSectionView) onDeleteSection() {
	index := t.sectionView.CurrentIndex()
	if index < 0 {
		return
	}
	if walk.MsgBox(t.sectionView.Form(), fmt.Sprintf("删除项目「%s」", t.model.conf.Items[index].Name), fmt.Sprintf("确定要删除项目「%s」吗?", t.model.conf.Items[index].Name), walk.MsgBoxOKCancel|walk.MsgBoxIconWarning) == walk.DlgCmdCancel {
		return
	}
	t.model.conf.Items = append(t.model.conf.Items[:index], t.model.conf.Items[index+1:]...)
	t.ResetModel()
	t.model.conf.Save()
	if running, _ := services.QueryService(t.model.conf.Name); running {
		t.toggleService(true)
	}
}

func (t *ConfSectionView) onEditSection(edit bool) {
	if !t.mustSelectConf() {
		return
	}
	var ret int
	if edit {
		index := t.sectionView.CurrentIndex()
		if index < 0 {
			return
		}
		ret, _ = NewEditSectionDialog(t.model.conf.Items[index]).Run(t.sectionView.Form())
	} else {
		esd := NewEditSectionDialog(nil)
		if ret, _ = esd.Run(t.sectionView.Form()); ret == walk.DlgCmdOK {
			t.model.conf.Items = append(t.model.conf.Items, esd.section)
		}
	}
	if ret == walk.DlgCmdOK {
		if edit {
			lastEditSection = t.sectionView.CurrentIndex()
		}
		t.ResetModel()
		t.model.conf.Save()
		if running, _ := services.QueryService(t.model.conf.Name); running {
			t.toggleService(true)
		}
	}
}

func (t *ConfSectionView) View() Widget {
	var sectionDataBinder *walk.DataBinder
	return Composite{
		DataBinder: DataBinder{DataSource: &struct {
			Selected func() bool
		}{func() bool {
			return t.sectionView != nil && t.sectionView.CurrentIndex() >= 0
		}}, Name: "section", AssignTo: &sectionDataBinder},
		Layout: VBox{MarginsZero: true},
		Children: []Widget{
			ToolBar{
				AssignTo:    &t.editToolBar,
				ButtonStyle: ToolBarButtonImageBeforeText,
				Orientation: Horizontal,
				Items: []MenuItem{
					Action{
						AssignTo: &t.newAction,
						Text:     "添加",
						Image:    loadSysIcon("shell32", 205, 16),
						OnTriggered: func() {
							t.onEditSection(false)
						},
					},
					Menu{
						Text:  "快速添加",
						Image: loadSysIcon("imageres", 111, 16),
						Items: []MenuItem{
							Action{
								AssignTo: &t.rdAction,
								Text:     "远程桌面",
								Image:    loadSysIcon("imageres", 20, 16),
								OnTriggered: func() {
									t.onQuickAddSection(NewSimpleSectionDialog("远程桌面", "rdp", []string{"tcp", "udp"}, 3389))
								},
							},
							Action{
								AssignTo: &t.sshAction,
								Text:     "SSH",
								Image:    loadSysIcon("shell32", 94, 16),
								OnTriggered: func() {
									t.onQuickAddSection(NewSimpleSectionDialog("SSH", "ssh", []string{"tcp"}, 22))
								},
							},
							Action{
								AssignTo: &t.webAction,
								Text:     "Web",
								Image:    loadSysIcon("shell32", 13, 16),
								OnTriggered: func() {
									t.onQuickAddSection(NewSimpleSectionDialog("Web", "web", []string{"tcp"}, 80))
								},
							},
						},
					},
					Action{
						AssignTo: &t.editAction,
						Image:    loadSysIcon("shell32", 269, 16),
						Text:     "编辑",
						Enabled:  Bind("section.Selected"),
						OnTriggered: func() {
							t.onEditSection(true)
						},
					},
					Action{
						AssignTo:    &t.deleteAction,
						Image:       loadSysIcon("shell32", 131, 16),
						Text:        "删除",
						Enabled:     Bind("section.Selected"),
						OnTriggered: t.onDeleteSection,
					},
					Action{
						Image: loadResourceIcon(22, 16),
						Text:  "打开配置文件",
						OnTriggered: func() {
							if !t.mustSelectConf() {
								return
							}
							if path, err := filepath.Abs(t.model.conf.Path); err == nil {
								openPath(path)
							}
						},
					},
				},
			},
			TableView{
				AssignTo: &t.sectionView,
				Columns: []TableViewColumn{
					{Title: "名称", DataMember: "Name"},
					{Title: "类型", DataMember: "Type"},
					{Title: "本地地址", DataMember: "LocalIP"},
					{Title: "本地端口", DataMember: "LocalPort"},
					{Title: "远程端口", DataMember: "RemotePort"},
				},
				ContextMenuItems: []MenuItem{
					ActionRef{&t.editAction},
					ActionRef{&t.newAction},
					Menu{
						Text:  "快速添加",
						Image: loadSysIcon("imageres", 111, 16),
						Items: []MenuItem{
							ActionRef{&t.rdAction},
							ActionRef{&t.sshAction},
							ActionRef{&t.webAction},
						},
					},
					ActionRef{&t.deleteAction},
				},
				OnCurrentIndexChanged: func() {
					sectionDataBinder.Reset()
				},
				OnItemActivated: func() {
					t.onEditSection(true)
				},
			},
		},
	}
}

type ConfStatusView struct {
	view *walk.GroupBox
	conf *config.Config

	nameChan    chan string
	running     bool
	status      *walk.Label
	statusImage *walk.ImageView
	address     *walk.Label
	toggle      *walk.PushButton
	svcOpen     *walk.PushButton
}

func NewConfStatusView() *ConfStatusView {
	v := new(ConfStatusView)
	v.nameChan = make(chan string)
	return v
}

func (t *ConfStatusView) SetConf(conf *config.Config) {
	t.conf = conf
	if conf == nil {
		t.nameChan <- ""
	} else {
		t.nameChan <- t.conf.Name
	}
}

func (t *ConfStatusView) queryState(name string) (bool, error) {
	if name == "" {
		return false, nil
	}
	return services.QueryService(name)
}

func (t *ConfStatusView) UpdateStatus(name string, running bool, err error) {
	t.running = running
	if name == "" || t.conf == nil {
		t.view.SetTitle("")
		t.status.SetText("-")
		t.address.SetText("-")
		t.toggle.SetEnabled(false)
		t.toggle.SetText("启动")
		t.statusImage.SetVisible(false)
		return
	}
	t.view.SetTitle(name)
	t.address.SetText(t.conf.ServerAddress)
	t.toggle.SetEnabled(true)
	t.svcOpen.SetEnabled(true)
	t.statusImage.SetVisible(true)
	if running {
		t.status.SetText("正在运行")
		t.toggle.SetText("停止")
		t.statusImage.SetImage(iconForState(config.StateStarted, 14))
	} else {
		t.status.SetText("已停止")
		t.toggle.SetText("启动")
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			t.svcOpen.SetEnabled(false)
		}
		t.statusImage.SetImage(iconForState(config.StateStopped, 14))
	}
}

func (t *ConfStatusView) View() Widget {
	return GroupBox{
		AssignTo: &t.view,
		Title:    "",
		Layout:   Grid{Margins: Margins{10, 5, 10, 5}, Spacing: 0},
		Children: []Widget{
			Composite{
				Layout:    HBox{MarginsZero: true, SpacingZero: true},
				Row:       0,
				Column:    0,
				Alignment: AlignHFarVFar,
				Children: []Widget{
					Label{Text: "状态:"},
				},
			},
			Composite{
				Layout:    HBox{MarginsZero: true, SpacingZero: true},
				Row:       1,
				Column:    0,
				Alignment: AlignHFarVFar,
				Children: []Widget{
					Label{Text: "远程地址:"},
				},
			},
			Composite{
				Layout: HBox{SpacingZero: true, MarginsZero: true},
				Row:    0, Column: 1,
				Children: []Widget{
					ImageView{
						AssignTo: &t.statusImage,
						Visible:  false,
						Margin:   2,
					},
					Label{AssignTo: &t.status, Text: "-", TextAlignment: Alignment1D(walk.AlignHNearVNear)},
				},
			},
			Label{AssignTo: &t.address, Text: "-", Row: 1, Column: 1, TextAlignment: Alignment1D(walk.AlignHNearVNear)},
			PushButton{AssignTo: &t.toggle, Text: "启动", Alignment: AlignHNearVNear,
				MaxSize: Size{80, 0}, Row: 2, Column: 1, Enabled: false},
			PushButton{AssignTo: &t.svcOpen, Text: "查看服务", Alignment: AlignHNearVNear,
				MaxSize: Size{80, 0}, Row: 2, Column: 2, Enabled: false, OnClicked: func() {
					services.ShowPropertyDialog("FRP Client: " + t.view.Title())
				}},
		},
	}
}

func (t *ConfStatusView) Initialize() {
	ticker := time.NewTicker(time.Second * 1)
	var name = ""
	var onTick = func() {
		state, err := t.queryState(name)
		t.view.Synchronize(func() {
			t.UpdateStatus(name, state, err)
		})
	}
	go func() {
		defer ticker.Stop()
		for {
			select {
			case name = <-t.nameChan:
				onTick()
			case <-ticker.C:
				onTick()
			}
		}
	}()
}
