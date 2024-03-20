@echo off
setlocal enabledelayedexpansion
set BUILDDIR=%~dp0
set PATH=%BUILDDIR%.deps;%PATH%
set ARCHS=amd64 386
cd /d %BUILDDIR% || exit /b 1

for /f "tokens=3" %%a in ('findstr /r "Number.*=.*[0-9.]*" .\pkg\version\version.go') do set VERSION=%%a
set VERSION=%VERSION:"=%

:packages
	echo [+] Downloading packages
	go mod tidy || goto :error

:resources
	echo [+] Generating resources
	go generate || goto :error

:build
	echo [+] Building program
	for /f "usebackq tokens=1,2 delims==" %%i in (`wmic os get LocalDateTime /VALUE 2^>NUL`) do if '.%%i.'=='.LocalDateTime.' set ldt=%%j
	set BUILD_DATE=%ldt:~0,4%-%ldt:~4,2%-%ldt:~6,2%
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
