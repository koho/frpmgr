package frpmgr

//go:generate go run resource.go
//go:generate powershell -Command "attrib -r $(Join-Path -Path $(go list -m -f '{{.Dir}}' github.com/fatedier/frp) -ChildPath web\\frpc\\*.*) /s /d"
//go:generate powershell -Command "mingw32-make -C $(Join-Path -Path $(go list -m -f '{{.Dir}}' github.com/fatedier/frp) -ChildPath web\\frpc) install build"
