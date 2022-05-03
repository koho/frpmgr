@echo off
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
	set WIX_LIGHT_FLAGS=-ext "%WIX%bin\WixUtilExtension.dll" -ext "%WIX%bin\WixUIExtension.dll" -cultures:zh-CN
	set WIX_OBJ=build/frpmgr.wixobj
	"%WIX%bin\candle" %WIX_CANDLE_FLAGS% -out %WIX_OBJ% -arch x64 frpmgr.wxs || goto :error
	"%WIX%bin\light" %WIX_LIGHT_FLAGS% -out %MSI_FILE% %WIX_OBJ% || goto :error

	windres -DVERSION_ARRAY=%VERSION:.=,%,0 -DVERSION_STR=%VERSION% -DMSI_FILE=%MSI_FILE% -i setup/resources.rc -o build/rsrc.o -O coff -c 65001 || goto :error
	cl /Fe..\bin\frpmgr-%VERSION%-Setup.exe /Fobuild\setup.obj setup\setup.c /link /subsystem:windows build\rsrc.o shlwapi.lib msi.lib user32.lib advapi32.lib

:success
	exit /b 0

:error
	exit /b %errorlevel%
