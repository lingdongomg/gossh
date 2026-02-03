# Change: GoSSH v1.2 增强 - 安全、功能与用户体验全面升级

## Why

GoSSH v1.1 已经是一个功能完善的 SSH 连接管理器，但在安全性、功能完整性和用户体验方面仍有提升空间：

1. **安全隐患**：当前使用 `ssh.InsecureIgnoreHostKey()` 跳过主机密钥验证，存在中间人攻击风险；无密码模式使用固定密钥加密，安全性不足
2. **功能缺失**：`StartupCommand` 字段未实现、SFTP cd 命令不完整、缺少连接健康检查、无法导入现有 SSH 配置
3. **代码质量**：SSH 连接逻辑重复、缺少单元测试、版本号不一致
4. **用户体验**：UI 中英文混用、缺少设置页面

## What Changes

### 安全增强
- **[P0] SSH HostKey 验证** - 实现 known_hosts 文件管理，首次连接提示确认指纹
- **[P0] 加密安全改进** - 无密码模式使用机器特征派生密钥替代固定字符串

### 功能完善
- **[P1] StartupCommand 实现** - 连接成功后自动执行配置的启动命令
- **[P1] SFTP 完善** - 实现 cd 命令的工作目录跟踪，添加传输进度显示
- **[P1] 连接健康检查** - 添加 ping/test 命令，列表显示连接状态指示器
- **[P1] SSH Config 导入** - 支持从 `~/.ssh/config` 导入现有配置

### 代码质量
- **[P2] 消除重复代码** - 抽取公共的 SSH 连接工厂函数
- **[P2] 添加单元测试** - 覆盖 crypto、config、ssh 核心模块
- **[P2] 修复版本号** - 统一版本号管理

### 用户体验
- **[P2] 设置页面** - 新增设置视图，支持语言切换（中/英）和主密码管理
- **[P2] UI 语言统一** - 实现 i18n 国际化支持

## Impact

### 受影响的模块
| 模块 | 变更类型 | 说明 |
|------|----------|------|
| `internal/ssh/client.go` | 修改 | 添加 HostKey 验证、抽取公共函数 |
| `internal/ssh/forward.go` | 修改 | 使用公共连接函数 |
| `internal/ssh/terminal.go` | 修改 | 实现 StartupCommand |
| `internal/sftp/sftp.go` | 修改 | 使用公共连接函数、完善 cd |
| `internal/config/config.go` | 修改 | 添加 known_hosts、语言设置管理 |
| `internal/crypto/encrypt.go` | 修改 | 改进无密码模式加密 |
| `internal/model/model.go` | 修改 | 添加语言、HostKey 相关字段 |
| `internal/ui/views/` | 新增/修改 | 新增设置页面、实现 i18n |
| `internal/app/app.go` | 修改 | 添加健康检查、SSH 配置导入命令 |

### 受影响的文件数
- 新增文件: ~8 个（测试文件、i18n 模块、设置视图）
- 修改文件: ~15 个

### 兼容性
- **配置文件**: 向后兼容，新字段使用默认值
- **known_hosts**: 首次运行时自动创建

## Dependencies

- 无外部依赖变更
- 需要 Go 1.24+ 支持

## Risks

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| HostKey 验证可能影响现有连接 | 中 | 提供"信任此主机"的 UI 交互 |
| 加密方式变更导致旧配置不兼容 | 低 | 检测旧格式并提示重新加密 |
| i18n 可能遗漏翻译 | 低 | 使用 fallback 机制 |

## Success Metrics

- [ ] 所有 SSH 连接通过 HostKey 验证
- [ ] 单元测试覆盖率 > 60%
- [ ] 支持中英文切换
- [ ] SSH 配置导入成功率 > 95%
