package ui

import (
	"fmt"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"path/filepath"
)

type SectionView struct {
	*walk.Composite

	model   *SectionModel
	toolbar *walk.ToolBar
	table   *walk.TableView

	// Actions
	newAction    *walk.Action
	rdAction     *walk.Action
	sshAction    *walk.Action
	webAction    *walk.Action
	editAction   *walk.Action
	deleteAction *walk.Action
}

func NewSectionView() *SectionView {
	return new(SectionView)
}

func (sv *SectionView) View() Widget {
	var sectionDB *walk.DataBinder
	return Composite{
		AssignTo: &sv.Composite,
		DataBinder: DataBinder{
			AssignTo: &sectionDB,
			Name:     "section",
			DataSource: &struct{ Selected func() bool }{
				func() bool { return sv.table != nil && sv.table.CurrentIndex() >= 0 },
			},
		},
		Layout: VBox{MarginsZero: true},
		Children: []Widget{
			ToolBar{
				AssignTo:    &sv.toolbar,
				ButtonStyle: ToolBarButtonImageBeforeText,
				Orientation: Horizontal,
				Items: []MenuItem{
					Action{
						AssignTo: &sv.newAction,
						Text:     "添加",
						Image:    loadSysIcon("shell32", consts.IconCreate, 16),
						OnTriggered: func() {
							sv.onEdit(false)
						},
					},
					Menu{
						Text:  "快速添加",
						Image: loadSysIcon("imageres", consts.IconQuickAdd, 16),
						Items: []MenuItem{
							Action{
								AssignTo: &sv.rdAction,
								Text:     "远程桌面",
								Image:    loadSysIcon("imageres", consts.IconRemote, 16),
								OnTriggered: func() {
									sv.onQuickAdd(NewSimpleProxyDialog("远程桌面", "rdp", []string{"tcp", "udp"}, 3389))
								},
							},
							Action{
								AssignTo: &sv.sshAction,
								Text:     "SSH",
								Image:    loadSysIcon("shell32", consts.IconComputer, 16),
								OnTriggered: func() {
									sv.onQuickAdd(NewSimpleProxyDialog("SSH", "ssh", []string{"tcp"}, 22))
								},
							},
							Action{
								AssignTo: &sv.webAction,
								Text:     "Web",
								Image:    loadSysIcon("shell32", consts.IconWeb, 16),
								OnTriggered: func() {
									sv.onQuickAdd(NewSimpleProxyDialog("Web", "web", []string{"tcp"}, 80))
								},
							},
						},
					},
					Action{
						AssignTo: &sv.editAction,
						Image:    loadSysIcon("shell32", consts.IconEdit, 16),
						Text:     "编辑",
						Enabled:  Bind("section.Selected"),
						OnTriggered: func() {
							sv.onEdit(true)
						},
					},
					Action{
						AssignTo:    &sv.deleteAction,
						Image:       loadSysIcon("shell32", consts.IconDelete, 16),
						Text:        "删除",
						Enabled:     Bind("section.Selected"),
						OnTriggered: sv.onDelete,
					},
					Action{
						Image: loadResourceIcon(consts.IconOpen, 16),
						Text:  "打开配置文件",
						OnTriggered: func() {
							if sv.model == nil {
								return
							}
							if path, err := filepath.Abs(sv.model.conf.Path); err == nil {
								openPath(path)
							}
						},
					},
				},
			},
			TableView{
				AssignTo: &sv.table,
				Columns: []TableViewColumn{
					{Title: "名称", DataMember: "Name", Width: 105},
					{Title: "类型", DataMember: "Type", Width: 60},
					{Title: "本地地址", DataMember: "LocalIP", Width: 110},
					{Title: "本地端口", DataMember: "LocalPort", Width: 90},
					{Title: "远程端口", DataMember: "RemotePort", Width: 90},
					{Title: "子域名", DataMember: "SubDomain", Width: 90},
					{Title: "自定义域名", DataMember: "CustomDomains", Width: 90},
					{Title: "插件", DataMember: "Plugin", Width: 100},
				},
				ContextMenuItems: []MenuItem{
					ActionRef{&sv.editAction},
					ActionRef{&sv.newAction},
					Menu{
						Text:  "快速添加",
						Image: loadSysIcon("imageres", consts.IconQuickAdd, 16),
						Items: []MenuItem{
							ActionRef{&sv.rdAction},
							ActionRef{&sv.sshAction},
							ActionRef{&sv.webAction},
						},
					},
					ActionRef{&sv.deleteAction},
				},
				OnCurrentIndexChanged: func() {
					sectionDB.Reset()
				},
				OnItemActivated: func() {
					sv.onEdit(true)
				},
			},
		},
	}
}

func (sv *SectionView) OnCreate() {

}

func (sv *SectionView) Invalidate() {
	if conf := getCurrentConf(); conf == nil {
		sv.model = nil
		sv.table.SetModel(nil)
	} else {
		sv.model = NewSectionModel(conf)
		sv.table.SetModel(sv.model)
	}
}

func (sv *SectionView) onDelete() {
	if sv.model == nil {
		return
	}
	index := sv.table.CurrentIndex()
	if index < 0 {
		return
	}
	section, ok := sv.model.conf.Data.ItemAt(index).(config.Section)
	if !ok {
		return
	}
	if walk.MsgBox(sv.Form(), fmt.Sprintf("删除项目「%s」", section.GetName()),
		fmt.Sprintf("确定要删除项目「%s」吗?", section.GetName()),
		walk.MsgBoxOKCancel|walk.MsgBoxIconWarning) == walk.DlgCmdCancel {
		return
	}
	sv.model.conf.Data.DeleteItem(index)
	sv.commit()
}

func (sv *SectionView) onEdit(edit bool) {
	if sv.model == nil {
		return
	}
	var ret int
	if edit {
		index := sv.table.CurrentIndex()
		if index < 0 {
			return
		}
		if ret, _ = NewEditProxyDialog(sv.model.conf.Data.ItemAt(index).(*config.Proxy)).Run(sv.Form()); ret == walk.DlgCmdOK {
			sv.commit()
			sv.table.SetCurrentIndex(index)
		}
	} else {
		esd := NewEditProxyDialog(nil)
		if ret, _ = esd.Run(sv.Form()); ret == walk.DlgCmdOK {
			sv.model.conf.Data.AddItem(esd.Proxy)
			sv.commit()
		}
	}
}

func (sv *SectionView) onQuickAdd(spd *SimpleProxyDialog) {
	if sv.model == nil {
		return
	}
	if res, _ := spd.Run(sv.Form()); res == walk.DlgCmdOK {
		for _, proxy := range spd.Proxies {
			if !sv.model.conf.Data.AddItem(proxy) {
				showWarningMessage(sv.Form(), "代理已存在", fmt.Sprintf("代理名「%s」已存在。", proxy.Name))
			}
		}
		sv.commit()
	}
}

// commit will update the views and save the config to disk, then reload service
func (sv *SectionView) commit() {
	sv.Invalidate()
	commitConf(sv.model.conf, false)
}
