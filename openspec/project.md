# Project Context

## Purpose
GoSSH 是一个基于终端用户界面（TUI）的 SSH 连接管理器，旨在提供安全、高效的服务器连接管理体验。主要目标：
- 通过直观的 TUI 界面管理 SSH 连接
- 安全存储敏感凭证（使用 AES-256-GCM 加密）
- 支持高级功能如 SFTP 文件传输、端口转发和批量命令执行

## Tech Stack
- **语言**: Go 1.24+
- **TUI 框架**: [Bubbletea](https://github.com/charmbracelet/bubbletea) - Elm 架构的 Go TUI 框架
- **UI 组件**: [Bubbles](https://github.com/charmbracelet/bubbles) - Bubbletea 组件库
- **样式**: [Lipgloss](https://github.com/charmbracelet/lipgloss) - 终端样式库
- **SSH**: golang.org/x/crypto/ssh - Go 官方 SSH 库
- **SFTP**: [pkg/sftp](https://github.com/pkg/sftp) - SFTP 客户端库
- **配置**: gopkg.in/yaml.v3 - YAML 解析
- **UUID**: github.com/google/uuid - 唯一标识符生成

## Project Conventions

### Code Style
- 遵循 Go 官方代码风格规范
- 使用 `gofmt` 格式化代码
- 使用 `golangci-lint` 进行代码检查
- 包名使用小写单词，无下划线
- 导出函数/类型使用 PascalCase，私有使用 camelCase
- 错误处理使用 `if err != nil` 模式，不使用 panic

### Architecture Patterns
- **分层架构**: 采用 internal 目录组织内部包
  - `internal/app` - 应用入口和 CLI 命令处理
  - `internal/config` - 配置管理和持久化
  - `internal/crypto` - 加密服务（AES-256-GCM, Argon2id）
  - `internal/model` - 数据模型定义
  - `internal/sftp` - SFTP 客户端实现
  - `internal/ssh` - SSH 连接、端口转发、批量执行
  - `internal/ui` - Bubbletea TUI 视图和组件
- **Elm 架构**: UI 遵循 Bubbletea 的 Model-Update-View 模式
- **单一职责**: 每个包专注于单一功能领域

### Testing Strategy
- 使用 `go test` 运行测试
- 测试命令: `make test`
- 重点测试加密模块和配置管理

### Git Workflow
- 主分支: `master`
- 使用语义化提交信息
- 构建产物 (`bin/`) 不提交到仓库

## Domain Context
- **SSH 连接**: 支持密码认证和密钥认证两种方式
- **加密安全**: 
  - 主密码使用 Argon2id 哈希存储
  - 敏感数据（密码、密钥密码）使用 AES-256-GCM 加密
- **连接状态**: 跟踪每个连接的最后连接时间和状态（成功/失败）
- **分组管理**: 支持将连接组织到不同分组（Production, Development, Testing）

## Important Constraints
- 配置文件权限必须为 0600（仅用户可读写）
- 主密码设置后不可更改（需重新初始化）
- 解锁失败超过限定次数后自动退出
- 跨平台支持（Linux, macOS, Windows）

## External Dependencies
- **配置存储位置**:
  - Linux/macOS: `~/.config/gossh/config.yaml`
  - Windows: `%APPDATA%\gossh\config.yaml`
- **SSH 密钥**: 支持读取本地 SSH 私钥文件
- **系统终端**: 需要支持 ANSI 转义序列的终端

## Build & Run
```bash
# 构建
make build          # 输出到 bin/gossh

# 运行
make run            # 直接运行
./bin/gossh         # 运行编译后的二进制

# 跨平台构建
make build-all      # 构建 Linux, macOS, Windows 版本

# 其他命令
make deps           # 下载依赖
make fmt            # 格式化代码
make lint           # 代码检查
make clean          # 清理构建产物
```

## Current Version
- **版本号**: v1.2.0
- **核心功能**: 连接管理、安全存储、分组、搜索、导入导出
- **高级功能 (v1.1)**: SFTP 文件传输、端口转发、批量执行
- **新增功能 (v1.2)**: 
  - SSH HostKey 验证（安全增强）
  - 机器特征加密（无密码模式安全改进）
  - StartupCommand 执行
  - SFTP cd 命令和进度显示
  - 连接健康检查
  - SSH Config 导入
  - 设置页面（语言切换、密码管理）
  - i18n 国际化支持（中/英）
  - 单元测试覆盖
