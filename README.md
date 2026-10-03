# sshpass

纯 Go、无 CGO 的跨平台口令注入工具，用法接近经典 `sshpass`：在伪终端中启动命令，自动应答 SSH 主机密钥确认，再写入登录口令。

默认会信任**第一次见到的主机指纹**（自动回复 `yes`）。只建议在内网、实验室或可接受该风险的环境使用。公网请预先写入 `known_hosts`，并加上 `--no-host-confirm`。

## 依赖

- Go 1.22+
- 本机已安装要包装的命令（如 `ssh` / `scp` / `rsync` / `sudo`）
- Windows 10 及以上（ConPTY）

## 安装

### 从 Release 下载

到 [Releases](https://github.com/aiqoder/sshpass/releases) 下载对应平台的压缩包，解压后放入 `PATH`。

### Go install

```bash
go install github.com/aiqoder/sshpass/cmd/sshpass@latest
```

### 源码编译

```bash
CGO_ENABLED=0 go build -o sshpass ./cmd/sshpass
```

多平台产物：

```bash
make build
```

输出位于 `build/sshpass-<os>-<arch>`（Windows 带 `.exe`）。

推送形如 `v*` 的 tag（例如 `v0.1.0`）会触发 GitHub Actions，自动构建多平台二进制并创建 Release。

## 用法

```text
sshpass [-p pass|-e|-f file|-d fd] [-P prompt] [-H host-prompt] [-y]
        [--no-host-confirm] [-t sec] [-v] -- command [args...]
```

| 选项 | 说明 |
|------|------|
| `-p` | 命令行口令（会出现在进程列表，不推荐） |
| `-e` | 从环境变量 `SSHPASS` 读取口令 |
| `-f` | 从文件第一行读取口令 |
| `-d` | 从文件描述符读取口令（Unix） |
| `-P` | 口令提示匹配串，默认 `assword` |
| `-H` | 主机确认匹配串，默认含 `(yes/no` |
| `-y` | 显式打开主机密钥自动应答（默认已打开） |
| `--no-host-confirm` | 关闭自动 `yes` |
| `-t` | 等待口令提示的秒数，默认 30 |
| `-v` | 诊断信息写到 stderr，不会打印口令 |

口令来源 `-p` / `-e` / `-f` / `-d` 四选一。

## 示例

```bash
export SSHPASS='your-password'
./sshpass -e -- ssh user@host hostname

./sshpass -f ./pass.txt -- scp a.txt user@host:
./sshpass -p 'secret' -v -- sudo -S id
./sshpass -e --no-host-confirm -- ssh user@host uname
```

## 安全说明

- `-p` 会把口令暴露给本机其他用户（`ps`）。优先用 `-e` 或 `-f`（文件权限 `600`）。
- 默认自动应答主机密钥，存在中间人风险。
- 本工具不实现 SSH 协议，只向已有命令的终端注入文本。
