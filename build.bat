@echo off
setlocal enabledelayedexpansion
set BUILDDIR=%~dp0
set ARCHS=amd64 386
cd /d %BUILDDIR% || exit /b 1

:packages
	echo [+] Downloading packages
	go mod tidy || goto :error

:resources
	echo [+] Generating resources
	for /f %%a in ('go generate') do set %%a
	if not defined VERSION exit /b 1

:build
	echo [+] Building program
	set MOD=github.com/koho/frpmgr
	set GO111MODULE=on
	set CGO_ENABLED=0
	for %%a in (%ARCHS%) do (
		set GOARCH=%%a
		go build -trimpath -ldflags="-H windowsgui -s -w -X %MOD%/pkg/version.BuildDate=%BUILD_DATE%" -o bin\x!GOARCH:~-2!\frpmgr.exe .\cmd\frpmgr || goto :error
	)

:archive
	echo [+] Creating archives
	for %%a in (%ARCHS%) do (
		set ARCH=%%a
		tar -ac -C bin\x!ARCH:~-2! -f bin\frpmgr-%VERSION%-x!ARCH:~-2!.zip frpmgr.exe || goto :error
	)

:installer
	echo [+] Building installer
	call installer\build.bat %VERSION% || goto :error

:success
	echo [+] Success
	exit /b 0

:error
	echo [-] Failed with error %errorlevel%.
	exit /b %errorlevel%
