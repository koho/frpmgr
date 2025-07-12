@echo off
setlocal enabledelayedexpansion
set VERSION=%~1
set ARCH=%~2
set STEP="%~3"
set BUILDDIR=%~dp0
cd /d %BUILDDIR% || exit /b 1
set TARGET_x64=x86_64
set TARGET_x86=i686
set TARGET_arm64=aarch64

if "%VERSION%" == "" (
	echo ERROR: no version provided.
	exit /b 1
)

if "%ARCH%" == "" (
	echo ERROR: no architecture provided.
	exit /b 1
)

if not defined TARGET_%ARCH% (
	echo ERROR: unsupported architecture.
	exit /b 1
)

:build
	if not exist build md build
	set PLAT_DIR=build\%ARCH%
	set SETUP_FILENAME=frpmgr-%VERSION%-setup-%ARCH%.exe
	if %STEP% == "dist" goto :dist
	set CC=!TARGET_%ARCH%!-w64-mingw32-gcc
	set WINDRES=!TARGET_%ARCH%!-w64-mingw32-windres
	if not exist %PLAT_DIR% md %PLAT_DIR%
	set MSI_FILE=%PLAT_DIR%\frpmgr.msi
	if %STEP:"actions"=""% == "" call :build_actions || goto :error
	if %STEP:"msi"=""% == "" call :build_msi || goto :error
	if %STEP:"setup"=""% == "" call :build_setup || goto :error
	if %STEP% == "" goto :dist

:success
	exit /b 0

:build_actions
	%WINDRES% -DVERSION_ARRAY=%VERSION:.=,% -DVERSION_STR=%VERSION% -o %PLAT_DIR%\actions.res.obj -i actions\version.rc -O coff -c 65001 || exit /b 1
	set CFLAGS=-O3 -Wall -std=gnu11 -DWINVER=0x0601 -D_WIN32_WINNT=0x0601 -municode -DUNICODE -D_UNICODE -DNDEBUG
	set LDFLAGS=-shared -s -Wl,--kill-at -Wl,--major-os-version=6 -Wl,--minor-os-version=1 -Wl,--major-subsystem-version=6 -Wl,--minor-subsystem-version=1 -Wl,--tsaware -Wl,--dynamicbase -Wl,--nxcompat -Wl,--export-all-symbols
	set LDLIBS=-lmsi -lole32 -lshlwapi -lshell32 -ladvapi32
	%CC% %CFLAGS% %LDFLAGS% -o %PLAT_DIR%\actions.dll actions\actions.c %PLAT_DIR%\actions.res.obj %LDLIBS% || exit /b 1
	goto :eof

:build_msi
	if not defined WIX (
		echo ERROR: WIX was not found.
		exit /b 1
	)
	set WIX_CANDLE_FLAGS=-dVERSION=%VERSION%
	set WIX_LIGHT_FLAGS=-ext "%WIX%bin\WixUtilExtension.dll" -ext "%WIX%bin\WixUIExtension.dll" -sval
	set WIX_OBJ=%PLAT_DIR%\frpmgr.wixobj
	"%WIX%bin\candle" %WIX_CANDLE_FLAGS% -out %WIX_OBJ% -arch %ARCH% msi\frpmgr.wxs || exit /b 1
	"%WIX%bin\light" %WIX_LIGHT_FLAGS% -cultures:en-US -loc msi\en-US.wxl -out %MSI_FILE% %WIX_OBJ% || exit /b 1
	for %%l in (zh-CN zh-TW ja-JP ko-KR es-ES) do (
		set WIX_LANG_MSI=%MSI_FILE:~0,-4%_%%l.msi
		"%WIX%bin\light" %WIX_LIGHT_FLAGS% -cultures:%%l -loc msi\%%l.wxl -out !WIX_LANG_MSI! %WIX_OBJ% || exit /b 1
		for /f "tokens=3 delims=><" %%a in ('findstr /r "Id.*=.*Language" msi\%%l.wxl') do set LANG_CODE=%%a
		"%WindowsSdkVerBinPath%x86\MsiTran" -g %MSI_FILE% !WIX_LANG_MSI! %PLAT_DIR%\!LANG_CODE! || exit /b 1
		"%WindowsSdkVerBinPath%x86\MsiDb" -d %MSI_FILE% -r %PLAT_DIR%\!LANG_CODE! || exit /b 1
	)
	goto :eof

:build_setup
	%WINDRES% -DFILENAME=%SETUP_FILENAME% -DVERSION_ARRAY=%VERSION:.=,% -DVERSION_STR=%VERSION% -DMSI_FILE=%MSI_FILE:\=\\% -o %PLAT_DIR%\setup.res.obj -i setup\resource.rc -O coff -c 65001 || exit /b 1
	set ARCH_LINE=-1
	for /f "tokens=1 delims=:" %%a in ('findstr /n /r ".*=.*\"%ARCH%\"" msi\frpmgr.wxs') do set ARCH_LINE=%%a
	if %ARCH_LINE% lss 0 (
		echo ERROR: unsupported architecture.
		exit /b 1
	)
	for /f "tokens=1,5 delims=: " %%a in ('findstr /n /r "UpgradeCode.*=.*\"[0-9a-fA-F-]*\"" msi\frpmgr.wxs') do (
		if %%a gtr %ARCH_LINE% if not defined UPGRADE_CODE set UPGRADE_CODE=%%b
	)
	if not defined UPGRADE_CODE (
		echo ERROR: UpgradeCode was not found.
		exit /b 1
	)
	set CFLAGS=-O3 -Wall -std=gnu11 -DWINVER=0x0601 -D_WIN32_WINNT=0x0601 -municode -DUNICODE -D_UNICODE -DNDEBUG -DUPGRADE_CODE=L\"{%UPGRADE_CODE%}\" -DVERSION=L\"%VERSION%\"
	set LDFLAGS=-s -Wl,--major-os-version=6 -Wl,--minor-os-version=1 -Wl,--major-subsystem-version=6 -Wl,--minor-subsystem-version=1 -Wl,--tsaware -Wl,--dynamicbase -Wl,--nxcompat -mwindows
	set LDLIBS=-lmsi -lole32 -lshlwapi -ladvapi32 -luser32
	%CC% %CFLAGS% %LDFLAGS% -o %PLAT_DIR%\setup.exe setup\setup.c %PLAT_DIR%\setup.res.obj %LDLIBS% || exit /b 1
	goto :eof

:dist
	echo [+] Creating %ARCH% archives
	tar -ac -C ..\bin\%ARCH% -f ..\bin\frpmgr-%VERSION%-%ARCH%.zip frpmgr.exe || goto :error
	echo [+] Creating %ARCH% installer
	copy %PLAT_DIR%\setup.exe ..\bin\%SETUP_FILENAME% /y

:error
	exit /b %errorlevel%
