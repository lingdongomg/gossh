# GoSSH

基于 Go 和 [Bubbletea](https://github.com/charmbracelet/bubbletea) 构建的 TUI（终端用户界面）SSH 连接管理器。

[English Documentation](README.md)

## 功能特性

### 核心功能
- **连接管理** - 通过 TUI 添加、编辑、删除 SSH 连接
- **安全存储** - 主密码保护，采用 AES-256-GCM 加密
- **分组管理** - 将连接组织到不同分组中
- **搜索** - 实时搜索和过滤连接
- **导入/导出** - 基于 YAML 的备份和恢复

### 高级功能 (v1.1)
- **SFTP 文件传输** - 交互式 SFTP Shell，支持文件操作
- **端口转发** - 本地转发 (-L) 和远程转发 (-R)
- **批量执行** - 在多台服务器上同时执行命令

### 新功能 (v1.2)
- **SSH 主机密钥验证** - 安全的主机密钥管理，支持 known_hosts
- **增强安全性** - 无密码模式使用机器特征派生密钥
- **启动命令** - SSH 连接后自动执行命令
- **连接健康检查** - 使用 `t` 键或 `gossh check` 命令测试连接
- **SSH Config 导入** - 从 `~/.ssh/config` 导入连接
- **设置页面** - 语言切换（English/中文）和密码管理
- **国际化** - 完整的中英文 i18n 支持
- **SFTP 改进** - `cd` 命令支持工作目录跟踪和进度显示

## 安装

### 从源码编译

```bash
git clone https://github.com/lingdongomg/gossh.git
cd gossh
make build
```

### 使用 Go Install

```bash
go install github.com/lingdongomg/gossh@latest
```

## 快速开始

```bash
# 启动 TUI 应用
./gossh

# 显示帮助
./gossh help

# 列出所有连接
./gossh list

# 通过名称连接服务器
./gossh connect myserver
```

## 使用方法

### TUI 模式

启动交互式 TUI：

```bash
./gossh
```

#### 快捷键

| 按键 | 操作 |
|-----|------|
| `↑/k` | 向上移动 |
| `↓/j` | 向下移动 |
| `g/G` | 跳转到顶部/底部 |
| `/` | 搜索连接 |
| `Enter` | 连接到选中的服务器 |
| `a` | 添加新连接 |
| `e` | 编辑选中的连接 |
| `d` | 删除选中的连接 |
| `t` | 测试连接 (v1.2) |
| `s` | 设置 (v1.2) |
| `?` | 显示帮助 |
| `q` | 退出 |

### 命令行模式

#### 基本命令

```bash
# 显示版本
gossh version

# 列出所有连接
gossh list

# 通过名称连接
gossh connect <name>

# 导出连接到文件
gossh export [filename]

# 从文件导入连接
gossh import <filename>

# 从 SSH Config 导入 (v1.2)
gossh import --ssh-config [路径]
```

#### 连接健康检查 (v1.2)

```bash
# 检查所有连接
gossh check --all

# 检查指定连接
gossh check --name=myserver

# 按分组检查
gossh check --group=Production
```

#### SFTP 会话

```bash
gossh sftp <连接名称>
```

SFTP Shell 命令：
- `ls [路径]` - 列出目录内容
- `cd <路径>` - 切换目录 (v1.2: 支持工作目录跟踪)
- `pwd` - 显示当前工作目录
- `get <远程> [本地]` - 下载文件 (v1.2: 带进度显示)
- `put <本地> [远程]` - 上传文件 (v1.2: 带进度显示)
- `mkdir <路径>` - 创建目录
- `rm <路径>` - 删除文件
- `rmdir <路径>` - 递归删除目录
- `exit/quit` - 退出 SFTP 会话

#### 端口转发

```bash
# 本地端口转发 (-L)
# 将本地 3306 端口转发到远程的 localhost:3306
gossh forward <name> -L 3306:localhost:3306

# 远程端口转发 (-R)
# 将远程 8080 端口转发到本地的 localhost:80
gossh forward <name> -R 8080:localhost:80
```

#### 批量执行

在多台服务器上执行命令：

```bash
# 在指定分组的所有服务器上执行
gossh exec "uptime" --group=Production

# 在具有特定标签的服务器上执行
gossh exec "df -h" --tags=web,nginx

# 在指定名称的服务器上执行
gossh exec "hostname" --names=server1,server2

# 设置自定义超时时间（默认：30秒）
gossh exec "long-running-command" --group=All --timeout=120
```

## 配置

配置以 YAML 格式存储：

- **Linux/macOS**: `~/.config/gossh/config.yaml`
- **Windows**: `%APPDATA%\gossh\config.yaml`

### 连接字段

| 字段 | 描述 |
|------|------|
| `name` | 连接名称（唯一标识） |
| `host` | 服务器主机名或 IP |
| `port` | SSH 端口（默认：22） |
| `user` | 用户名 |
| `password` | 密码（加密存储） |
| `key_path` | SSH 私钥路径 |
| `key_passphrase` | 私钥密码（加密存储） |
| `group` | 用于组织的分组名称 |
| `tags` | 用于过滤的标签列表 |
| `startup_command` | 连接后执行的命令 |

## 安全性

- **主密码**：首次运行时设置，使用 Argon2id 密钥派生
- **机器特征加密** (v1.2)：无密码模式使用机器派生密钥（主机名 + 用户名 + 机器 UUID）
- **加密**：使用 AES-256-GCM 存储敏感数据（密码、密钥密码）
- **主机密钥验证** (v1.2)：支持 known_hosts 管理和指纹确认

## 依赖

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI 框架
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI 组件
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - 样式库
- [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) - SSH 和加密
- [pkg/sftp](https://github.com/pkg/sftp) - SFTP 支持
- [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml) - 配置文件

## 许可证

MIT License
