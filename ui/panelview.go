package ui

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/res"
	"github.com/koho/frpmgr/pkg/util"
	"github.com/koho/frpmgr/services"
)

var configStateDescription = map[consts.ConfigState]string{
	consts.ConfigStateUnknown:  i18n.Sprintf("Unknown"),
	consts.ConfigStateStarted:  i18n.Sprintf("Running"),
	consts.ConfigStateStopped:  i18n.Sprintf("Stopped"),
	consts.ConfigStateStarting: i18n.Sprintf("Starting"),
	consts.ConfigStateStopping: i18n.Sprintf("Stopping"),
}

type PanelView struct {
	*walk.GroupBox

	stateText   *walk.Label
	stateImage  *walk.ImageView
	addressText *walk.Label
	protoText   *walk.Label
	protoImage  *walk.ImageView
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
			Label{Text: i18n.SprintfColon("Protocol"), Row: 2, Column: 0, Alignment: AlignHFarVCenter},
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
				Layout: HBox{Spacing: 2, MarginsZero: true},
				Row:    2, Column: 1,
				Alignment: AlignHNearVCenter,
				Children: []Widget{
					ImageView{
						AssignTo:    &pv.protoImage,
						Image:       loadIcon(res.IconFlatLock, 14),
						ToolTipText: i18n.Sprintf("Your connection to the server is encrypted"),
					},
					Label{AssignTo: &pv.protoText},
				},
			},
			Composite{
				Layout: HBox{MarginsZero: true},
				Row:    3, Column: 1,
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

func (pv *PanelView) setState(state consts.ConfigState) {
	pv.stateImage.SetImage(iconForConfigState(state, 14))
	pv.stateText.SetText(configStateDescription[state])
	pv.toggleBtn.SetEnabled(state != consts.ConfigStateStarting && state != consts.ConfigStateStopping && state != consts.ConfigStateUnknown)
	if state == consts.ConfigStateStarted || state == consts.ConfigStateStopping {
		pv.toggleBtn.SetText(i18n.Sprintf("Stop"))
	} else {
		pv.toggleBtn.SetText(i18n.Sprintf("Start"))
	}
}

func (pv *PanelView) ToggleService() {
	conf := getCurrentConf()
	if conf == nil {
		return
	}
	var err error
	if conf.State == consts.ConfigStateStarted {
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
	oldState := conf.State
	setConfState(conf, consts.ConfigStateStarting)
	pv.setState(consts.ConfigStateStarting)
	go func() {
		if err := services.InstallService(conf.Name(), conf.Path, !conf.Data.AutoStart()); err != nil {
			pv.Synchronize(func() {
				showErrorMessage(pv.Form(), i18n.Sprintf("Start config \"%s\"", conf.Name()), err.Error())
				if conf.State == consts.ConfigStateStarting {
					setConfState(conf, oldState)
					if getCurrentConf() == conf {
						pv.setState(oldState)
					}
				}
			})
		}
	}()
	return nil
}

// StopService stops the service of the given config, then removes it
func (pv *PanelView) StopService(conf *Conf) (err error) {
	oldState := conf.State
	setConfState(conf, consts.ConfigStateStopping)
	pv.setState(consts.ConfigStateStopping)
	defer func() {
		if err != nil {
			setConfState(conf, oldState)
			pv.setState(oldState)
		}
	}()
	err = services.UninstallService(conf.Path, false)
	return
}

// Invalidate updates views using the current config
func (pv *PanelView) Invalidate(state bool) {
	conf := getCurrentConf()
	if conf == nil {
		pv.SetTitle("")
		pv.setState(consts.ConfigStateUnknown)
		pv.addressText.SetText("")
		pv.protoText.SetText("")
		pv.protoImage.SetVisible(false)
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
	pv.protoImage.SetVisible(data.TLSEnable || data.Protocol == consts.ProtoWSS || data.Protocol == consts.ProtoQUIC)
	proto := data.Protocol
	if proto == "" {
		proto = consts.ProtoTCP
	} else if proto == consts.ProtoWebsocket {
		proto = "ws"
	}
	proto = strings.ToUpper(proto)
	if data.HTTPProxy != "" {
		if u, err := url.Parse(data.HTTPProxy); err == nil {
			proto += " + " + strings.ToUpper(u.Scheme)
		}
	}
	if pv.protoText.Text() != proto {
		pv.protoText.SetText(proto)
	}
	if state {
		pv.setState(conf.State)
	}
}
