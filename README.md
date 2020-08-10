# frpmgr

Windows 下的 [frp](https://github.com/fatedier/frp) 图形界面客户端。

## 编译
**安装依赖**:
- Visual Studio
- [WiX Toolset](https://wixtoolset.org/)

**环境配置**:

把`VsDevCmd.bat`, `cl.exe`添加到环境变量。通常目录为：
- `C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\Common7\Tools`
- `C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\VC\Tools\MSVC\14.26.28801\bin\Hostx64\x64`。

**执行编译**：

```shell script
git clone https://github.com/koho/frpmgr
cd frpmgr
build.bat
```

在 `bin` 目录找到生成的安装文件。