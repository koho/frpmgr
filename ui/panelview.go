package ui

import (
	"os"
	"path/filepath"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/res"
	"github.com/koho/frpmgr/pkg/util"
	"github.com/koho/frpmgr/services"
)

var stateDescription = map[consts.ServiceState]string{
	consts.StateUnknown:  i18n.Sprintf("Unknown"),
	consts.StateStarted:  i18n.Sprintf("Running"),
	consts.StateStopped:  i18n.Sprintf("Stopped"),
	consts.StateStarting: i18n.Sprintf("Starting"),
	consts.StateStopping: i18n.Sprintf("Stopping"),
}

type PanelView struct {
	*walk.GroupBox

	stateText   *walk.Label
	stateImage  *walk.ImageView
	addressText *walk.Label
	toggleBtn   *walk.PushButton
}

func NewPanelView() *PanelView {
	return new(PanelView)
}

func (pv *PanelView) View() Widget {
	var cpIcon *walk.CustomWidget
	cpIconColor := res.ColorDarkGray
	setCopyIconColor := func(button walk.MouseButton, color walk.Color) {
		if button == walk.LeftButton {
			cpIconColor = color
			cpIcon.Invalidate()
		}
	}
	return GroupBox{
		AssignTo: &pv.GroupBox,
		Title:    "",
		Layout:   Grid{Margins: Margins{Left: 10, Top: 10, Right: 10, Bottom: 10}, Spacing: 10},
		Children: []Widget{
			Label{Text: i18n.SprintfColon("Status"), Row: 0, Column: 0, Alignment: AlignHFarVCenter},
			Label{Text: i18n.SprintfColon("Remote Address"), Row: 1, Column: 0, Alignment: AlignHFarVCenter},
			Composite{
				Layout: HBox{SpacingZero: true, MarginsZero: true},
				Row:    0, Column: 1,
				Alignment: AlignHNearVCenter,
				Children: []Widget{
					ImageView{AssignTo: &pv.stateImage, Margin: 0},
					HSpacer{Size: 4},
					Label{AssignTo: &pv.stateText},
				},
			},
			Composite{
				Layout: HBox{SpacingZero: true, MarginsZero: true},
				Row:    1, Column: 1,
				Alignment: AlignHNearVCenter,
				Children: []Widget{
					Label{AssignTo: &pv.addressText},
					HSpacer{Size: 5},
					CustomWidget{
						AssignTo:            &cpIcon,
						Background:          TransparentBrush{},
						ClearsBackground:    true,
						InvalidatesOnResize: true,
						MinSize:             Size{Width: 16, Height: 16},
						ToolTipText:         i18n.Sprintf("Copy"),
						PaintPixels: func(canvas *walk.Canvas, updateBounds walk.Rectangle) error {
							return drawCopyIcon(canvas, cpIconColor)
						},
						OnMouseDown: func(x, y int, button walk.MouseButton) {
							setCopyIconColor(button, res.ColorLightBlue)
						},
						OnMouseUp: func(x, y int, button walk.MouseButton) {
							setCopyIconColor(button, res.ColorDarkGray)
							bounds := cpIcon.ClientBoundsPixels()
							if x >= 0 && x <= bounds.Right() && y >= 0 && y <= bounds.Bottom() {
								walk.Clipboard().SetText(pv.addressText.Text())
							}
						},
					},
					VSpacer{Size: 20},
				},
			},
			Composite{
				Layout: HBox{MarginsZero: true},
				Row:    2, Column: 1,
				Alignment: AlignHNearVCenter,
				Children: []Widget{
					PushButton{
						AssignTo:  &pv.toggleBtn,
						Text:      i18n.Sprintf("Start"),
						MaxSize:   Size{Width: 80},
						Enabled:   false,
						OnClicked: pv.ToggleService,
					},
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
		if walk.MsgBox(pv.Form(), i18n.Sprintf("Stop config \"%s\"", conf.Name()),
			i18n.Sprintf("Are you sure you would like to stop config \"%s\"?", conf.Name()),
			walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) == walk.DlgCmdNo {
			return
		}
		err = pv.StopService(conf)
	} else {
		if !util.FileExists(conf.Path) {
			warnConfigRemoved(pv.Form(), conf.Name())
			return
		}
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
	return services.InstallService(conf.Name(), conf.Path, !conf.Data.AutoStart())
}

// StopService stops the service of the given config, then removes it
func (pv *PanelView) StopService(conf *Conf) error {
	pv.toggleBtn.SetEnabled(false)
	pv.setState(consts.StateStopping)
	return services.UninstallService(conf.Path, false)
}

// Invalidate updates views using the current config
func (pv *PanelView) Invalidate() {
	conf := getCurrentConf()
	if conf == nil {
		pv.SetTitle("")
		pv.setState(consts.StateUnknown)
		pv.addressText.SetText("")
		pv.toggleBtn.SetEnabled(false)
		pv.toggleBtn.SetText(i18n.Sprintf("Start"))
		return
	}
	data := conf.Data.(*config.ClientConfig)
	if pv.Title() != conf.Name() {
		pv.SetTitle(conf.Name())
	}
	addr := data.ServerAddress
	if addr == "" {
		addr = "0.0.0.0"
	}
	if pv.addressText.Text() != addr {
		pv.addressText.SetText(addr)
	}
	pv.toggleBtn.SetEnabled(true)
	if conf.State == consts.StateStarted {
		pv.setState(consts.StateStarted)
		pv.toggleBtn.SetText(i18n.Sprintf("Stop"))
	} else {
		pv.setState(consts.StateStopped)
		pv.toggleBtn.SetText(i18n.Sprintf("Start"))
	}
}
