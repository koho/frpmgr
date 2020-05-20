@echo off
set FRPMGR_VERSION=1.0.0
set BUILDDIR=%~dp0
cd /d %BUILDDIR% || exit /b 1
call "C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\Common7\Tools\VsDevCmd.bat"
set SDK_PLATFORM=VS2017
cl /Zc:wchar_t /D "_UNICODE" /D "UNICODE" /nologo /O2 /W3 /GL /DNDEBUG /MD "-I%WIX%sdk\%SDK_PLATFORM%\inc" actions.cpp /Fo.\build\ /link /LTCG /NOLOGO "/LIBPATH:%WIX%sdk\%SDK_PLATFORM%\lib\x86" msi.lib wcautil.lib dutil.lib kernel32.lib user32.lib advapi32.lib Version.lib shell32.lib /DLL /DEF:"actions.def" /IMPLIB:.\build\actions.lib /OUT:.\build\actions.dll || exit /b 1
set WIX_CANDLE_FLAGS=-nologo -dFRPMGR_VERSION=%FRPMGR_VERSION%
set WIX_LIGHT_FLAGS=-nologo -spdb -ext "%WIX%bin\\WixUtilExtension.dll" -ext "%WIX%bin\\WixUIExtension.dll" -cultures:zh-CN
"%WIX%bin\candle" %WIX_CANDLE_FLAGS% -out build\ -arch x86 frpmgr.wxs
"%WIX%bin\light" %WIX_LIGHT_FLAGS% -out "..\\bin\\frpmgr-%FRPMGR_VERSION%.msi" "build\\frpmgr.wixobj"