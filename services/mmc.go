package services

import (
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

type mmcApp struct {
	handle   *ole.IDispatch
	document *ole.IDispatch
	view     *ole.IDispatch
	list     *ole.IDispatch
}

var mmc mmcApp

func setupMMC() {
	ole.CoInitialize(0)
	unknown, _ := oleutil.CreateObject("MMC20.Application")
	mmc.handle, _ = unknown.QueryInterface(ole.IID_IDispatch)
}

func closeDocument() {
	if mmc.document != nil {
		oleutil.MustCallMethod(mmc.document, "Close", false)
		mmc.list.Release()
		mmc.view.Release()
		mmc.document.Release()
		mmc.document = nil
	}
}

func Cleanup() {
	if mmc.handle != nil {
		closeDocument()
		// Wait for the popup dialog to close by itself
		time.Sleep(2 * time.Second)
		// Exit MMC
		oleutil.CallMethod(mmc.handle, "Quit")
		mmc.handle.Release()
		mmc.handle = nil
		ole.CoUninitialize()
	}
}

// ShowPropertyDialog shows up the service property dialog of the given service
func ShowPropertyDialog(displayName string) {
	if mmc.handle == nil {
		setupMMC()
	} else {
		closeDocument()
	}
	// Load services
	oleutil.MustCallMethod(mmc.handle, "Load", "services.msc")
	mmc.document = oleutil.MustGetProperty(mmc.handle, "Document").ToIDispatch()
	mmc.view = oleutil.MustGetProperty(mmc.document, "ActiveView").ToIDispatch()
	mmc.list = oleutil.MustGetProperty(mmc.view, "ListItems").ToIDispatch()
	count := int(oleutil.MustGetProperty(mmc.list, "Count").Val)
	for i := 1; i <= count; i++ {
		item := oleutil.MustCallMethod(mmc.list, "Item", i).ToIDispatch()
		name := oleutil.MustGetProperty(item, "Name").ToString()
		if name == displayName {
			oleutil.MustCallMethod(mmc.view, "Select", item)
			oleutil.MustCallMethod(mmc.view, "DisplaySelectionPropertySheet")
			item.Release()
			return
		}
		item.Release()
	}
}
