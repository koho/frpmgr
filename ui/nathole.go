package ui

import (
	"fmt"

	"github.com/fatedier/frp/pkg/nathole"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/res"
)

type NATDiscoveryDialog struct {
	*walk.Dialog

	table   *walk.TableView
	barView *walk.ProgressBar

	// STUN server address
	serverAddr string
	closed     bool
}

func NewNATDiscoveryDialog(serverAddr string) *NATDiscoveryDialog {
	return &NATDiscoveryDialog{serverAddr: serverAddr}
}

func (nd *NATDiscoveryDialog) Run(owner walk.Form) (int, error) {
	dlg := NewBasicDialog(&nd.Dialog, i18n.Sprintf("NAT Discovery"), loadIcon(res.IconNat, 32),
		DataBinder{}, nil,
		VSpacer{Size: 1},
		Composite{
			Layout: HBox{MarginsZero: true},
			Children: []Widget{
				Label{Text: i18n.SprintfColon("STUN Server")},
				TextEdit{Text: nd.serverAddr, ReadOnly: true, CompactHeight: true},
			},
		},
		VSpacer{Size: 1},
		TableView{
			Name:     "tb",
			Visible:  false,
			AssignTo: &nd.table,
			Columns: []TableViewColumn{
				{Title: i18n.Sprintf("Item"), DataMember: "Title", Width: 180},
				{Title: i18n.Sprintf("Value"), DataMember: "Value", Width: 180},
			},
			ColumnsOrderable: false,
		},
		ProgressBar{AssignTo: &nd.barView, Visible: Bind("!tb.Visible"), MarqueeMode: true},
		VSpacer{},
	)
	dlg.MinSize = Size{Width: 400, Height: 350}
	if err := dlg.Create(owner); err != nil {
		return 0, err
	}
	nd.barView.SetFocus()
	nd.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		nd.closed = true
	})

	// Start discovering NAT type
	go nd.discover()

	return nd.Dialog.Run(), nil
}

func (nd *NATDiscoveryDialog) discover() (err error) {
	defer nd.Synchronize(func() {
		if err != nil && !nd.closed {
			nd.barView.SetMarqueeMode(false)
			showError(err, nd.Form())
			nd.Cancel()
		}
	})
	addrs, localAddr, err := nathole.Discover([]string{nd.serverAddr}, "")
	if err != nil {
		return err
	}
	if len(addrs) < 2 {
		return fmt.Errorf("can not get enough addresses")
	}

	localIPs, _ := nathole.ListLocalIPsForNatHole(10)

	natFeature, err := nathole.ClassifyNATFeature(addrs, localIPs)
	if err != nil {
		return err
	}
	items := []*ListItem{
		{Title: i18n.Sprintf("NAT Type"), Value: natFeature.NatType},
		{Title: i18n.Sprintf("Behavior"), Value: natFeature.Behavior},
		{Title: i18n.Sprintf("Local Address"), Value: localAddr.String()},
	}
	for _, addr := range addrs {
		items = append(items, &ListItem{
			Title: i18n.Sprintf("External Address"),
			Value: addr,
		})
	}
	var public string
	if natFeature.PublicNetwork {
		public = i18n.Sprintf("Yes")
	} else {
		public = i18n.Sprintf("No")
	}
	items = append(items, &ListItem{
		Title: i18n.Sprintf("Public Network"),
		Value: public,
	})
	nd.table.Synchronize(func() {
		nd.table.SetVisible(true)
		if err = nd.table.SetModel(NewNonSortedModel(items)); err != nil {
			showError(err, nd.Form())
		}
	})
	return nil
}
