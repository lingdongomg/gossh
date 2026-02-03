# Design: GoSSH v1.2 增强

## Context

GoSSH 是一个基于 TUI 的 SSH 连接管理器。当前版本 (v1.1) 功能基本完善，但存在以下技术债务需要解决：

1. SSH 安全验证被禁用
2. 代码重复较多
3. 缺少测试覆盖
4. UI 国际化缺失

本设计文档描述如何在保持向后兼容的前提下解决这些问题。

## Goals / Non-Goals

### Goals
- 实现安全的 SSH HostKey 验证机制
- 提供用户友好的设置界面
- 建立可维护的 i18n 架构
- 消除代码重复，提高可维护性
- 建立单元测试基础

### Non-Goals
- 不支持 SSH Agent（可作为后续版本）
- 不支持 ProxyJump（可作为后续版本）
- 不实现会话录制功能

## Decisions

### 1. SSH HostKey 验证架构

**决策**: 采用分层验证策略

```
用户连接 → 检查 known_hosts → 匹配? → 允许连接
                              ↓
                         不匹配/未知?
                              ↓
                    显示指纹确认对话框
                              ↓
                    用户确认 → 保存到 known_hosts
                              ↓
                    用户拒绝 → 终止连接
```

**实现细节**:
- known_hosts 文件位置: `~/.config/gossh/known_hosts`
- 格式兼容 OpenSSH `~/.ssh/known_hosts`
- 使用 `golang.org/x/crypto/ssh/knownhosts` 包

**替代方案**:
- 仅警告不阻止: 安全性不足，不采用
- 强制验证无 UI: 用户体验差，不采用

### 2. 无密码模式加密改进

**决策**: 使用机器特征派生密钥

```go
// 组合多个机器特征生成唯一标识
machineID := hash(hostname + username + machineUUID)
key := Argon2id(machineID, salt)
```

**特征来源**:
- 主机名 (`os.Hostname()`)
- 用户名 (`os.Getenv("USER")`)
- 机器 UUID（Linux: `/etc/machine-id`, macOS: `IOPlatformUUID`, Windows: `MachineGuid`）

**替代方案**:
- 使用硬件 ID: 跨平台兼容性差
- 使用固定密钥: 当前方案，安全性不足

### 3. SSH 连接工厂模式

**决策**: 抽取公共连接函数到 `internal/ssh/factory.go`

```go
// ConnectOptions 统一连接配置
type ConnectOptions struct {
    Host           string
    Port           int
    User           string
    AuthMethod     ssh.AuthMethod
    Timeout        time.Duration
    HostKeyCallback ssh.HostKeyCallback
}

// Connect 统一连接入口
func Connect(opts ConnectOptions) (*ssh.Client, error)
```

**受影响模块**:
- `ssh/client.go` - 使用 Connect()
- `ssh/forward.go` - 使用 Connect()
- `sftp/sftp.go` - 使用 Connect()

### 4. i18n 国际化架构

**决策**: 采用简单的 map-based 翻译方案

```go
// internal/i18n/i18n.go
type Language string

const (
    LangEN Language = "en"
    LangZH Language = "zh"
)

var translations = map[Language]map[string]string{
    LangEN: {
        "welcome": "Welcome to GoSSH",
        "connect": "Connect",
        ...
    },
    LangZH: {
        "welcome": "欢迎使用 GoSSH",
        "connect": "连接",
        ...
    },
}

func T(key string) string // 获取当前语言翻译
func SetLanguage(lang Language) // 设置语言
```

**替代方案**:
- 使用第三方库 (go-i18n): 过度复杂
- 编译时生成: 不够灵活

### 5. 设置页面设计

**决策**: 新增 `SettingsModel` 视图

```
┌─────────────────────────────────────┐
│           ⚙️  设置                   │
├─────────────────────────────────────┤
│  语言 / Language                    │
│  [ ] 中文                           │
│  [●] English                        │
│                                     │
│  安全设置                            │
│  [ ] 启用主密码保护                   │
│  [ ] 修改主密码                       │
│                                     │
│  关于                                │
│  版本: v2.0.0                        │
│                                     │
│  [保存]  [取消]                       │
└─────────────────────────────────────┘
```

**快捷键**: `s` 从主列表进入设置

### 6. 连接健康检查

**决策**: 支持两种检查模式

1. **快速检查 (TCP)**: 仅验证端口可达
   ```go
   net.DialTimeout("tcp", host+":"+port, 5*time.Second)
   ```

2. **完整检查 (SSH)**: 完成 SSH 握手（不建立会话）
   ```go
   ssh.Dial("tcp", addr, config) // 成功后立即关闭
   ```

**UI 集成**:
- 列表视图显示状态图标: ✓ 可达 / ✗ 不可达 / ? 未知
- 快捷键 `t` 测试选中连接

### 7. SSH Config 导入

**决策**: 解析标准 OpenSSH 配置格式

支持的指令:
- `Host` - 连接别名
- `HostName` - 实际主机名
- `User` - 用户名
- `Port` - 端口（默认 22）
- `IdentityFile` - 密钥文件路径

```go
// internal/sshconfig/parser.go
func ParseSSHConfig(path string) ([]model.Connection, error)
```

**CLI 命令**:
```bash
gossh import --ssh-config [path]  # 默认 ~/.ssh/config
```

### 8. 版本号管理

**决策**: 使用构建时注入

```go
// main.go
var Version = "dev"

// 构建时注入
// go build -ldflags "-X main.Version=v2.0.0"
```

**Makefile 更新**:
```makefile
VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -X main.Version=$(VERSION)
```

## Risks / Trade-offs

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|----------|
| HostKey 验证中断现有工作流 | 中 | 中 | 提供"信任所有"临时选项 |
| 机器特征在容器中不稳定 | 低 | 低 | 检测容器环境并警告 |
| i18n 翻译不完整 | 中 | 低 | 缺失时 fallback 到英文 |
| SSH Config 格式不兼容 | 低 | 低 | 仅导入支持的字段，忽略其他 |

## Migration Plan

### Phase 1: 基础设施 (无破坏性变更)
1. 添加 i18n 模块
2. 添加 SSH 连接工厂
3. 添加测试框架

### Phase 2: 功能增强
1. 实现 HostKey 验证
2. 实现设置页面
3. 实现健康检查
4. 实现 SSH Config 导入

### Phase 3: 清理
1. 迁移所有 UI 文本到 i18n
2. 替换所有直接连接调用为工厂方法
3. 修复版本号

### Rollback
- 所有变更应可通过 git revert 回退
- 配置文件保持向后兼容

## Open Questions

1. **Q**: known_hosts 是否应该与 OpenSSH 共享？
   **A**: 建议独立存储，避免污染用户的 SSH 配置

2. **Q**: 设置是否应该同步到配置文件？
   **A**: 是，存储在 `config.yaml` 的 `settings` 部分

3. **Q**: 测试是否需要 mock SSH 服务器？
   **A**: 对于集成测试是，单元测试使用接口抽象
