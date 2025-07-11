package ui

import (
	"math"
	"sort"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"github.com/samber/lo"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/res"
	"github.com/koho/frpmgr/pkg/sec"
	"github.com/koho/frpmgr/pkg/validators"
)

type PrefPage struct {
	*walk.TabPage

	usePassword *walk.CheckBox
}

func NewPrefPage() *PrefPage {
	return new(PrefPage)
}

func (pp *PrefPage) OnCreate() {
	pp.usePassword.CheckedChanged().Attach(pp.switchPassword)
}

func (pp *PrefPage) Page() TabPage {
	return TabPage{
		AssignTo: &pp.TabPage,
		Title:    i18n.Sprintf("Preferences"),
		Layout:   VBox{},
		Children: []Widget{
			pp.passwordSection(),
			pp.languageSection(),
			pp.advancedSection(),
			VSpacer{},
		},
	}
}

func (pp *PrefPage) passwordSection() GroupBox {
	return GroupBox{
		Title:  i18n.Sprintf("Master password"),
		Layout: Grid{Alignment: AlignHNearVCenter, Columns: 2},
		Children: []Widget{
			ImageView{Image: loadIcon(res.IconKey, 32)},
			Label{Text: i18n.Sprintf("You can set a password to restrict access to this program.\nYou will be asked to enter it the next time you use this program.")},
			HSpacer{Size: 42},
			CheckBox{
				AssignTo: &pp.usePassword,
				Name:     "usePwd",
				Text:     i18n.Sprintf("Use master password"),
				Checked:  appConf.Password != "",
			},
			Composite{
				Row: 2, Column: 1,
				Layout: HBox{Margins: Margins{Top: 5, Bottom: 5}},
				Children: []Widget{
					PushButton{
						MinSize: Size{Width: 150},
						Text:    i18n.Sprintf("Change Password"),
						Enabled: Bind("usePwd.Checked"),
						OnClicked: func() {
							pp.changePassword()
						},
					},
					HSpacer{},
				},
			},
		},
	}
}

