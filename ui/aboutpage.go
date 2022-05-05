package ui

import (
	"encoding/json"
	"fmt"
	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/version"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"io/ioutil"
	"net/http"
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
		Title:      Bind("vm.NewVersion ? '发现更新！' : '关于'"),
		Image:      Bind(fmt.Sprintf("vm.NewVersion ? sysIcon('imageres', 16, %d, %d) : ''", consts.IconNewVersion1, consts.IconNewVersion2)),
		DataBinder: DataBinder{AssignTo: &ap.db, Name: "vm", DataSource: &ap.viewModel},
		Layout:     VBox{},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					ImageView{Image: loadLogoIcon(96), Alignment: AlignHCenterVNear},
					Composite{
						Layout: VBox{Margins: Margins{12, 0, 0, 0}},
						Children: []Widget{
							Label{Text: "FRP Manager", Font: consts.TextLarge, TextColor: consts.ColorBlue},
							Label{Text: fmt.Sprintf("版本：%s", version.Number)},
							Label{Text: fmt.Sprintf("FRP 版本：%s", version.FRPVersion)},
							Label{Text: fmt.Sprintf("构建日期：%s", version.BuildDate)},
							VSpacer{Size: 3},
							PushButton{
								Enabled: Bind("!vm.Checking"),
								Text:    Bind("vm.NewVersion ? ' 下载更新' : (vm.Checking ? '正在' : '') + '检查更新'"),
								Font:    consts.TextMedium,
								OnClicked: func() {
									if ap.viewModel.NewVersion {
										openPath(ap.viewModel.HtmlUrl)
									} else {
										ap.checkUpdate(true)
									}
								},
								Image:   Bind(fmt.Sprintf("vm.NewVersion ? sysIcon('shell32', 32, %d) : ''", consts.IconUpdate)),
								MinSize: Size{Height: 38},
							},
							VSpacer{Size: 6},
							Label{Text: "如有任何意见或报告错误，请访问项目地址："},
							LinkLabel{
								Alignment: AlignHNearVCenter,
								Text:      `<a id="home" href="https://github.com/koho/frpmgr">https://github.com/koho/frpmgr</a>`,
								OnLinkActivated: func(link *walk.LinkLabelLink) {
									openPath(link.URL())
								},
							},
							VSpacer{Size: 6},
							Label{Text: "了解 FRP 软件配置文档，请访问 FRP 项目地址："},
							LinkLabel{
								Alignment: AlignHNearVCenter,
								Text:      `<a id="frp" href="https://github.com/fatedier/frp">https://github.com/fatedier/frp</a>`,
								OnLinkActivated: func(link *walk.LinkLabelLink) {
									openPath(link.URL())
								},
							},
						},
					},
					HSpacer{},
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
		resp, err := http.Get("https://api.github.com/repos/koho/frpmgr/releases/latest")
		if err != nil {
			goto Fin
		}
		defer resp.Body.Close()
		if body, err = ioutil.ReadAll(resp.Body); err != nil {
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
					showErrorMessage(ap.Form(), "错误", "检查更新时出现错误。")
				}
				return
			}
			if ap.viewModel.TagName != "" && ap.viewModel.TagName[1:] != version.Number {
				ap.viewModel.NewVersion = true
			} else {
				ap.viewModel.NewVersion = false
				if showErr {
					showInfoMessage(ap.Form(), "提示", "已是最新版本。")
				}
			}
		})
	}()
}
