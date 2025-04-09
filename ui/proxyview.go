package ui

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	frpconfig "github.com/fatedier/frp/pkg/config"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/config"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/res"
	"github.com/koho/frpmgr/pkg/util"
)

var proxyStateDescription = map[consts.ProxyState]string{
	consts.ProxyStateUnknown: i18n.Sprintf("Unknown"),
	consts.ProxyStateRunning: i18n.Sprintf("Running"),
	consts.ProxyStateError:   i18n.Sprintf("Error"),
}

var cachedProxyViewIconsForWidthAndState = make(map[widthAndProxyState]*walk.Bitmap)

type ProxyView struct {
	*walk.Composite

	model   *ProxyModel
	toolbar *walk.ToolBar
	table   *walk.TableView
	tracker *ProxyTracker

	// Actions
	newAction       *walk.Action
	portAction      *walk.Action
	rdAction        *walk.Action
	sshAction       *walk.Action
	webAction       *walk.Action
	vncAction       *walk.Action
	dnsAction       *walk.Action
	ftpAction       *walk.Action
	httpFileAction  *walk.Action
	httpProxyAction *walk.Action
	vpnAction       *walk.Action
	editAction      *walk.Action
	deleteAction    *walk.Action
	toggleAction    *walk.Action
}

func NewProxyView() *ProxyView {
	return new(ProxyView)
}

func (pv *ProxyView) View() Widget {
	return Composite{
		AssignTo: &pv.Composite,
		Layout:   VBox{MarginsZero: true},
		Children: []Widget{
			Composite{
				Layout:          HBox{MarginsZero: true, SpacingZero: true},
				Alignment:       AlignHNearVNear,
				DoubleBuffering: true,
				Children: []Widget{
					pv.createToolbar(),
				},
			},
			pv.createProxyTable(),
		},
	}
}

func (pv *ProxyView) OnCreate() {
	pv.editAction.SetDefault(true)
	pv.table.SelectedIndexesChanged().Attach(func() {
		if indexes := pv.table.SelectedIndexes(); len(indexes) == 1 {
			pv.table.SetCurrentIndex(indexes[0])
		}
	})
	pv.table.SelectedIndexesChanged().Attach(pv.switchToggleAction)
}

func (pv *ProxyView) Invalidate() {
	pv.stopTracker()
	if conf := getCurrentConf(); conf != nil {
		if _, ok := conf.Data.(*config.ClientConfig); ok {
			pv.model = NewProxyModel(conf)
			pv.table.SetModel(pv.model)
			if conf.State == consts.ConfigStateStarted {
				pv.startTracker(false)
			}
			return
		}
	}
	pv.model = nil
	pv.table.SetModel(nil)
}

func (pv *ProxyView) startTracker(refresh bool) bool {
	if pv.tracker == nil && pv.model != nil {
		pv.tracker = NewProxyTracker(pv.Form(), pv.model, refresh)
		return true
	}
	return false
}

func (pv *ProxyView) stopTracker() bool {
	if pv.tracker != nil {
		pv.tracker.Close()
		pv.tracker = nil
		return true
	}
	return false
}

func (pv *ProxyView) resetProxyState(row int) {
	if pv.model == nil {
		return
	}
	items := pv.model.items
	if row >= 0 {
		items = pv.model.items[row : row+1]
	} else {
		defer pv.table.DisposeImageList()
	}
	for i, item := range items {
		item.State = consts.ProxyStateUnknown
		item.StateSource = ""
		item.Error = ""
		item.RemoteAddr = ""
		if item.RemotePort != item.DisplayRemotePort {
			item.DisplayRemotePort = item.RemotePort
			if row < 0 {
				pv.model.PublishRowChanged(i)
			}
		}
	}
}

