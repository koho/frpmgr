# FRP Manager

[![GitHub Release](https://img.shields.io/github/tag/koho/frpmgr.svg?label=release)](https://github.com/koho/frpmgr/releases)
[![FRP Version](https://img.shields.io/endpoint?url=https%3A%2F%2Fgo.shields.workers.dev%2Fkoho%2Ffrpmgr%2Fmaster%3Fname%3Dfrp)](https://github.com/fatedier/frp)
[![GitHub Downloads](https://img.shields.io/github/downloads/koho/frpmgr/total.svg)](https://github.com/koho/frpmgr/releases)

English | [简体中文](README_zh.md)

FRP Manager is a multi-node, graphical reverse proxy tool designed for [FRP](https://github.com/fatedier/frp) on Windows. It allows users to setup reverse proxy easily without writing the configuration file. FRP Manager offers a complete solution including editor, launcher, status tracking, and hot reload.

The tool was inspired by a common use case where we often need to combine multiple tools including client, configuration file, and launcher to create a stable service that exposes a local server behind a NAT or firewall to the Internet. Now, with FRP Manager, an all-in-one solution, you can avoid many tedious operations when deploying a reverse proxy.

The latest release requires at least Windows 10 or Server 2016. Please visit the **[Wiki](https://github.com/koho/frpmgr/wiki)** for comprehensive guides.

![screenshot](/docs/screenshot_en.png)

## Features

- **Closable GUI:** All launched configurations will run independently as background services, so you can close the GUI after finishing all settings.
- **Auto-start:** A launched configuration is registered as an auto-start service by default and starts automatically during system boot (no login required).
- **Hot reload:** Allows users to apply proxy changes to a running configuration without restarting the service and without losing proxy state.
- **Multiple configurations:** It's easy to connect to multiple nodes by creating multiple configurations.
- **Import and export configurations:** Provides the option to import configuration file from multiple sources, including local file, clipboard, and HTTP.
- **Self-destructing configuration:** A special configuration that disappears and becomes unreachable after a certain amount of time.
- **Status tracking:** You can check the proxy status directly in the table view without looking at the logs.

Visit the **[Wiki](https://github.com/koho/frpmgr/wiki)** for comprehensive guides, including:

- **[Installation Instructions](https://github.com/koho/frpmgr/wiki#how-to-install):** Install or upgrade FRP Manager on Windows.
- **[Quick Start Guide](https://github.com/koho/frpmgr/wiki/Quick-Start):** Learn how to connect to your node and setup a proxy in minutes.
- **[Configuration](https://github.com/koho/frpmgr/wiki/Configuration):** Explore configuration, proxy, visitor, and log.
- **[Examples](https://github.com/koho/frpmgr/wiki/Examples):** There are some common examples to help you learn FRP Manager.

## Building

To build FRP Manager from source, you need to install the following dependencies:

- Go
- Visual Studio
- [MinGW](https://www.mingw-w64.org/)
- [WiX Toolset](https://wixtoolset.org/) v3.14

Once Visual Studio is installed, add the [developer command file directory](https://learn.microsoft.com/en-us/cpp/build/building-on-the-command-line?view=msvc-170#developer_command_file_locations) (e.g., `C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build`) to the `PATH` environment variable. Likewise, do the same for the `bin` directory of MinGW.

You can compile the project by opening the terminal:

```shell
git clone https://github.com/koho/frpmgr
cd frpmgr
build.bat
```

The generated installation files are located in the `bin` directory.

### Debugging

If you're building the project for the first time, you need to compile resources:

```shell
go generate
```

The command does not need to be executed again unless the project's resources change.

After that, the application can be run directly:

```shell
go run ./cmd/frpmgr
```

## Donation

If this project is useful to you, consider supporting its development in one of the following ways:

- [**WeChat**](/docs/donate-wechat.jpg)
