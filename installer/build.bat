@echo off
set FRPMGR_VERSION=%1
set BUILDDIR=%~dp0
cd /d %BUILDDIR% || exit /b 1
if "%WIX%"=="" (
    echo ERROR: WIX was not found.
    exit /b 1
)
if not exist build md build
call VsMSBuildCmd.bat
msbuild actions/actions.sln /t:Rebuild /p:Configuration=Release /p:Platform="x64" || exit /b 1
copy actions\actions\bin\x64\Release\actions.CA.dll build\actions.dll /y || exit /b 1
set WIX_CANDLE_FLAGS=-nologo -dFRPMGR_VERSION=%FRPMGR_VERSION%
set WIX_LIGHT_FLAGS=-nologo -spdb -ext "%WIX%bin\\WixUtilExtension.dll" -ext "%WIX%bin\\WixUIExtension.dll" -cultures:zh-CN
"%WIX%bin\candle" %WIX_CANDLE_FLAGS% -out build\ -arch x64 frpmgr.wxs
"%WIX%bin\light" %WIX_LIGHT_FLAGS% -out "..\\bin\\frpmgr-%FRPMGR_VERSION%.msi" "build\\frpmgr.wixobj"