func (pv *ProxyView) createToolbar() ToolBar {
	mc := movingConditions()
	return ToolBar{
		AssignTo:    &pv.toolbar,
		ButtonStyle: ToolBarButtonImageBeforeText,
		Orientation: Horizontal,
		Items: []MenuItem{
			Action{
				AssignTo: &pv.newAction,
				Text:     i18n.Sprintf("Add"),
				Image:    loadIcon(res.IconCreate, 16),
				OnTriggered: func() {
					pv.onEdit(nil, true)
				},
			},
			Menu{
				Text:  i18n.Sprintf("Quick Add"),
				Image: loadIcon(res.IconQuickAdd, 16),
				Items: []MenuItem{
					Action{
						AssignTo: &pv.portAction,
						Text:     i18n.Sprintf("Open Port"),
						Image:    loadIcon(res.IconOpenPort, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewPortProxyDialog())
						},
					},
					Action{
						AssignTo: &pv.rdAction,
						Text:     i18n.Sprintf("Remote Desktop"),
						Image:    loadIcon(res.IconRemote, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog(i18n.Sprintf("Remote Desktop"), loadIcon(res.IconRemote, 32),
								"rdp", []string{consts.ProxyTypeTCP, consts.ProxyTypeUDP}, ":3389"))
						},
					},
					Action{
						AssignTo: &pv.vncAction,
						Text:     "VNC",
						Image:    loadIcon(res.IconVNC, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog("VNC", loadIcon(res.IconVNC, 32),
								"vnc", []string{consts.ProxyTypeTCP, consts.ProxyTypeUDP}, ":5900"))
						},
					},
					Action{
						AssignTo: &pv.sshAction,
						Text:     "SSH",
						Image:    loadIcon(res.IconSSH, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog("SSH", loadIcon(res.IconSSH, 32),
								"ssh", []string{consts.ProxyTypeTCP}, ":22"))
						},
					},
					Action{
						AssignTo: &pv.webAction,
						Text:     "Web",
						Image:    loadIcon(res.IconWeb, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog("Web", loadIcon(res.IconWeb, 32),
								"web", []string{consts.ProxyTypeTCP}, ":80"))
						},
					},
					Action{
						AssignTo: &pv.dnsAction,
						Text:     "DNS",
						Image:    loadIcon(res.IconDns, 16),
						OnTriggered: func() {
							systemDns := util.GetSystemDnsServer()
							if systemDns == "" {
								systemDns = "114.114.114.114"
							}
							pv.onQuickAdd(NewSimpleProxyDialog("DNS", loadIcon(res.IconDns, 32),
								"dns", []string{consts.ProxyTypeUDP}, systemDns+":53"))
						},
					},
					Action{
						AssignTo: &pv.vpnAction,
						Text:     "OpenVPN",
						Image:    loadIcon(res.IconLock, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog("OpenVPN", loadIcon(res.IconLock, 32),
								"openvpn", []string{consts.ProxyTypeTCP, consts.ProxyTypeUDP}, ":1194"))
						},
					},
					Action{
						AssignTo: &pv.ftpAction,
						Text:     "FTP",
						Image:    loadIcon(res.IconFtp, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewSimpleProxyDialog("FTP", loadIcon(res.IconFtp, 32),
								"ftp", []string{consts.ProxyTypeTCP}, ":21"))
						},
					},
					Action{
						AssignTo: &pv.httpFileAction,
						Text:     i18n.Sprintf("HTTP File Server"),
						Image:    loadIcon(res.IconHttpFile, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewPluginProxyDialog(i18n.Sprintf("HTTP File Server"), loadIcon(res.IconHttpFile, 32),
								consts.PluginStaticFile))
						},
					},
					Action{
						AssignTo: &pv.httpProxyAction,
						Text:     i18n.Sprintf("HTTP Proxy"),
						Image:    loadIcon(res.IconHttpProxy, 16),
						OnTriggered: func() {
							pv.onQuickAdd(NewPluginProxyDialog(i18n.Sprintf("HTTP Proxy"), loadIcon(res.IconHttpProxy, 32),
								consts.PluginHttpProxy))
						},
					},
				},
			},
			Action{
				AssignTo:    &pv.editAction,
				Image:       loadIcon(res.IconEdit, 16),
				Text:        i18n.Sprintf("Edit"),
				Enabled:     Bind("proxy.SelectedCount == 1"),
				OnTriggered: pv.editCurrent,
			},
			Action{
				AssignTo:    &pv.toggleAction,
				Image:       loadIcon(res.IconDisable, 16),
				Text:        i18n.Sprintf("Disable"),
				Enabled:     false,
				OnTriggered: pv.onToggleProxy,
			},
			Action{
				Image:   loadIcon(res.IconArrowUp, 16),
				Text:    i18n.Sprintf("Move Up"),
				Enabled: mc[0],
				OnTriggered: func() {
					pv.onMove(-1)
				},
			},
			Action{
				Image:   flipIcon(res.IconArrowUp, 16),
				Text:    i18n.Sprintf("Move Down"),
				Enabled: mc[1],
				OnTriggered: func() {
					pv.onMove(1)
				},
			},
			Action{
				AssignTo:    &pv.deleteAction,
				Image:       loadIcon(res.IconDelete, 16),
				Text:        i18n.Sprintf("Delete"),
				Enabled:     Bind("proxy.SelectedCount > 0"),
				OnTriggered: pv.onDelete,
			},
		},
	}
}

