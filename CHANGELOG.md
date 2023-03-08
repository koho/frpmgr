# Changelog[Deprecated]
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.11.0] - 2023-02-10

### What's Changed
* View old history log files by @koho in https://github.com/koho/frpmgr/pull/30
* Show item name in dialog title
* Show blue config name when config is set to start manually
* Set gray background color for disabled proxy
* Support go1.20
* Support `bandwidth_limit_mode` option

### Update
* FRP v0.47.0

**Full Changelog**: https://github.com/koho/frpmgr/compare/v1.10.1...v1.11.0

## [1.10.1] - 2023-01-10
### 新增
* 自定义 DNS 服务
* 支持降级安装

### 更新
* FRP 版本 0.46.1

## [1.10.0] - 2022-12-19
### 新增
* 认证支持 `oidc_scope` 参数
* 支持 quic 协议

### 优化
* 展示 `xtcp`、`stcp`、`sudp` 类型的访问者参数（#27）

### 更新
* FRP 版本 0.46.0

## [1.9.2] - 2022-10-27
### 更新
* FRP 版本 0.45.0

## [1.9.1] - 2022-07-11
### 更新
* FRP 版本 0.44.0

## [1.9.0] - 2022-07-01
### 新增
* 多语言支持（#19）
* 支持创建/导入分享链接
* 可创建某个配置的副本，避免参数的重复输入
* 从 URL 导入配置

### 优化
* 配置文件统一存放到 `profiles` 目录

## [1.8.1] - 2022-05-28
### 新增
* 新的代理参数「路由用户」(`route_by_http_user`)

### 更新
* FRP 版本 0.43.0

## [1.8.0] - 2022-05-15
### 新增
* 从剪贴板导入配置/代理
* 支持拖拽文件导入配置
* 在文件夹中显示配置

### 优化
* 减少安装包体积（-48%）
* 升级时默认选择上次安装的目录
* 导入文件前验证配置文件

## [1.7.2] - 2022-04-22
### 更新
* FRP 版本 0.42.0

## [1.7.1] - 2022-04-13
### 新增
* "快速添加"支持更多类型，如 FTP、文件服务等
* 快捷启用/禁用代理条目
* 新增 TLS、心跳、复用等配置选项
* 代理条目右键菜单新增"复制访问地址"功能

### 修复
* 修复 Win7 下无法打开服务窗口

### 优化
* 防止同一用户下 GUI 窗口多开
* 启动配置前验证配置文件
* 保存代理条目前验证代理条目
* 使用范围端口时自动添加前缀

## [1.7.0] - 2022-03-24
### 新增
* 支持全部代理类型(本次新增`sudp`, `http`, `https`, `tcpmux`)的图形化配置
* 新增插件编辑
* 新增负载均衡
* 新增健康检查
* 新增带宽限制，代理协议版本配置
* 代理项目表格新增了子域名，自定义域名，插件列
* 添加连接超时时间，心跳间隔时间配置
* 添加 pprof 开关

### 修复
* 修复在中文配置名下，打开服务按钮无反应的问题
* 修复随机名称按钮会生成相同名称问题
* 修复了小概率界面崩溃问题

### 优化
* 无法添加相同名称的代理
* 无法导入相同名称的配置，当以压缩包导入时，忽略同名配置导入
* 减少了不必要的 IO 查询
* 代理项目表格各列宽调整，以充分利用空间
* 手动指定日志文件后修改配置名不再自动改变日志路径配置
* 路径配置的输入框添加浏览文件按钮

### 更新
* FRP 版本 0.41.0

## [1.6.1] - 2022-03-07
### 优化
* 安装包改用 exe 格式，避免无法关闭占用程序
* 升级完成后自动重启之前运行的服务

