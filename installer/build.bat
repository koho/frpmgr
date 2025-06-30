@echo off
setlocal enabledelayedexpansion
set VERSION=%~1
set ARCH=%~2
set STEP="%~3"
set BUILDDIR=%~dp0
cd /d %BUILDDIR% || exit /b 1

if "%VERSION%" == "" (
	echo ERROR: no version provided.
	exit /b 1
)

if "%ARCH%" == "" (
	echo ERROR: no architecture provided.
	exit /b 1
)

:build
	if not exist build md build
	set PLAT_DIR=build\%ARCH%
	set SETUP_FILENAME=frpmgr-%VERSION%-setup-%ARCH%.exe
	if %STEP% == "dist" goto :dist
	call vcvarsall.bat %ARCH%
	if not exist %PLAT_DIR% md %PLAT_DIR%
	set MSI_FILE=%PLAT_DIR%\frpmgr.msi
	if %STEP:"actions"=""% == "" call :build_actions
	if %STEP:"msi"=""% == "" call :build_msi
	if %STEP:"setup"=""% == "" call :build_setup
	if %STEP% == "" goto :dist

:success
	exit /b 0

:build_actions
	rc /DVERSION_ARRAY=%VERSION:.=,% /DVERSION_STR=%VERSION% /Fo %PLAT_DIR%\actions.res actions\version.rc || goto :error
	cl /O2 /LD /MD /DNDEBUG /Fe%PLAT_DIR%\actions.dll /Fo%PLAT_DIR%\actions.obj actions\actions.c %PLAT_DIR%\actions.res msi.lib shell32.lib advapi32.lib shlwapi.lib ole32.lib || goto :error
	goto :eof

:build_msi
	if not defined WIX (
		echo ERROR: WIX was not found.
		exit /b 1
	)
	set WIX_CANDLE_FLAGS=-dVERSION=%VERSION%
	set WIX_LIGHT_FLAGS=-ext "%WIX%bin\WixUtilExtension.dll" -ext "%WIX%bin\WixUIExtension.dll" -sval
	set WIX_OBJ=%PLAT_DIR%\frpmgr.wixobj
	"%WIX%bin\candle" %WIX_CANDLE_FLAGS% -out %WIX_OBJ% -arch %ARCH% msi\frpmgr.wxs || goto :error
	"%WIX%bin\light" %WIX_LIGHT_FLAGS% -cultures:en-US -loc msi\en-US.wxl -out %MSI_FILE% %WIX_OBJ% || goto :error
	for %%l in (zh-CN zh-TW ja-JP ko-KR es-ES) do (
		set WIX_LANG_MSI=%MSI_FILE:~0,-4%_%%l.msi
		"%WIX%bin\light" %WIX_LIGHT_FLAGS% -cultures:%%l -loc msi\%%l.wxl -out !WIX_LANG_MSI! %WIX_OBJ% || goto :error
		for /f "tokens=3 delims=><" %%a in ('findstr /r "Id.*=.*Language" msi\%%l.wxl') do set LANG_CODE=%%a
		"%WindowsSdkVerBinPath%x86\MsiTran" -g %MSI_FILE% !WIX_LANG_MSI! %PLAT_DIR%\!LANG_CODE! || goto :error
		"%WindowsSdkVerBinPath%x86\MsiDb" -d %MSI_FILE% -r %PLAT_DIR%\!LANG_CODE! || goto :error
	)
	goto :eof

:build_setup
	rc /DFILENAME=%SETUP_FILENAME% /DVERSION_ARRAY=%VERSION:.=,% /DVERSION_STR=%VERSION% /DMSI_FILE=%MSI_FILE:\=\\% /Fo %PLAT_DIR%\setup.res setup\resource.rc || goto :error
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
	cl /O2 /MD /DUPGRADE_CODE=L\"{%UPGRADE_CODE%}\" /DNDEBUG /Fe%PLAT_DIR%\setup.exe /Fo%PLAT_DIR%\setup.obj setup\setup.c /link /subsystem:windows %PLAT_DIR%\setup.res shlwapi.lib msi.lib user32.lib advapi32.lib ole32.lib || goto :error
	goto :eof

:dist
	echo [+] Creating %ARCH% archives
	tar -ac -C ..\bin\%ARCH% -f ..\bin\frpmgr-%VERSION%-%ARCH%.zip frpmgr.exe || goto :error
	echo [+] Creating %ARCH% installer
	copy %PLAT_DIR%\setup.exe ..\bin\%SETUP_FILENAME% /y

:error
	exit /b %errorlevel%
