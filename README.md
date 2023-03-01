# FRP Manager

[![GitHub release](https://img.shields.io/github/tag/koho/frpmgr.svg?label=release)](https://github.com/koho/frpmgr/releases)
[![Frp version](https://img.shields.io/endpoint?url=https%3A%2F%2Fgo.shields.workers.dev%2Fkoho%2Ffrpmgr%2Fmaster%3Fname%3Dfrp)](https://github.com/fatedier/frp)

简体中文 | [English](README_en.md)

Windows 下的 [frp](https://github.com/fatedier/frp) 图形界面客户端。

![screenshot](/docs/screenshot_zh.png)

系统需求：Windows 7 及以上版本

使用说明：

1. 启动配置将以后台服务的形式独立运行，**关闭界面并不影响配置的运行**
2. 已启动的配置，**下次开机会自动启动**，除非手动禁用自启
3. 通过界面修改配置后，会自动重启该配置的实例
4. 手动停止配置后，该配置将不会开机自启

## :sparkles: 特征

* :pencil2: 简易的编辑界面，快速完成配置
* :play_or_pause_button: 快捷启用/禁用代理条目
* 📚 多配置文件管理
* :inbox_tray: 支持导入/导出配置文件
* :computer: 开机自启动
* :lock: 支持密码保护

## :gear: 构建

#### 安装依赖

- Go
- Visual Studio
- [MinGW](https://www.mingw-w64.org/)
- [WiX Toolset](https://wixtoolset.org/)

#### 环境配置

1. 把 `vcvars64.bat` 添加到环境变量。通常目录为：

   `C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build`

2. 确保 MinGW 的 `bin` 目录已添加到环境变量

3. 确保环境变量 `WIX` 已设置为 Wix 的安装目录

#### 编译项目

```shell
git clone https://github.com/koho/frpmgr
cd frpmgr
build.bat
```

在 `bin` 目录找到生成的安装文件。

#### 调试

编译程序需要渲染图标生成资源和对某些文件打补丁，这些操作都在 `build.bat` 完成。因此在第一次调试前需要运行一次此脚本，后续调试不需要再执行该脚本。

## 捐助

如果您觉得本项目对你有帮助，欢迎给予我们[支持](/docs/donate-wechat.jpg)。