func (pv *ProxyView) createProxyTable() TableView {
	mc := movingConditions()
	return TableView{
		Name:     "proxy",
		AssignTo: &pv.table,
		Columns: []TableViewColumn{
			{Title: i18n.Sprintf("Name"), DataMember: "Name", Width: 100},
			{Title: i18n.Sprintf("Type"), DataMember: "Type", Width: 55},
			{Title: i18n.Sprintf("Local Address"), DataMember: "DisplayLocalIP", Width: 110},
			{Title: i18n.Sprintf("Local Port"), DataMember: "DisplayLocalPort", Width: 90},
			{Title: i18n.Sprintf("Remote Port"), DataMember: "DisplayRemotePort", Width: 90},
			{Title: i18n.Sprintf("Domains"), DataMember: "Domains", Width: 80},
			{Title: i18n.Sprintf("Plugin"), DataMember: "Plugin", Width: 80},
			{Title: i18n.Sprintf("Remote Address"), DataMember: "RemoteAddr", Width: 110, Name: "remoteAddr", Hidden: true},
		},
		MultiSelection: true,
		ContextMenuItems: []MenuItem{
			ActionRef{Action: &pv.editAction},
			ActionRef{Action: &pv.toggleAction},
			Menu{
				Text:    i18n.Sprintf("Move"),
				Image:   loadIcon(res.IconMove, 16),
				Enabled: Bind("proxy.SelectedCount == 1 && proxy.ItemCount > 1"),
				Items: []MenuItem{
					Action{
						Text:    i18n.Sprintf("Up"),
						Enabled: mc[0],
						OnTriggered: func() {
							pv.onMove(-1)
						},
					},
					Action{
						Text:    i18n.Sprintf("Down"),
						Enabled: mc[1],
						OnTriggered: func() {
							pv.onMove(1)
						},
					},
					Action{
						Text:    i18n.Sprintf("To Top"),
						Enabled: mc[0],
						OnTriggered: func() {
							pv.onMove(-pv.table.CurrentIndex())
						},
					},
					Action{
						Text:    i18n.Sprintf("To Bottom"),
						Enabled: mc[1],
						OnTriggered: func() {
							if pv.model == nil {
								return
							}
							pv.onMove(len(pv.model.items) - pv.table.CurrentIndex() - 1)
						},
					},
				},
			},
			Separator{},
			ActionRef{Action: &pv.newAction},
			Menu{
				Text:  i18n.Sprintf("Quick Add"),
				Image: loadIcon(res.IconQuickAdd, 16),
				Items: []MenuItem{
					ActionRef{Action: &pv.portAction},
					ActionRef{Action: &pv.rdAction},
					ActionRef{Action: &pv.vncAction},
					ActionRef{Action: &pv.sshAction},
					ActionRef{Action: &pv.webAction},
					ActionRef{Action: &pv.dnsAction},
					ActionRef{Action: &pv.vpnAction},
					ActionRef{Action: &pv.ftpAction},
					ActionRef{Action: &pv.httpFileAction},
					ActionRef{Action: &pv.httpProxyAction},
				},
			},
			Action{
				Text:        i18n.Sprintf("Import from Clipboard"),
				Image:       loadIcon(res.IconClipboard, 16),
				OnTriggered: pv.onClipboardImport,
			},
			Separator{},
			Action{
				Text:      i18n.Sprintf("Show Remote Address"),
				Checkable: true,
				OnTriggered: func() {
					if col := pv.table.Columns().ByName("remoteAddr"); col != nil {
						col.SetVisible(!col.Visible())
						if col.Visible() {
							pv.table.ScrollToLastColumn()
						}
					}
				},
			},
			Action{
				Enabled:     Bind("proxy.SelectedCount == 1"),
				Text:        i18n.Sprintf("Copy Access Address"),
				Image:       loadIcon(res.IconSysCopy, 16),
				OnTriggered: pv.onCopyAccessAddr,
			},
			Action{
				Enabled: Bind("proxy.SelectedCount < proxy.ItemCount"),
				Text:    i18n.Sprintf("Select all"),
				Image:   loadIcon(res.IconSelectAll, 16),
				OnTriggered: func() {
					pv.table.SetSelectedIndexes([]int{-1})
				},
			},
			Separator{},
			ActionRef{Action: &pv.deleteAction},
		},
		OnItemActivated: pv.editCurrent,
		ImageAsState:    true,
		StyleCell: func(style *walk.CellStyle) {
			if pv.model == nil {
				return
			}
			proxy := pv.model.items[style.Row()]
			if proxy.Disabled {
				// Disabled proxy
				style.TextColor = res.ColorGray
				style.BackgroundColor = res.ColorGrayBG
			} else if proxy.IsVisitor() {
				// Visitor proxy
				style.TextColor = res.ColorBlue
			}
			// Normal proxy is default black text
			if style.Col() == 0 && pv.model.conf.State == consts.ConfigStateStarted {
				margin := pv.IntFrom96DPI(2)
				bitmapWidth := pv.IntFrom96DPI(16)
				cacheKey := widthAndProxyState{bitmapWidth, proxy.State}
				if cacheValue, ok := cachedProxyViewIconsForWidthAndState[cacheKey]; ok {
					style.Image = cacheValue
					return
				}
				bitmap, err := walk.NewBitmapWithTransparentPixelsForDPI(walk.Size{Width: bitmapWidth, Height: bitmapWidth}, pv.DPI())
				if err != nil {
					return
				}
				canvas, err := walk.NewCanvasFromImage(bitmap)
				if err != nil {
					return
				}
				bounds := walk.Rectangle{X: margin, Y: margin, Height: bitmapWidth - 2*margin, Width: bitmapWidth - 2*margin}
				err = canvas.DrawImageStretchedPixels(iconForProxyState(proxy.State, 12), bounds)
				canvas.Dispose()
				if err != nil {
					return
				}
				cachedProxyViewIconsForWidthAndState[cacheKey] = bitmap
				style.Image = bitmap
			} else {
				style.Image = nil
			}
		},
		ItemToolTip: func(i int) string {
			if pv.model == nil || pv.model.items[i].Disabled {
				return ""
			}
			if pv.model.conf.State == consts.ConfigStateStarted {
				proxy := pv.model.items[i]
				tooltip := i18n.SprintfColon("Status") + " " + proxyStateDescription[proxy.State]
				if proxy.Error != "" {
					tooltip += "\n" + i18n.SprintfColon("Error message") + " " + proxy.Error
					if proxy.StateSource != proxy.Name {
						tooltip += "\n" + i18n.SprintfColon("Source") + " " + proxy.StateSource
					}
				}
				return tooltip
			}
			return ""
		},
	}
}

