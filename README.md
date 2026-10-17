# kit

一个用 Go 编写的轻量级 Linux 命令行工具集

## 命令行用法

```
kit <tool> [args]
```

- `tool`：工具名，目前支持 `file`、`port`
- 无参数或 `help` / `-h` / `--help`：打印总帮助
- 每个工具和子命令都支持 `-h` / `--help` 查看自己的帮助

## 工具用法

### file —— 文件占用查询

```
kit file <action> [args]
```

#### file scan —— 查看哪些进程正在使用某个文件

```
kit file scan [--json] <path>
```

| 参数 | 说明 |
|------|------|
| `<path>` | 文件路径，必填 |
| `--json` | 以 JSON 输出结果 |

默认以表格输出：`PID / USER / NAME / FD / ACCESS / TARGET / COMMAND`；
没有进程占用时输出 `no process is using <path>`。

示例：

```bash
kit file scan /var/log/syslog           # 表格列出使用该文件的进程
kit file scan --json /var/log/syslog    # 以 JSON 输出
```

### port —— 端口反查与清理

```
kit port <action> [args]
```

#### port kill —— 杀掉占用某端口的进程

```
kit port kill [--force] [--json] <port>
```

| 参数 | 说明 |
|------|------|
| `<port>` | 端口号（1–65535），必填 |
| `--force` | 跳过确认直接 kill |
| `--json` | 以 JSON 输出被杀进程信息 |

默认会先打印占用的进程，再询问是否确认；`--force` 时直接执行。

示例：

```bash
kit port kill 8080          # 列出 8080 进程，确认后杀掉
kit port kill --force 8080  # 不确认直接杀
```

#### port list —— 列出范围内占用端口的进程

```
kit port list [--json] <start-end | start end>
```

| 参数 | 说明 |
|------|------|
| `<start-end>` | 端口区间写法，如 `8000-9000` |
| `<start> <end>` | 两个参数分开写，如 `8000 9000` |
| `--json` | 以 JSON 输出结果 |

默认以表格输出：`PORT / PROTOCOL / PID / USER / NAME / COMMAND / ADDR / STATE`。

示例：

```bash
kit port list 8000-9000     # 表格列出 8000–9000 的进程
kit port list --json 8000 9000
```

## 编译

要求 Go 1.26+，无第三方依赖：

```bash
go build -o kit ./cmd/kit
```

其他常用命令：

```bash
go run ./cmd/kit port list 8000-9000  # 免编译直接运行
go test ./...                                  # 运行测试
go install ./cmd/kit                   # 安装到 GOPATH/bin
```
