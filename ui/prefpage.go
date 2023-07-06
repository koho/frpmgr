package ui

import (
	"os"
	"sort"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/consts"
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
			pp.defaultSection(),
			VSpacer{},
		},
	}
}

func (pp *PrefPage) passwordSection() GroupBox {
	return GroupBox{
		Title:  i18n.Sprintf("Master password"),
		Layout: Grid{Alignment: AlignHNearVCenter, Columns: 2},
		Children: []Widget{
			ImageView{Image: loadResourceIcon(consts.IconKey, 32)},
			Label{Text: i18n.Sprintf("You can set a password to restrict access to this program.\nYou will be asked to enter it the next time you use this program.")},
			CheckBox{
				AssignTo: &pp.usePassword,
				Name:     "usePwd",
				Text:     i18n.Sprintf("Use master password"),
				Checked:  appConf.Password != "",
				Row:      1,
				Column:   1,
			},
			Composite{
				Row: 2, Column: 1,
				Layout: HBox{MarginsZero: true, Margins: Margins{Top: 5, Bottom: 5}, Spacing: 10},
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
			ImageView{Image: loadSysIcon("imageres", consts.IconLanguage, 32)},
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
			Composite{
				Row: 1, Column: 1,
				Layout: HBox{Margins: Margins{Top: 5, Bottom: 5}, Spacing: 10},
				Children: []Widget{
					Label{Text: i18n.SprintfColon("Select language")},
					ComboBox{
						AssignTo:      &langSelect,
						Model:         NewStringPairModel(keys, names, ""),
						MinSize:       Size{Width: 200},
						DisplayMember: "DisplayName",
						BindingMember: "Name",
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

func (pp *PrefPage) defaultSection() GroupBox {
	return GroupBox{
		Title:  i18n.Sprintf("Defaults"),
		Layout: Grid{Alignment: AlignHNearVCenter, Columns: 2, Spacing: 10, Margins: Margins{Left: 9, Top: 9, Right: 9, Bottom: 16}},
		Children: []Widget{
			ImageView{Image: loadSysIcon("imageres", consts.IconDefaults, 32)},
			Label{Text: i18n.Sprintf("Define the default value when creating a new configuration.\nThe value here will not affect the existing configuration.")},
			Composite{
				Row: 1, Column: 1,
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					PushButton{Text: i18n.Sprintf("Set Defaults"), MinSize: Size{Width: 150}, OnClicked: func() {
						if r, err := pp.setDefaultValue(); err == nil && r == win.IDOK {
							if err = saveAppConfig(); err != nil {
								showError(err, pp.Form())
							}
						}
					}},
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
	NewBasicDialog(nil, i18n.Sprintf("Master password"), loadResourceIcon(consts.IconKey, 32),
		DataBinder{
			AssignTo:       &db,
			DataSource:     &vm,
			ErrorPresenter: validators.SilentToolTipErrorPresenter{},
		}, nil, Composite{
			Layout:  VBox{MarginsZero: true},
			MinSize: Size{Width: 280},
			Children: []Widget{
				Label{Text: i18n.SprintfColon("New master password")},
				LineEdit{AssignTo: &pwdEdit, Text: Bind("Password", consts.ValidateNonEmpty), PasswordMode: true},
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
	if err := os.WriteFile(i18n.LangFile, []byte(lc), 0660); err != nil {
		showError(err, pp.Form())
	}
}

func (pp *PrefPage) setDefaultValue() (int, error) {
	dlg := NewBasicDialog(nil, i18n.Sprintf("Defaults"),
		loadSysIcon("imageres", consts.IconDefaults, 32),
		DataBinder{DataSource: &appConf.Defaults}, nil, Composite{
			Layout: Grid{Columns: 2, MarginsZero: true},
			Children: []Widget{
				Label{Text: i18n.SprintfColon("Protocol")},
				ComboBox{
					Name:  "proto",
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
				NumberEdit{Value: Bind("LogMaxDays"), Suffix: i18n.SprintfLSpace("Days")},
				Label{Text: i18n.SprintfColon("Auto Delete")},
				NumberEdit{Value: Bind("DeleteAfterDays"), Suffix: i18n.SprintfLSpace("Days")},
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
					Layout: Grid{MarginsZero: true, SpacingZero: true, Columns: 2},
					Children: []Widget{
						CheckBox{Text: i18n.Sprintf("TCP Mux"), Checked: Bind("TCPMux")},
						CheckBox{Text: "TLS", Checked: Bind("TLSEnable")},
						CheckBox{Text: i18n.Sprintf("Disable auto-start at boot"), Checked: Bind("ManualStart")},
					},
				},
			},
		}, VSpacer{})
	dlg.MinSize = Size{Width: 300}
	return dlg.Run(pp.Form())
}