func (pv *ProxyView) editCurrent() {
	idx := pv.table.CurrentIndex()
	if idx < 0 || pv.model == nil {
		return
	}
	pv.onEdit(pv.model.items[idx].Proxy, false)
}

func (pv *ProxyView) onCopyAccessAddr() {
	idx := pv.table.CurrentIndex()
	if idx < 0 || pv.model == nil {
		return
	}
	proxy := pv.model.items[idx]
	var access string
	switch proxy.Type {
	case consts.ProxyTypeTCP, consts.ProxyTypeUDP:
		if proxy.RemoteAddr != "" {
			access = proxy.RemoteAddr
		} else if proxy.RemotePort != "" {
			access = pv.model.data.ServerAddress + ":" + strings.Split(strings.Split(proxy.RemotePort, ",")[0], "-")[0]
		}
	case consts.ProxyTypeXTCP, consts.ProxyTypeSTCP, consts.ProxyTypeSUDP:
		if proxy.Role == "visitor" {
			if proxy.BindPort > 0 {
				access = util.GetOrElse(proxy.BindAddr, "127.0.0.1") + ":" + strconv.Itoa(proxy.BindPort)
			}
		} else {
			access = util.GetOrElse(proxy.LocalIP, "127.0.0.1") + ":" + proxy.LocalPort
		}
	case consts.ProxyTypeHTTP, consts.ProxyTypeHTTPS, consts.ProxyTypeTCPMUX:
		scheme := proxy.Type
		if proxy.Type == consts.ProxyTypeTCPMUX {
			scheme = "http"
		}
		if proxy.RemoteAddr != "" {
			access = fmt.Sprintf("%s://%s", scheme, strings.Split(proxy.RemoteAddr, ",")[0])
		} else if proxy.CustomDomains != "" {
			access = fmt.Sprintf("%s://%s", scheme, strings.Split(proxy.CustomDomains, ",")[0])
		} else if proxy.SubDomain != "" && net.ParseIP(pv.model.data.ServerAddress) == nil {
			// Assume subdomain_host is equal to server_address
			access = fmt.Sprintf("%s://%s.%s", scheme, proxy.SubDomain, pv.model.data.ServerAddress)
		}
	}
	walk.Clipboard().SetText(access)
}

