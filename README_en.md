# FRP Manager

[![Build status](https://img.shields.io/github/workflow/status/koho/frpmgr/Releaser)](https://github.com/koho/frpmgr/actions/workflows/releaser.yml)
[![GitHub release](https://img.shields.io/github/tag/koho/frpmgr.svg?label=release)](https://github.com/koho/frpmgr/releases)

[简体中文](README.md) | English

A user-friendly desktop GUI client for [frp](https://github.com/fatedier/frp) on Windows.

![screenshot](/docs/screenshot_en.png)

System requirements: Windows 7 and above

Instructions for use:
1. All started configurations will run independently in the form of a background service. **Close the GUI does not affect the running state of the configuration**.
2. The configuration that has been started, **the next system boot will automatically start**, unless the automatic start is manually disabled.
3. After modifying the configuration through the GUI, the instance of the configuration will be automatically restarted.
4. After manually stopping the configuration, the configuration will not start automatically.

## :sparkles: Features

* :pencil2: Simple GUI for quick configuration
* :play_or_pause_button: Quick enable/disable proxy entry
* 📚 Multiple configuration files management
* :inbox_tray: Support import/export configuration files
* :computer: Auto-start at system boot

## :gear: Build

#### Install dependencies
- go >=1.16
- Visual Studio
- [MinGW](https://www.mingw-w64.org/)
- [WiX Toolset](https://wixtoolset.org/)

#### Setup environment

1. Add `vcvars64.bat` to the `PATH` environment variable. Usually the directory is:
- `C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build`

2. Make sure the `bin` directory of MinGW is added to the `PATH` environment variable.

3. Make sure the environment variable `WIX` is set to the Wix installation directory.

#### Compile

```shell
git clone https://github.com/koho/frpmgr
cd frpmgr
build.bat
```

Find the generated installation files in the `bin` directory.

#### Debugging

The build process needs to render icons to generate resources and patch some files, all of which are done in `build.bat`. 
Therefore, this script needs to be run once before the first debugging, and the subsequent debugging does not need to execute the script again.

## Donation

If this project helps you a lot, you can [support us](/docs/donate-wechat.jpg).