func (pp *PrefPage) languageSection() GroupBox {
	keys := make([]string, 0, len(i18n.IDToName))
	for k := range i18n.IDToName {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	names := make([]string, len(keys))
	for i := range keys {
		names[i] = i18n.IDToName[keys[i]]
	}
	var langSelect *walk.ComboBox
	return GroupBox{
		Title:  i18n.Sprintf("Languages"),
		Layout: Grid{Alignment: AlignHNearVCenter, Columns: 2},
		Children: []Widget{
			ImageView{Image: loadIcon(res.IconLanguage, 32)},
			Composite{
				Layout: VBox{MarginsZero: true},
				Children: []Widget{
					Composite{
						Layout: HBox{MarginsZero: true},
						Children: []Widget{
							Label{Text: i18n.SprintfColon("The current display language is")},
							LineEdit{Text: i18n.IDToName[i18n.GetLanguage()], ReadOnly: true, MaxSize: Size{Width: 200}},
							HSpacer{},
						},
					},
					Label{Text: i18n.Sprintf("You must restart program to apply the modification.")},
				},
			},
			HSpacer{Size: 42},
			Composite{
				Layout: HBox{Margins: Margins{Top: 5, Bottom: 5}, Spacing: 10},
				Children: []Widget{
					Label{Text: i18n.SprintfColon("Select language")},
					ComboBox{
						AssignTo:      &langSelect,
						Model:         NewListModel(keys, lo.ToAnySlice(names)...),
						MinSize:       Size{Width: 200},
						DisplayMember: "Title",
						BindingMember: "Value",
						Value:         i18n.GetLanguage(),
						OnCurrentIndexChanged: func() {
							pp.switchLanguage(keys[langSelect.CurrentIndex()])
						},
					},
					HSpacer{},
				},
			},
		},
	}
}

func (pp *PrefPage) advancedSection() GroupBox {
	return GroupBox{
		Title:  i18n.Sprintf("Advanced"),
		Layout: Grid{Alignment: AlignHNearVCenter, Columns: 2},
		Children: []Widget{
			ImageView{Image: loadIcon(res.IconSettings, 32)},
			Label{Text: i18n.Sprintf("You can find more settings here.\nIncludes application updates, initial default values, etc.")},
			HSpacer{Size: 42},
			Composite{
				Layout: HBox{Margins: Margins{Top: 5, Bottom: 5}},
				Children: []Widget{
					PushButton{
						Text:    i18n.Sprintf("Settings"),
						MinSize: Size{Width: 100},
						OnClicked: func() {
							if r, err := pp.setAdvancedSettings(); err == nil && r == win.IDOK {
								if err = saveAppConfig(); err != nil {
									showError(err, pp.Form())
								}
							}
						},
					},
					HSpacer{},
				},
			},
		},
	}
}

func (pp *PrefPage) switchPassword() {
	if pp.usePassword.Checked() {
		if newPassword := pp.changePassword(); newPassword == "" && appConf.Password == "" {
			pp.usePassword.SetChecked(false)
		}
	} else {
		if appConf.Password != "" {
			appConf.Password = ""
			if err := saveAppConfig(); err != nil {
				showError(err, pp.Form())
				return
			}
			showInfoMessage(pp.Form(), "", i18n.Sprintf("Password removed."))
		}
	}
}

func (pp *PrefPage) changePassword() string {
	var db *walk.DataBinder
	var pwdEdit *walk.LineEdit
	var vm struct {
		Password string
	}
	NewBasicDialog(nil, i18n.Sprintf("Master password"), loadIcon(res.IconKey, 32),
		DataBinder{
			AssignTo:       &db,
			DataSource:     &vm,
			ErrorPresenter: validators.SilentToolTipErrorPresenter{},
		}, nil, Composite{
			Layout:  VBox{MarginsZero: true},
			MinSize: Size{Width: 280},
			Children: []Widget{
				Label{Text: i18n.SprintfColon("New master password")},
				LineEdit{AssignTo: &pwdEdit, Text: Bind("Password", res.ValidateNonEmpty), PasswordMode: true},
				Label{Text: i18n.SprintfColon("Re-enter password")},
				LineEdit{Text: Bind("", validators.ConfirmPassword{Password: &pwdEdit}), PasswordMode: true},
			},
		}, VSpacer{}).Run(pp.Form())
	if vm.Password != "" {
		oldPassword := appConf.Password
		appConf.Password = sec.EncryptPassword(vm.Password)
		if err := saveAppConfig(); err != nil {
			appConf.Password = oldPassword
			showError(err, pp.Form())
		} else {
			showInfoMessage(pp.Form(), "", i18n.Sprintf("Password is set."))
		}
	}
	return vm.Password
}

func (pp *PrefPage) switchLanguage(lc string) {
	appConf.Lang = lc
	if err := saveAppConfig(); err != nil {
		showError(err, pp.Form())
	}
}

func (pp *PrefPage) setAdvancedSettings() (int, error) {
	var w *walk.Dialog
	var dbs [2]*walk.DataBinder
	dlg := NewBasicDialog(&w, i18n.Sprintf("Advanced"),
		loadIcon(res.IconSettings, 32),
		DataBinder{}, func() {
			for _, db := range dbs {
				if err := db.Submit(); err != nil {
					return
				}
			}
			w.Accept()
		}, Composite{
			Layout: VBox{Margins: Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}},
			Children: []Widget{
				GroupBox{
					Title:      i18n.Sprintf("General"),
					Layout:     HBox{},
					DataBinder: DataBinder{AssignTo: &dbs[0], DataSource: &appConf},
					Children: []Widget{
						CheckBox{
							Text:    i18n.Sprintf("Automatically check for updates"),
							Checked: Bind("CheckUpdate"),
						},
						HSpacer{},
					},
				},
				GroupBox{
					Title:      i18n.Sprintf("Defaults"),
					Layout:     Grid{Columns: 2},
					DataBinder: DataBinder{AssignTo: &dbs[1], DataSource: &appConf.Defaults},
					Children: []Widget{
						Label{Text: i18n.SprintfColon("Protocol")},
						ComboBox{
							Value: Bind("Protocol"),
							Model: consts.Protocols,
						},
						Label{Text: i18n.SprintfColon("User")},
						LineEdit{Text: Bind("User")},
						Label{Text: i18n.SprintfColon("Log Level")},
						ComboBox{
							Value: Bind("LogLevel"),
							Model: consts.LogLevels,
						},
						Label{Text: i18n.SprintfColon("Log retention")},
						NewNumberInput(NIOption{Value: Bind("LogMaxDays"), Suffix: i18n.Sprintf("Days"), Max: math.MaxFloat64}),
						Label{Text: "DNS:"},
						LineEdit{Text: Bind("DNSServer")},
						Label{Text: i18n.SprintfColon("STUN Server")},
						LineEdit{Text: Bind("NatHoleSTUNServer")},
						Label{Text: i18n.SprintfColon("Source Address")},
						LineEdit{Text: Bind("ConnectServerLocalIP")},
						Composite{
							Layout: VBox{MarginsZero: true, SpacingZero: true},
							Children: []Widget{
								VSpacer{Size: 6},
								Label{Text: i18n.SprintfColon("Other Options"), Alignment: AlignHNearVNear},
							},
						},
						Composite{
							Layout: VBox{MarginsZero: true, SpacingZero: true, Alignment: AlignHNearVNear},
							Children: []Widget{
								Composite{
									Layout: HBox{MarginsZero: true},
									Children: []Widget{
										CheckBox{Text: i18n.Sprintf("TCP Mux"), Checked: Bind("TCPMux")},
										CheckBox{Text: "TLS", Checked: Bind("TLSEnable")},
									},
								},
								CheckBox{Text: i18n.Sprintf("Disable auto-start at boot"), Checked: Bind("ManualStart")},
								CheckBox{Text: i18n.Sprintf("Use legacy file format"), Checked: Bind("LegacyFormat")},
							},
						},
					},
				},
			},
		})
	dlg.MinSize = Size{Width: 350}
	dlg.FixedSize = true
	return dlg.Run(pp.Form())
}
