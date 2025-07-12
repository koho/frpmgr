# FRP 管理器

[![GitHub Release](https://img.shields.io/github/tag/koho/frpmgr.svg?label=release)](https://github.com/koho/frpmgr/releases)
[![FRP Version](https://img.shields.io/endpoint?url=https%3A%2F%2Fgo.shields.workers.dev%2Fkoho%2Ffrpmgr%2Fmaster%3Fname%3Dfrp)](https://github.com/fatedier/frp)
[![GitHub Downloads](https://img.shields.io/github/downloads/koho/frpmgr/total.svg)](https://github.com/koho/frpmgr/releases)

[English](README.md) | 简体中文

FRP 管理器是一个多节点、图形化反向代理工具，专为 Windows 上的 [FRP](https://github.com/fatedier/frp) 设计。它允许用户轻松设置反向代理，而无需编写配置文件。FRP 管理器提供了一套完整的解决方案，包括编辑器、启动器、状态跟踪和热重载。

该工具的灵感来自于一个常见的用例，我们经常需要组合使用多种工具，包括客户端、配置文件和启动器，以创建一个稳定的服务，将位于 NAT 或防火墙后的本地服务器暴露到互联网。现在，有了 FRP 管理器这个一体化解决方案，您可以在部署反向代理时省去许多繁琐的操作。

最新版本至少需要 Windows 10 或 Server 2016。请访问 **[Wiki](https://github.com/koho/frpmgr/wiki)** 获取完整指南。

![screenshot](/docs/screenshot_zh.png)

## 特征

- **界面可退出：**&#8203;所有已启动的配置都将作为后台服务独立运行，因此您可以在完成所有设置后关闭界面。
- **开机自启：**&#8203;已启动的配置默认注册为自动启动服务，并在系统启动时自动启动（无需登录）。
- **热重载：**&#8203;允许用户将代理更改应用于正在运行的配置，而无需重启服务，也不会丢失代理状态。
- **多配置文件管理：**&#8203;通过创建多个配置，可以轻松连接到多个节点。
- **导入和导出配置：**&#8203;提供从多个来源导入配置文件的选项，包括本地文件、剪贴板和 HTTP。
- **自毁配置：**&#8203;一种特殊配置，会在指定的时间后删除并无法访问。
- **状态跟踪：**&#8203;您可以直接在表格视图中查看代理状态，而无需查看日志。

访问 **[Wiki](https://github.com/koho/frpmgr/wiki)** 获取完整指南，包括：

- **[安装说明](https://github.com/koho/frpmgr/wiki#how-to-install)：**&#8203;在 Windows 上安装或升级 FRP 管理器。
- **[快速入门指南](https://github.com/koho/frpmgr/wiki/Quick-Start)：**&#8203;了解如何在几分钟内连接到您的节点并设置代理。
- **[配置](https://github.com/koho/frpmgr/wiki/Configuration)：**&#8203;探索配置、代理、访问者和日志。
- **[示例](https://github.com/koho/frpmgr/wiki/Examples)：**&#8203;这里有一些常见的示例可以帮助您学习 FRP 管理器。

## 构建

要从源代码构建 FRP 管理器，您需要安装以下依赖项：

- Go
- [Windows SDK](https://developer.microsoft.com/en-us/windows/downloads/windows-sdk/)
- [MinGW](https://github.com/mstorsjo/llvm-mingw)
- [WiX Toolset](https://wixtoolset.org/) v3.14

安装完成后，您需要设置 `WindowsSdkVerBinPath` 环境变量，以指示构建脚本在哪里找到特定版本的 Windows SDK，例如 `set WindowsSdkVerBinPath=C:\Program Files (x86)\Windows Kits\10\bin\10.0.26100.0`。您还需要将 MinGW 的 `bin` 目录添加到 `PATH` 环境变量中。

您可以通过打开终端来编译项目：

```shell
git clone https://github.com/koho/frpmgr
cd frpmgr
build.bat
```

生成的安装文件位于 `bin` 目录。

如果您想跳过构建安装包，可以在 `build` 命令中添加 `-p` 选项来获得一个便携软件：

```shell
build.bat -p
```

在这种情况下，您只需安装 Go 和 MinGW 即可。

### 调试

如果您是首次构建项目，则需要编译资源：

```shell
go generate
```

除非项目资源发生变化，否则无需再次执行该命令。

之后，即可直接运行该应用程序：

```shell
go run ./cmd/frpmgr
```

## 赞助商

> 我们非常感谢所有为项目发展而努力的用户、贡献者和赞助者。同时也感谢这些公司/组织为我们提供服务。

1. SignPath Foundation 为我们提供免费的代码签名！
<p align=center>
	<a href="https://about.signpath.io/">
 		<img src="./docs/sponsor_signpath.png" alt="SignPath Logo" height=50 />
	</a>
</p>

## 代码签名政策

免费代码签名由 [SignPath.io](https://about.signpath.io/) 提供，证书由 [SignPath Foundation](https://signpath.org/) 提供。

团队角色：

- 提交者和审阅者：[团队成员](https://github.com/koho/frpmgr/graphs/contributors)
- 审批者：[所有者](https://github.com/koho)

请阅读我们的完整[隐私政策](#隐私政策)。

## 隐私政策

除非得到用户、安装或操作人员的许可，否则该程序不会将任何信息传输到其他联网系统。

FRP 管理器集成了以下服务以实现附加功能，您可以随时在设置中启用或禁用这些服务：

- [api.github.com](https://docs.github.com/en/site-policy/privacy-policies/github-general-privacy-statement)（检查程序更新）

## 捐助

如果本项目对您有帮助，请考虑通过以下方式支持其开发：

- [**微信**](/docs/donate-wechat.jpg)