func (pv *ProxyView) onClipboardImport() {
	text, err := walk.Clipboard().Text()
	if err != nil || strings.TrimSpace(text) == "" {
		return
	}
	var proxy *config.Proxy
	if strings.HasPrefix(text, "[[proxies]]") {
		var proxies struct {
			C []config.TypedProxyConfig `json:"proxies"`
		}
		if err = frpconfig.LoadConfigure([]byte(text), &proxies, false); err == nil && len(proxies.C) > 0 {
			proxy = config.ClientProxyFromV1(proxies.C[0])
		}
	} else if strings.HasPrefix(text, "[[visitors]]") {
		var visitors struct {
			C []config.TypedVisitorConfig `json:"visitors"`
		}
		if err = frpconfig.LoadConfigure([]byte(text), &visitors, false); err == nil && len(visitors.C) > 0 {
			proxy = config.ClientVisitorFromV1(visitors.C[0])
		}
	} else if strings.HasPrefix(text, "[") {
		proxy, err = config.UnmarshalProxyFromIni([]byte(text))
	} else {
		showErrorMessage(pv.Form(), "", i18n.Sprintf("This feature only supports text in INI or TOML format."))
		return
	}
	if err != nil {
		showError(err, pv.Form())
		return
	}
	if proxy == nil {
		return
	}
	pv.onEdit(proxy, true)
}

func (pv *ProxyView) onDelete() {
	if pv.model == nil {
		return
	}
	indexes := pv.table.SelectedIndexes()
	count := len(indexes)
	if count == 1 {
		name := pv.model.items[indexes[0]].Name
		if walk.MsgBox(pv.Form(), i18n.Sprintf("Delete proxy \"%s\"", name),
			i18n.Sprintf("Are you sure you would like to delete proxy \"%s\"?", name),
			walk.MsgBoxYesNo|walk.MsgBoxIconWarning) == walk.DlgCmdNo {
			return
		}
		pv.model.Remove(indexes[0])
	} else if count > 1 {
		if walk.MsgBox(pv.Form(), i18n.Sprintf("Delete %d proxies", count),
			i18n.Sprintf("Are you sure that you want to delete these %d proxies?", count),
			walk.MsgBoxYesNo|walk.MsgBoxIconWarning) == walk.DlgCmdNo {
			return
		}
		pv.model.Remove(indexes...)
	}
	pv.commit()
}

func (pv *ProxyView) onEdit(proxy *config.Proxy, create bool) {
	if pv.model == nil {
		return
	}
	except := proxy
	if create {
		except = nil
	}
	var oldName string
	if proxy != nil {
		oldName = proxy.Name
	}
	dlg := NewEditProxyDialog(proxy, pv.visitors(except), create, pv.model.data.LegacyFormat, pv.model.HasName)
	if result, _ := dlg.Run(pv.Form()); result == walk.DlgCmdOK {
		if create {
			pv.model.Add(dlg.Proxy)
			pv.table.SetCurrentIndex(len(pv.model.items) - 1)
			pv.table.SetFocus()
		} else {
			if dlg.Proxy.Name != oldName {
				pv.model.PublishRowRenamed(pv.table.CurrentIndex())
			}
			if i := pv.table.CurrentIndex(); i >= 0 {
				pv.model.Reset(i)
			}
		}
		pv.commit()
	}
}