## [1.6.0] - 2022-02-14
### 新增
* 配置编辑支持自定义参数(#12)
* 打开配置文件入口
* 项目编辑可生成随机名称
* 复制服务器地址入口
* 添加`connect_server_local_ip`，`http_proxy`，`user`编辑入口

### 优化
* 减少不必要的视图更新
* 优化系统缩放时的界面显示

### 更新
* FRP 版本 0.39.1

## [1.5.0] - 2022-01-05
### 更新
* FRP 版本 0.38.0

## [1.4.2] - 2021-09-08
### 新增
* 可单独设定配置的服务启动方式(手动/自动)(#9)

### 修复
* 修复某些情况下无法查看服务的异常

## [1.4.1] - 2021-09-07
### 新增
* 支持配置xtcp/stcp类型(#8)
* 添加自定义选项支持
* 查看服务属性入口(#9)

### 更新
* FRP 版本 0.37.1

## [1.4.0] - 2021-07-12
### 修复
* 修复日志文件的卸载错误提示

### 更新
* FRP 版本 0.37.0

## [1.3.2] - 2020-12-16
### 新增
* 支持双击编辑

### 优化
* 小幅UI优化

## [1.3.1] - 2020-12-16
### 新增
* 添加文件版本信息

### 修复
* 修复卸载程序时的DLL错误

## [1.3.0] - 2020-12-13
### 新增
* 添加关于页面
* 支持导出配置文件

### 优化
* 日志实时显示
* 小幅UI优化

### 修复
* 修复卸载时日志文件无法删除的问题

## [1.2.5] - 2020-12-03
### 优化
* 小幅 UI 逻辑优化
* 相关日志文件重命名/删除

### 修复
* 修复 Windows 7 下的闪退问题(#2)

## [1.2.4] - 2020-08-17
### 新增
* 添加自定义DNS服务器的支持，对于使用动态DNS的服务器可以减少离线时间

### 修复
* 修复了一些编译错误

## [1.2.3] - 2020-05-24
### 修复
* 解决某些情况下电脑重启后服务没有自动运行问题
* 更新软件后需打开软件，选择左侧配置项后右键编辑，然后直接确定，再启动即可

[Unreleased]: https://github.com/koho/frpmgr/compare/v1.11.0...HEAD
[1.11.0]: https://github.com/koho/frpmgr/compare/v1.10.1...v1.11.0
[1.10.1]: https://github.com/koho/frpmgr/compare/v1.10.0...v1.10.1
[1.10.0]: https://github.com/koho/frpmgr/compare/v1.9.2...v1.10.0
[1.9.2]: https://github.com/koho/frpmgr/compare/v1.9.1...v1.9.2
[1.9.1]: https://github.com/koho/frpmgr/compare/v1.9.0...v1.9.1
[1.9.0]: https://github.com/koho/frpmgr/compare/v1.8.1...v1.9.0
[1.8.1]: https://github.com/koho/frpmgr/compare/v1.8.0...v1.8.1
[1.8.0]: https://github.com/koho/frpmgr/compare/v1.7.2...v1.8.0
[1.7.2]: https://github.com/koho/frpmgr/compare/v1.7.1...v1.7.2
[1.7.1]: https://github.com/koho/frpmgr/compare/v1.7.0...v1.7.1
[1.7.0]: https://github.com/koho/frpmgr/compare/v1.6.1...v1.7.0
[1.6.1]: https://github.com/koho/frpmgr/compare/v1.6.0...v1.6.1
[1.6.0]: https://github.com/koho/frpmgr/compare/v1.5.0...v1.6.0
[1.5.0]: https://github.com/koho/frpmgr/compare/v1.4.2...v1.5.0
[1.4.2]: https://github.com/koho/frpmgr/compare/v1.4.1...v1.4.2
[1.4.1]: https://github.com/koho/frpmgr/compare/v1.4.0...v1.4.1
[1.4.0]: https://github.com/koho/frpmgr/compare/v1.3.2...v1.4.0
[1.3.2]: https://github.com/koho/frpmgr/compare/v1.3.1...v1.3.2
[1.3.1]: https://github.com/koho/frpmgr/compare/v1.3.0...v1.3.1
[1.3.0]: https://github.com/koho/frpmgr/compare/v1.2.5...v1.3.0
[1.2.5]: https://github.com/koho/frpmgr/compare/v1.2.4...v1.2.5
[1.2.4]: https://github.com/koho/frpmgr/compare/v1.2.3...v1.2.4
[1.2.3]: https://github.com/koho/frpmgr/releases/tag/v1.2.3
