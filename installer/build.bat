@echo off
set FRPMGR_VERSION=%1
set BUILDDIR=%~dp0
cd /d %BUILDDIR% || exit /b 1
if "%WIX%"=="" (
    echo ERROR: WIX was not found.
    exit /b 1
)
if not exist build md build
call vcvars64.bat
msbuild actions/actions.sln /t:Rebuild /p:Configuration=Release /p:Platform="x64" || exit /b 1
copy actions\actions\bin\x64\Release\actions.CA.dll build\actions.dll /y || exit /b 1
set MSI_FILE=build\\frpmgr-%FRPMGR_VERSION%.msi
set WIX_CANDLE_FLAGS=-nologo -dFRPMGR_VERSION=%FRPMGR_VERSION%
set WIX_LIGHT_FLAGS=-nologo -spdb -ext "%WIX%bin\\WixUtilExtension.dll" -ext "%WIX%bin\\WixUIExtension.dll" -cultures:zh-CN
"%WIX%bin\candle" %WIX_CANDLE_FLAGS% -out build\ -arch x64 frpmgr.wxs || exit /b 1
"%WIX%bin\light" %WIX_LIGHT_FLAGS% -out %MSI_FILE% "build\\frpmgr.wixobj" || exit /b 1
pushd setup
windres -DVERSION_ARRAY=%FRPMGR_VERSION:.=,% -DVERSION_STR=%FRPMGR_VERSION% -DMSI_FILE=..\\%MSI_FILE% -DAPP_ICO=..\\..\\icon\\app_install.ico -i resources.rc -o ..\\build\\rsrc.o -O coff -c 65001 || (popd & exit /b 1)
popd
cl /Fe..\bin\frpmgr-%FRPMGR_VERSION%-Setup.exe /Fobuild/setup.obj setup/setup.c /link /subsystem:windows build/rsrc.o shlwapi.lib msi.lib user32.lib advapi32.lib
