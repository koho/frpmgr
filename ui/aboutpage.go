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
	"time"
)

type AboutPage struct {
	*walk.TabPage

	db             *walk.DataBinder
	viewModel      aboutViewModel
	checkUpdateBtn *walk.PushButton
}

type GithubRelease struct {
	TagName     string `json:"tag_name"`
	PublishedAt string `json:"published_at"`
	HtmlUrl     string `json:"html_url"`
}

type aboutViewModel struct {
	GithubRelease
	NewVersion bool
}

func NewAboutPage() *AboutPage {
	return new(AboutPage)
}

func (ap *AboutPage) Page() TabPage {
	return TabPage{
		AssignTo:   &ap.TabPage,
		Title:      "关于",
		DataBinder: DataBinder{AssignTo: &ap.db, DataSource: &ap.viewModel},
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
							Label{Text: fmt.Sprintf("版本：%s", version.Version)},
							Label{Text: fmt.Sprintf("FRP 版本：%s", version.FRPVersion)},
							Label{Text: fmt.Sprintf("构建日期：%s", version.BuildDate)},
							PushButton{AssignTo: &ap.checkUpdateBtn, Text: "检查更新", OnClicked: func() { ap.checkUpdate(true) }},
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
			Composite{
				Visible: Bind("NewVersion"),
				Layout:  VBox{},
				Children: []Widget{
					Label{Text: "新版本可用", Font: consts.TextMiddle, TextColor: consts.ColorGreen},
					Composite{
						Layout: HBox{MarginsZero: true, Spacing: 100},
						Children: []Widget{
							Label{Text: Bind("TagName")},
							Label{Text: Bind("PublishedAt")},
							PushButton{Text: "下载", OnClicked: func() { openPath(ap.viewModel.HtmlUrl) }},
							HSpacer{},
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
	ap.checkUpdateBtn.SetEnabled(false)
	ap.checkUpdateBtn.SetText("正在检查更新")
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
		ap.viewModel = aboutViewModel{}
		err = json.Unmarshal(body, &ap.viewModel)
	Fin:
		ap.Synchronize(func() {
			ap.checkUpdateBtn.SetEnabled(true)
			ap.checkUpdateBtn.SetText("检查更新")
			if err != nil || resp.StatusCode != http.StatusOK {
				if showErr {
					showErrorMessage(ap.Form(), "错误", "检查更新时出现错误。")
				}
				return
			}
			if ap.viewModel.TagName != "" && ap.viewModel.TagName != version.Version {
				ap.viewModel.NewVersion = true
				if pubDate, err := time.Parse("2006-01-02T15:04:05Z", ap.viewModel.PublishedAt); err == nil {
					ap.viewModel.PublishedAt = pubDate.Format("2006-01-02")
				}
				ap.SetTitle("新版本可用")
				ap.SetImage(loadNewVersionIcon(16))
			} else {
				ap.viewModel.NewVersion = false
				ap.SetTitle("关于")
				ap.SetImage(nil)
				if ap.viewModel.TagName == version.Version && showErr {
					showInfoMessage(ap.Form(), "提示", "已是最新版本。")
				}
			}
			ap.db.Reset()
		})
	}()
}
