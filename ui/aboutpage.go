package ui

import (
	"encoding/json"
	"fmt"
	"frpmgr/config"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"
)

type AboutPage struct {
	view                  *walk.TabPage
	homeLink              *walk.PushButton
	frpLink               *walk.PushButton
	checkUpdateBtn        *walk.PushButton
	newVersionView        *walk.Composite
	newVersionDownloadBtn *walk.PushButton
	newVersionTag         *walk.Label
	newVersionDate        *walk.Label
}

func NewAboutPage() *AboutPage {
	v := new(AboutPage)
	return v
}

func (t *AboutPage) Initialize() {
	t.homeLink.SetCursor(walk.CursorHand())
	t.frpLink.SetCursor(walk.CursorHand())
	t.checkUpdate(false)
}

func (t *AboutPage) checkUpdate(showErr bool) {
	t.checkUpdateBtn.SetEnabled(false)
	t.checkUpdateBtn.SetText("正在检查更新")
	go func() {
		data := map[string]interface{}{}
		var body []byte
		r, err := http.Get("https://api.github.com/repos/koho/frpmgr/releases/latest")
		if err != nil {
			goto Fin
		}
		defer r.Body.Close()
		body, err = ioutil.ReadAll(r.Body)
		json.Unmarshal(body, &data)
	Fin:
		t.checkUpdateBtn.Synchronize(func() {
			t.checkUpdateBtn.SetEnabled(true)
			t.checkUpdateBtn.SetText("检查更新")
			if err != nil {
				if showErr {
					walk.MsgBox(t.checkUpdateBtn.Form(), "错误", "检查更新时出现错误。", walk.MsgBoxOK|walk.MsgBoxIconError)
				}
				return
			}
			if tagName, ok := data["tag_name"]; ok {
				if tagName.(string) != config.Version {
					t.newVersionView.SetVisible(true)
					t.newVersionTag.SetText(data["tag_name"].(string))
					pubDate, err := time.Parse("2006-01-02T15:04:05Z", data["published_at"].(string))
					if err == nil {
						t.newVersionDate.SetText(pubDate.Format("2006-01-02"))
					}
					t.newVersionDownloadBtn.SetName(data["html_url"].(string))
					t.view.SetTitle("新版本可用")
					t.view.SetImage(loadSysIcon("imageres", 1, 16))
				} else if showErr {
					walk.MsgBox(t.checkUpdateBtn.Form(), "提示", "已是最新版本。", walk.MsgBoxOK|walk.MsgBoxIconInformation)
				}
			} else {
				t.newVersionView.SetVisible(false)
				t.view.SetTitle("关于")
				t.view.SetImage(nil)
			}
		})
	}()
}

func (t *AboutPage) View() TabPage {
	logo, _ := loadLogoIcon(96)
	return TabPage{
		AssignTo: &t.view,
		Title:    "关于",
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					ImageView{Image: logo, Alignment: AlignHCenterVNear},
					Composite{
						Layout: VBox{Margins: Margins{12, 0, 0, 0}},
						Children: []Widget{
							Label{Text: "FRP Manager", Font: Font{Family: "微软雅黑", PointSize: 16}, TextColor: walk.RGB(11, 53, 137)},
							Label{Text: fmt.Sprintf("版本：%s", config.Version)},
							Label{Text: fmt.Sprintf("FRP 版本：%s", config.FRPVersion)},
							PushButton{AssignTo: &t.checkUpdateBtn, Text: "检查更新", OnClicked: func() {
								t.checkUpdate(true)
							}},
							VSpacer{Size: 6},
							Label{Text: "如有任何意见或报告错误，请访问项目地址："},
							PushButton{AssignTo: &t.homeLink, Text: "https://github.com/koho/frpmgr", OnClicked: func() {
								t.openURL(t.homeLink.Text())
							}},
							VSpacer{Size: 6},
							Label{Text: "了解 FRP 软件配置文档，请访问 FRP 项目地址："},
							PushButton{AssignTo: &t.frpLink, Text: "https://github.com/fatedier/frp", OnClicked: func() {
								t.openURL(t.frpLink.Text())
							}},
						},
					},
					HSpacer{},
				},
			},
			Composite{
				Visible:  false,
				AssignTo: &t.newVersionView,
				Layout:   VBox{},
				Children: []Widget{
					Label{Text: "新版本可用", Font: Font{Family: "微软雅黑", PointSize: 12}, TextColor: walk.RGB(0, 100, 0)},
					Composite{
						Layout: HBox{MarginsZero: true, Spacing: 100},
						Children: []Widget{
							Label{AssignTo: &t.newVersionTag},
							Label{AssignTo: &t.newVersionDate},
							PushButton{AssignTo: &t.newVersionDownloadBtn, Text: "下载", OnClicked: func() {
								t.openURL(t.newVersionDownloadBtn.Name())
							}},
							HSpacer{},
						},
					},
				},
			},
			VSpacer{},
		},
	}
}

func (t *AboutPage) openURL(url string) {
	if url == "" {
		return
	}
	exec.Command("cmd", "/c", "start", url).Start()
}
