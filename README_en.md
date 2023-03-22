# FRP Manager

[![GitHub release](https://img.shields.io/github/tag/koho/frpmgr.svg?label=release)](https://github.com/koho/frpmgr/releases)
[![Frp version](https://img.shields.io/endpoint?url=https%3A%2F%2Fgo.shields.workers.dev%2Fkoho%2Ffrpmgr%2Fmaster%3Fname%3Dfrp)](https://github.com/fatedier/frp)

[简体中文](README.md) | English

A user-friendly desktop GUI client for [frp](https://github.com/fatedier/frp) on Windows.

![screenshot](/docs/screenshot_en.png)

System requirements: Windows 7 and above

Instructions for use:

1. All started configurations will run independently in the form of a background service. **Close the GUI does not
   affect the running state of the configuration**.
2. The configuration that has been started, **the next system boot will automatically start**, unless the automatic
   start is manually disabled.
3. After modifying the configuration through the GUI, the instance of the configuration will be automatically restarted.
4. After manually stopping the configuration, the configuration will not start automatically.

## :sparkles: Features

* :pencil2: Simple GUI for quick configuration
* :play_or_pause_button: Quick enable/disable proxy entry
* 📚 Multiple configuration files management
* :inbox_tray: Support import/export configuration files
* :computer: Auto-start at system boot
* :lock: Support password protection
* :clock4: Support automatic deletion of configuration files

## :gear: Build

#### Install dependencies

- Go
- Visual Studio
- [MinGW](https://www.mingw-w64.org/)
- [WiX Toolset](https://wixtoolset.org/)

#### Setup environment

1. Add `vcvarsall.bat` to the `PATH` environment variable. Usually the directory is:

   `C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build`

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

When building the project for the first time, you need to render the icon and generate resources:

```shell
go generate
```

The above command does not need to be run for every build, it just needs to be executed again when the resource changes.

After the command is completed, the program can be run directly:

```shell
go run ./cmd/frpmgr
```

## Donation

If this project helps you a lot, you can [support us](/docs/donate-wechat.jpg).
