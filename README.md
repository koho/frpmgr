# frpmgr

Windows 下的 [frp](https://github.com/fatedier/frp) 图形界面客户端。

![frpmgr](/docs/frpmgr.png)

## 编译
**安装依赖**:
- Visual Studio
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