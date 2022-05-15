@echo off
setlocal enabledelayedexpansion
set BUILDDIR=%~dp0
set PATH=%BUILDDIR%.deps;%PATH%
cd /d %BUILDDIR% || exit /b 1

for /f "tokens=3" %%a in ('findstr /r "Number.*=.*[0-9.]*" .\pkg\version\version.go') do set VERSION=%%a
set VERSION=%VERSION:"=%

:render
	echo [+] Rendering icons
	for %%a in ("icon\*.svg") do convert -background none "%%~fa" -define icon:auto-resize="256,192,128,96,64,48,40,32,24,20,16" -compress zip "%%~dpna.ico" || goto :error

:resources
	echo [+] Assembling resources
	windres -DVERSION_ARRAY=%VERSION:.=,%,0 -DVERSION_STR=%VERSION% -i cmd/frpmgr/resources.rc -o cmd/frpmgr/rsrc.syso -O coff -c 65001 || goto :error

:packages
	echo [+] Downloading packages
	go mod tidy || goto :error

:patch
	echo [+] Patching files
	for %%f in (patches\*.patch) do patch -N -r - -d %GOPATH% -p0 < %%f

:build
	echo [+] Building program
	for /f "usebackq tokens=1,2 delims==" %%i in (`wmic os get LocalDateTime /VALUE 2^>NUL`) do if '.%%i.'=='.LocalDateTime.' set ldt=%%j
	set BUILD_DATE=%ldt:~0,4%-%ldt:~4,2%-%ldt:~6,2%
	set MOD=github.com/koho/frpmgr
	set GO111MODULE=on
	set CGO_ENABLED=0
	for /f "tokens=2 delims=@" %%y in ('go mod graph ^| findstr %MOD% ^| findstr frp@') do (set FRP_VERSION=%%y)
	go build -trimpath -ldflags="-H windowsgui -s -w -X %MOD%/pkg/version.FRPVersion=%FRP_VERSION:~1% -X %MOD%/pkg/version.BuildDate=%BUILD_DATE%" -o bin/frpmgr.exe ./cmd/frpmgr || goto :error

:installer
	echo [+] Building installer
	call installer/build.bat %VERSION% || goto :error

:success
	echo [+] Success
	exit /b 0

:error
	echo [-] Failed with error %errorlevel%.
	exit /b %errorlevel%
