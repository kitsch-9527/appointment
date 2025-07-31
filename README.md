# appointment
用于预约北京地铁进站时段的命令行工具

## 使用说明

>以编译文件见 releases 文件夹

```
.\appointment.exe -h
用于预约地铁进站时段的命令行工具

Usage:
  appointment [flags]

Flags:
  -h, --help                     help for appointment
  -L, --line string              地铁线路名称 (default "昌平线")
  -p, --loop                     启用2分钟循环预约模式
  -d, --loop-duration duration   循环模式最大时长 (default 2m0s)
  -r, --max-retry int            最大重试次数(有限模式) (default 15)
  -s, --sleep duration           重试间隔时间 (default 1s)
  -n, --station string           车站名称 (default "沙河站")
  -l, --time-slot string         预约具体时段(如:0820-0830) (default "0820-0830")
  -o, --timeout duration         请求超时时间 (default 10s)
  -t, --token string             预约系统授权token (必填)

```

### web token获取

打开网址[https://webui.mybti.cn/#/login](https://webui.mybti.cn/#/login)进行登陆，打开控制台，选中network，点击请求，进行抓取token

#### 借鉴项目
[沙河地铁站预约脚本](https://github.com/congwa/shahe)
