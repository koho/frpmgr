@echo off
setlocal enabledelayedexpansion
set VERSION=%1
set BUILDDIR=%~dp0
set RESx64=pe-x86-64
set RESx86=pe-i386
cd /d %BUILDDIR% || exit /b 1

if "%WIX%"=="" (
	echo ERROR: WIX was not found.
	exit /b 1
)

:build
	if not exist build md build
	for %%a in (%ARCHS%) do (
		set PLAT=%%a
		call :build_plat x!PLAT:~-2! || goto :error
	)

:success
	exit /b 0

:build_plat
	set ARCH=%~1
	call vcvarsall.bat %ARCH%
	set PLAT_DIR=build\%ARCH%
	if not exist %PLAT_DIR% md %PLAT_DIR%

	msbuild actions\actions.sln /t:Rebuild /p:Configuration=Release /p:Platform="%ARCH%" || goto :error
	copy actions\actions\bin\%ARCH%\Release\actions.CA.dll %PLAT_DIR%\actions.dll /y || goto :error

	set MSI_FILE=%PLAT_DIR%\frpmgr-%VERSION%.msi
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
	windres -DARCH=%ARCH% -DVERSION_ARRAY=%VERSION:.=,%,0 -DVERSION_STR=%VERSION% -DMSI_FILE=%MSI_FILE:\=\\% -i setup\resource.rc -o %PLAT_DIR%\rsrc.o -O coff -c 65001 -F !RES%ARCH%! || goto :error
	cl /Fe..\bin\frpmgr-%VERSION%-setup-%ARCH%.exe /Fo%PLAT_DIR%\setup.obj /utf-8 setup\setup.c /link /subsystem:windows %PLAT_DIR%\rsrc.o shlwapi.lib msi.lib user32.lib advapi32.lib
	goto :eof

:error
	exit /b %errorlevel%
