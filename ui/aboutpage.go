package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/version"
)

type AboutPage struct {
	*walk.TabPage

	db        *walk.DataBinder
	viewModel aboutViewModel
}

type GithubRelease struct {
	TagName string `json:"tag_name"`
	HtmlUrl string `json:"html_url"`
}

type aboutViewModel struct {
	GithubRelease
	Checking   bool
	NewVersion bool
}

func NewAboutPage() *AboutPage {
	return new(AboutPage)
}

func (ap *AboutPage) Page() TabPage {
	return TabPage{
		AssignTo:   &ap.TabPage,
		Title:      Bind(fmt.Sprintf("vm.NewVersion ? '%s' : '%s'", i18n.Sprintf("New Version!"), i18n.Sprintf("About"))),
		Image:      Bind(fmt.Sprintf("vm.NewVersion ? sysIcon('imageres', 16, %d, %d) : ''", consts.IconNewVersion1, consts.IconNewVersion2)),
		DataBinder: DataBinder{AssignTo: &ap.db, Name: "vm", DataSource: &ap.viewModel},
		Layout:     VBox{},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					ImageView{Image: loadLogoIcon(96), Alignment: AlignHCenterVNear},
					Composite{
						Layout: VBox{Margins: Margins{Left: 12}},
						Children: []Widget{
							Label{Text: AppName, Font: consts.TextLarge, TextColor: consts.ColorDarkBlue},
							Label{Text: i18n.Sprintf("Version: %s", version.Number)},
							Label{Text: i18n.Sprintf("FRP version: %s", version.FRPVersion)},
							Label{Text: i18n.Sprintf("Built on: %s", version.BuildDate)},
							VSpacer{Size: 3},
							PushButton{
								Enabled: Bind("!vm.Checking"),
								Text: Bind(fmt.Sprintf("vm.NewVersion ? ' %s' : (vm.Checking ? '%s' : '%s')",
									i18n.Sprintf("Download updates"), i18n.Sprintf("Checking for updates"),
									i18n.Sprintf("Check for updates"),
								)),
								Font: consts.TextMedium,
								OnClicked: func() {
									if ap.viewModel.NewVersion {
										openPath(ap.viewModel.HtmlUrl)
									} else {
										ap.checkUpdate(true)
									}
								},
								Image:   Bind(fmt.Sprintf("vm.NewVersion ? sysIcon('shell32', 32, %d) : ''", consts.IconUpdate)),
								MinSize: Size{Width: 250, Height: 38},
							},
						},
					},
					HSpacer{},
				},
			},
			Composite{
				Layout:    VBox{Margins: Margins{Left: 123}},
				Alignment: AlignHNearVNear,
				Children: []Widget{
					Label{Text: i18n.Sprintf("For comments or to report bugs, please visit the project page:")},
					LinkLabel{
						Alignment: AlignHNearVCenter,
						Text:      fmt.Sprintf(`<a id="home" href="%s">%s</a>`, consts.ProjectURL, consts.ProjectURL),
						OnLinkActivated: func(link *walk.LinkLabelLink) {
							openPath(link.URL())
						},
					},
					VSpacer{Size: 6},
					Label{Text: i18n.Sprintf("For FRP configuration documentation, please visit the FRP project page:")},
					LinkLabel{
						Alignment: AlignHNearVCenter,
						Text:      fmt.Sprintf(`<a id="frp" href="%s">%s</a>`, consts.FRPProjectURL, consts.FRPProjectURL),
						OnLinkActivated: func(link *walk.LinkLabelLink) {
							openPath(link.URL())
						},
					},
				},
			},
			VSpacer{},
		},
	}
}

func (ap *AboutPage) OnCreate() {
	// Check update at launch
	ap.checkUpdate(false)
}

func (ap *AboutPage) checkUpdate(showErr bool) {
	ap.viewModel.Checking = true
	ap.db.Reset()
	go func() {
		var body []byte
		resp, err := http.Get(consts.UpdateURL)
		if err != nil {
			goto Fin
		}
		defer resp.Body.Close()
		if body, err = io.ReadAll(resp.Body); err != nil {
			goto Fin
		}
		ap.viewModel.GithubRelease = GithubRelease{}
		err = json.Unmarshal(body, &ap.viewModel.GithubRelease)
	Fin:
		ap.Synchronize(func() {
			ap.viewModel.Checking = false
			defer ap.db.Reset()
			if err != nil || resp.StatusCode != http.StatusOK {
				if showErr {
					showErrorMessage(ap.Form(), "", i18n.Sprintf("An error occurred while checking for a software update."))
				}
				return
			}
			if ap.viewModel.TagName != "" && ap.viewModel.TagName[1:] != version.Number {
				ap.viewModel.NewVersion = true
			} else {
				ap.viewModel.NewVersion = false
				if showErr {
					showInfoMessage(ap.Form(), "", i18n.Sprintf("There are currently no updates available."))
				}
			}
		})
	}()
}
