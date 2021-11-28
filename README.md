# daemonupgrader
一个简单的通用服务守护与自动升级服务

编译

```
go build -ldflags "-w -X 'github.com/vrherog/daemonupgrader/version.BuildVersion=${VERSION}'"
```

使用方法：

配置 config.yaml，此配置文件可改名为编译后的可执行文件同名，扩展名支持 .yaml、.conf。

```
name: 安装本程序为系统服务的名称，不可包含空格，可选
displayName: 用于系统服务管理器的显示名称，可包含空格，可选
description: 用于系统服务管理器的注释内容，可选
username: 运行此服务的用户，可选，默认 linux 为 root、windows 为 LocalSystem
Option: System specific options.

# 监视服务列表，在此列表中的服务停止运行时将被再启动
services:
  - name: 系统服务名称，必填
    interval: 检测周期，可选，默认为 3s，格式为 59s 59m 99h 59m59s

# 升级包列表
packages:
  - name: 程序名
    interval: 检测周期，可选，默认为 30s
    uriCheckVersion: 检查新版本 URI，应返回纯文本版本号：如 1.1.2
    uriDownloadPackage: 新版本程序包 URI，应为可直接复制的压缩包，支持 ZIP、GZIP，不支持可执行安装程序
    workDirectory: 程序包复制目标路径，即程序安装目录
    commandGetVersion: 获取本地程序版本号的命令行命令，应返回纯文本版本号：如 1.1.2，程序路径中包含空格
                       的必须添加"号
    needShutdown: true 升级时需要关闭程序，此处为避免强制退出可能导致的问题以及 UI 程序可提醒用户处理升级，
                  只将升级包信息写入 upgrade.ready 文件由该程序自行处理，或者程序将 name 写入
                  upgrade.ok 中并安全退出，由守护服务处理升级。
                  false 无需关闭程序的升级包，守护服务下载升级包后即覆盖升级。
```
