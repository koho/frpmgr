# frpmgr

Windows 下的 [frp](https://github.com/fatedier/frp) 图形界面客户端。

![frpmgr](/docs/frpmgr.jpg)

系统需求：win7及以上版本

## 特征
* 简易的编辑界面
* 支持导入/导出配置文件
* 开机自启动
* 多配置文件管理

## 编译
**安装依赖**:
- go >=1.16
- Visual Studio
- [MinGW](https://www.mingw-w64.org/)
- [WiX Toolset](https://wixtoolset.org/)

**环境配置**:

把`VsMSBuildCmd.bat`添加到环境变量。通常目录为：
- `C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\Common7\Tools`

**执行编译**:

```shell script
git clone https://github.com/koho/frpmgr
cd frpmgr
build.bat
```

在 `bin` 目录找到生成的安装文件。

**调试**:

编译程序需要渲染图标生成资源和对某些文件打补丁，这些操作都在`build.bat`完成。因此在第一次调试前需要运行一次`build.bat`，后续调试不需要再执行该脚本。

## 捐助

如果您觉得本项目对你有帮助，欢迎给予我们支持。

### 微信支付捐赠

![donate-wechat](/docs/donate-wechat.jpg)