func (pv *ProxyView) onToggleProxy() {
	indexes := pv.table.SelectedIndexes()
	count := len(indexes)
	if count == 0 || pv.model == nil {
		return
	}
	proxy := pv.model.items[indexes[0]]
	if !proxy.Disabled {
		// We can't disable all proxies
		if pv.model.data.CountStart()-count <= 0 {
			return
		}
		if cc := getCurrentConf(); cc != nil && cc.State == consts.ConfigStateStarted {
			if count == 1 {
				if walk.MsgBox(pv.Form(), i18n.Sprintf("Disable proxy \"%s\"", proxy.Name),
					i18n.Sprintf("Are you sure you would like to disable proxy \"%s\"?", proxy.Name),
					walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) == walk.DlgCmdNo {
					return
				}
			} else {
				if walk.MsgBox(pv.Form(), i18n.Sprintf("Disable %d proxies", count),
					i18n.Sprintf("Are you sure that you want to disable these %d proxies?", count),
					walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) == walk.DlgCmdNo {
					return
				}
			}
		}
	} else {
		defer pv.model.PublishRowEdited(indexes[0])
	}
	if pv.tracker != nil {
		pv.tracker.Lock()
	}
	for _, idx := range indexes {
		proxy = pv.model.items[idx]
		proxy.Disabled = !proxy.Disabled
		if proxy.Disabled {
			pv.resetProxyState(idx)
		}
		pv.model.PublishRowChanged(idx)
	}
	if pv.tracker != nil {
		pv.tracker.Unlock()
	}
	pv.switchToggleAction()
	pv.commit()
}

func (pv *ProxyView) onQuickAdd(qa QuickAdd) {
	if pv.model == nil {
		return
	}
	var proxies []*config.Proxy
	if r, _ := qa.Run(pv.Form()); r == walk.DlgCmdOK {
		for _, proxy := range qa.GetProxies() {
			if pv.model.HasName(proxy.Name) {
				showWarningMessage(pv.Form(), i18n.Sprintf("Proxy already exists"), i18n.Sprintf("The proxy name \"%s\" already exists.", proxy.Name))
			} else {
				proxies = append(proxies, proxy)
			}
		}
		if len(proxies) > 0 {
			pv.model.Add(proxies...)
			pv.table.SetCurrentIndex(len(pv.model.items) - 1)
			pv.table.SetFocus()
			pv.commit()
		}
	}
}

func (pv *ProxyView) onMove(delta int) {
	if pv.model == nil {
		return
	}
	curIdx := pv.table.CurrentIndex()
	if curIdx < 0 || curIdx >= len(pv.model.items) {
		return
	}
	targetIdx := curIdx + delta
	if targetIdx < 0 || targetIdx >= len(pv.model.items) {
		return
	}
	pv.model.Move(curIdx, targetIdx)
	commitConf(pv.model.conf, runFlagReload)
	pv.table.SetCurrentIndex(targetIdx)
}

// switchToggleAction updates the toggle action based on the current selected proxies
func (pv *ProxyView) switchToggleAction() {
	indexes := pv.table.SelectedIndexes()
	if len(indexes) == 0 || pv.model == nil {
		pv.toggleAction.SetEnabled(false)
		return
	}
	count := 0
	for _, idx := range indexes {
		if !pv.model.items[idx].Disabled {
			count++
		}
	}
	if count == 0 {
		pv.toggleAction.SetText(i18n.Sprintf("Enable"))
		pv.toggleAction.SetImage(loadIcon(res.IconEnable, 16))
		pv.toggleAction.SetEnabled(true)
	} else if count == len(indexes) {
		pv.toggleAction.SetText(i18n.Sprintf("Disable"))
		pv.toggleAction.SetImage(loadIcon(res.IconDisable, 16))
		pv.toggleAction.SetEnabled(pv.model.data.CountStart()-len(indexes) >= 1)
	} else {
		pv.toggleAction.SetEnabled(false)
	}
}

// commit will update the views and save the config to disk, then reload service
func (pv *ProxyView) commit() {
	if pv.model != nil {
		commitConf(pv.model.conf, runFlagReload)
	}
}

// visitors returns a list of visitor names except the given proxy.
func (pv *ProxyView) visitors(except *config.Proxy) (visitors []string) {
	if pv.model == nil {
		return
	}
	for _, item := range pv.model.items {
		if item.Proxy != except && item.IsVisitor() {
			visitors = append(visitors, item.Name)
		}
	}
	return
}

// Conditions for moving up/down proxy.
func movingConditions() [2]Property {
	return [2]Property{
		Bind("proxy.SelectedCount == 1 && proxy.CurrentIndex > 0"),
		Bind("proxy.SelectedCount == 1 && proxy.CurrentIndex < proxy.ItemCount - 1"),
	}
}
