package ui

import (
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/services"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"os"
	"path/filepath"
)

var stateDescription = map[consts.ServiceState]string{
	consts.StateUnknown:  "未知",
	consts.StateStarted:  "正在运行",
	consts.StateStopped:  "已停止",
	consts.StateStarting: "正在启动",
	consts.StateStopping: "正在停止",
}

type PanelView struct {
	*walk.GroupBox

	stateText   *walk.Label
	stateImage  *walk.ImageView
	addressText *walk.Label
	toggleBtn   *walk.PushButton
	svcOpenBtn  *walk.PushButton
	copyImage   *walk.ImageView
}

func NewPanelView() *PanelView {
	return new(PanelView)
}

func (pv *PanelView) View() Widget {
	return GroupBox{
		AssignTo: &pv.GroupBox,
		Title:    "",
		Layout:   Grid{Margins: Margins{10, 5, 10, 9}, Spacing: 0},
		Children: []Widget{
			Composite{
				Layout:    HBox{MarginsZero: true, SpacingZero: true},
				Row:       0,
				Column:    0,
				Alignment: AlignHFarVCenter,
				Children: []Widget{
					Label{Text: "状态:"},
				},
			},
			Composite{
				Layout:    HBox{MarginsZero: true, SpacingZero: true},
				Row:       1,
				Column:    0,
				Alignment: AlignHFarVCenter,
				Children: []Widget{
					Label{Text: "远程地址:"},
				},
			},
			Composite{
				Layout: HBox{SpacingZero: true, MarginsZero: true},
				Row:    0, Column: 1,
				Alignment: AlignHNearVCenter,
				Children: []Widget{
					ImageView{
						AssignTo: &pv.stateImage,
						Visible:  false,
						Margin:   0,
					},
					HSpacer{Size: 4},
					Label{AssignTo: &pv.stateText, Text: "-", TextAlignment: Alignment1D(walk.AlignHNearVCenter)},
				},
			},
			Composite{
				Layout: HBox{SpacingZero: true, MarginsZero: true},
				Row:    1, Column: 1,
				Alignment: AlignHNearVCenter,
				Children: []Widget{
					Label{AssignTo: &pv.addressText, Text: "-"},
					HSpacer{Size: 5},
					ImageView{AssignTo: &pv.copyImage, Image: loadResourceIcon(consts.IconCopy, 16), ToolTipText: "复制",
						OnMouseDown: func(x, y int, button walk.MouseButton) {
							if button == walk.LeftButton {
								pv.copyImage.SetImage(loadResourceIcon(consts.IconCopyActive, 16))
							}
						}, OnMouseUp: func(x, y int, button walk.MouseButton) {
							if button == walk.LeftButton {
								pv.copyImage.SetImage(loadResourceIcon(consts.IconCopy, 16))
								walk.Clipboard().SetText(pv.addressText.Text())
							}
						}},
				},
			},
			Composite{
				Layout: HBox{MarginsZero: true},
				Row:    2, Column: 1,
				Alignment: AlignHNearVCenter,
				Children: []Widget{
					PushButton{AssignTo: &pv.toggleBtn, Text: "启动", MaxSize: Size{80, 0}, Enabled: false, OnClicked: pv.ToggleService},
					PushButton{AssignTo: &pv.svcOpenBtn, Text: "查看服务", MaxSize: Size{80, 0}, Enabled: false,
						OnClicked: func() {
							services.ShowPropertyDialog("FRP Client: " + pv.Title())
						}},
					HSpacer{},
				},
			},
		},
	}
}

func (pv *PanelView) OnCreate() {

}

func (pv *PanelView) setState(state consts.ServiceState) {
	pv.stateImage.SetImage(iconForState(state, 14))
	pv.stateText.SetText(stateDescription[state])
}

func (pv *PanelView) ToggleService() {
	conf := getCurrentConf()
	if conf == nil {
		return
	}
	var err error
	if conf.State == consts.StateStarted {
		err = pv.StopService(conf)
	} else {
		err = pv.StartService(conf)
	}
	if err != nil {
		showError(err, pv.Form())
	}
}

// StartService creates a daemon service of the given config, then starts it
func (pv *PanelView) StartService(conf *Conf) error {
	// Verify the config file
	if err := services.VerifyClientConfig(conf.Path); err != nil {
		return err
	}
	// Ensure log directory is valid
	if logFile := conf.Data.GetLogFile(); logFile != "" && logFile != "console" {
		if err := os.MkdirAll(filepath.Dir(logFile), os.ModePerm); err != nil {
			return err
		}
	}
	pv.toggleBtn.SetEnabled(false)
	pv.setState(consts.StateStarting)
	conf.Lock()
	conf.State = consts.StateStarting
	conf.Unlock()
	return services.InstallService(conf.Name, conf.Path, !conf.Data.AutoStart())
}

// StopService stops the service of the given config, then removes it
func (pv *PanelView) StopService(conf *Conf) error {
	pv.toggleBtn.SetEnabled(false)
	pv.setState(consts.StateStopping)
	return services.UninstallService(conf.Name, false)
}

// Invalidate updates views using the current config
func (pv *PanelView) Invalidate() {
	conf := getCurrentConf()
	if conf == nil {
		pv.SetTitle("")
		pv.stateText.SetText("")
		pv.addressText.SetText("")
		pv.toggleBtn.SetEnabled(false)
		pv.toggleBtn.SetText("启动")
		pv.svcOpenBtn.SetEnabled(false)
		pv.stateImage.SetVisible(false)
		return
	}
	data := conf.Data.(*config.ClientConfig)
	if pv.Title() != conf.Name {
		pv.SetTitle(conf.Name)
	}
	addr := data.ServerAddress
	if addr == "" {
		addr = "0.0.0.0"
	}
	if pv.addressText.Text() != addr {
		pv.addressText.SetText(addr)
	}
	pv.toggleBtn.SetEnabled(true)
	pv.svcOpenBtn.SetEnabled(conf.Install)
	pv.stateImage.SetVisible(true)
	if conf.State == consts.StateStarted {
		pv.setState(consts.StateStarted)
		pv.toggleBtn.SetText("停止")
	} else {
		pv.setState(consts.StateStopped)
		pv.toggleBtn.SetText("启动")
	}
}
