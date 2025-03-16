package ui

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/samber/lo"

	"github.com/koho/frpmgr/i18n"
	"github.com/koho/frpmgr/pkg/res"
	"github.com/koho/frpmgr/pkg/util"
)

type URLImportDialog struct {
	*walk.Dialog

	db        *walk.DataBinder
	viewModel urlImportViewModel

	// Views
	statusText *walk.Label

	// Items contain the downloaded data from URLs
	Items []URLConf
}

type urlImportViewModel struct {
	URLs    string
	Working bool
}

// URLConf provides config data downloaded from URL
type URLConf struct {
	// Filename is the name of the downloaded file
	Filename string
	// Zip defines whether the Data is a zip file
	Zip bool
	// Downloaded raw Data from URL
	Data []byte
}

func NewURLImportDialog() *URLImportDialog {
	return &URLImportDialog{Items: make([]URLConf, 0)}
}

func (ud *URLImportDialog) Run(owner walk.Form) (int, error) {
	return NewBasicDialog(&ud.Dialog, i18n.Sprintf("Import from URL"), loadIcon(res.IconURLImport, 32),
		DataBinder{AssignTo: &ud.db, DataSource: &ud.viewModel, Name: "vm"}, ud.onImport,
		Label{Text: i18n.Sprintf("* Support batch import, one link per line.")},
		TextEdit{
			Enabled: Bind("!vm.Working"),
			Text:    Bind("URLs", res.ValidateNonEmpty),
			VScroll: true,
			MinSize: Size{Width: 430, Height: 130},
		},
		Label{
			AssignTo:     &ud.statusText,
			Text:         fmt.Sprintf("%s: %s", i18n.Sprintf("Status"), i18n.Sprintf("Ready")),
			EllipsisMode: EllipsisEnd,
		},
		VSpacer{Size: 4},
	).Run(owner)
}

func (ud *URLImportDialog) onImport() {
	if err := ud.db.Submit(); err != nil {
		return
	}
	urls := strings.Split(ud.viewModel.URLs, "\n")
	urls = lo.FilterMap(urls, func(s string, i int) (string, bool) {
		s = strings.TrimSpace(s)
		return s, s != ""
	})
	if len(urls) == 0 {
		showWarningMessage(ud.Form(),
			i18n.Sprintf("Import Config"),
			i18n.Sprintf("Please enter the correct URL list."))
		return
	}
	ud.viewModel.Working = true
	ud.DefaultButton().SetEnabled(false)
	ud.db.Reset()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	ud.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		cancel()
		wg.Wait()
	})
	go ud.urlImport(ctx, &wg, urls)
}

func (ud *URLImportDialog) urlImport(ctx context.Context, wg *sync.WaitGroup, urls []string) {
	result := walk.DlgCmdOK
	defer func() { ud.Close(result) }()
	defer wg.Done()
	for i, url := range urls {
		ud.statusText.SetText(fmt.Sprintf("%s: [%d/%d] %s %s",
			i18n.Sprintf("Status"), i+1, len(urls), i18n.Sprintf("Download"), url,
		))
		filename, mediaType, data, err := util.DownloadFile(ctx, url)
		if errors.Is(err, context.Canceled) {
			result = walk.DlgCmdCancel
			return
		} else if err != nil {
			showError(err, ud.Form())
			continue
		}
		ud.Items = append(ud.Items, URLConf{
			Filename: filename,
			Zip:      mediaType == "application/zip" || strings.ToLower(filepath.Ext(filename)) == ".zip",
			Data:     data,
		})
	}
}
