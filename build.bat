@echo off
setlocal enabledelayedexpansion
set GOARCH_x64=amd64
set GOARCH_x86=386
set BUILDDIR=%~dp0
cd /d %BUILDDIR% || exit /b 1

if "%~1" == "-p" (
	set TARGET=%~2
) else (
	set TARGET=%~1
)

if "%TARGET%" == "" set TARGET=x64 x86

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
	for %%a in (%TARGET%) do (
		if defined GOARCH_%%a (
			set GOARCH=!GOARCH_%%a!
		) else (
			set GOARCH=%%a
		)
		go build -trimpath -ldflags="-H windowsgui -s -w -X %MOD%/pkg/version.BuildDate=%BUILD_DATE%" -o bin\%%a\frpmgr.exe .\cmd\frpmgr || goto :error
	)

if "%~1" == "-p" goto :success

:installer
	echo [+] Building installer
	for %%a in (%TARGET%) do (
		call installer\build.bat %VERSION% %%a || goto :error
	)

:success
	echo [+] Success
	exit /b 0

:error
	echo [-] Failed with error %errorlevel%.
	exit /b %errorlevel%
