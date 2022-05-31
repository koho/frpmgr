@echo off
setlocal enabledelayedexpansion
set VERSION=%1
set BUILDDIR=%~dp0
cd /d %BUILDDIR% || exit /b 1

if "%WIX%"=="" (
	echo ERROR: WIX was not found.
	exit /b 1
)

:build
	if not exist build md build
	call vcvars64.bat
	msbuild actions\actions.sln /t:Rebuild /p:Configuration=Release /p:Platform="x64" || goto :error
	copy actions\actions\bin\x64\Release\actions.CA.dll build\actions.dll /y || goto :error

	set MSI_FILE=build/frpmgr-%VERSION%.msi
	set WIX_CANDLE_FLAGS=-dVERSION=%VERSION%
	set WIX_LIGHT_FLAGS=-ext "%WIX%bin\WixUtilExtension.dll" -ext "%WIX%bin\WixUIExtension.dll" -sval
	set WIX_OBJ=build/frpmgr.wixobj
	"%WIX%bin\candle" %WIX_CANDLE_FLAGS% -out %WIX_OBJ% -arch x64 msi\frpmgr.wxs || goto :error
	"%WIX%bin\light" %WIX_LIGHT_FLAGS% -cultures:en-US -loc msi\en-US.wxl -out %MSI_FILE% %WIX_OBJ% || goto :error
	for %%l in (zh-CN zh-TW ja-JP ko-KR) do (
		set WIX_LANG_MSI=%MSI_FILE:~0,-4%_%%l.msi
		"%WIX%bin\light" %WIX_LIGHT_FLAGS% -cultures:%%l -loc msi\%%l.wxl -out !WIX_LANG_MSI! %WIX_OBJ% || goto :error
		for /f "tokens=3 delims=><" %%a in ('findstr /r "Id.*=.*Language" msi\%%l.wxl') do set LANG_CODE=%%a
		"%WIX%bin\torch" -t language %MSI_FILE% !WIX_LANG_MSI! -out build\!LANG_CODE! || goto :error
		"%WindowsSdkVerBinPath%x86\MsiDb" -d %MSI_FILE% -r build\!LANG_CODE! || goto :error
	)
	windres -DVERSION_ARRAY=%VERSION:.=,%,0 -DVERSION_STR=%VERSION% -DMSI_FILE=%MSI_FILE% -i setup/resources.rc -o build/rsrc.o -O coff -c 65001 || goto :error
	cl /Fe..\bin\frpmgr-%VERSION%-Setup.exe /Fobuild\setup.obj /utf-8 setup\setup.c /link /subsystem:windows build\rsrc.o shlwapi.lib msi.lib user32.lib advapi32.lib

:success
	exit /b 0

:error
	exit /b %errorlevel%
