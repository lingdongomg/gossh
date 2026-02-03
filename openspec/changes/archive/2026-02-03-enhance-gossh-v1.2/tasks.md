# Tasks: GoSSH v1.2 增强

## 1. 基础设施搭建

### 1.1 SSH 连接工厂
- [x] 1.1.1 创建 `internal/ssh/factory.go`，定义 `ConnectOptions` 和 `Connect()` 函数
- [x] 1.1.2 添加 HostKey 回调接口支持
- [x] 1.1.3 重构 `ssh/client.go` 使用工厂函数
- [x] 1.1.4 重构 `ssh/forward.go` 使用工厂函数
- [x] 1.1.5 重构 `sftp/sftp.go` 使用工厂函数

### 1.2 i18n 国际化模块
- [x] 1.2.1 创建 `internal/i18n/i18n.go`，实现语言管理和翻译函数
- [x] 1.2.2 创建 `internal/i18n/messages_en.go`，添加英文翻译
- [x] 1.2.3 创建 `internal/i18n/messages_zh.go`，添加中文翻译
- [x] 1.2.4 更新 `internal/model/model.go`，添加 `Language` 字段到 Settings

### 1.3 版本号管理
- [x] 1.3.1 更新 `main.go`，添加 Version 变量和构建时注入支持
- [x] 1.3.2 更新 `Makefile`，添加 LDFLAGS 版本注入
- [x] 1.3.3 更新 `internal/ui/views/help.go`，显示动态版本号

## 2. 安全增强

### 2.1 SSH HostKey 验证
- [x] 2.1.1 创建 `internal/ssh/hostkey.go`，实现 known_hosts 管理
- [x] 2.1.2 实现 `LoadKnownHosts()` 函数，加载已知主机
- [x] 2.1.3 实现 `SaveHostKey()` 函数，保存新主机密钥
- [x] 2.1.4 实现 `FormatFingerprint()` 函数，格式化指纹显示
- [x] 2.1.5 创建 `internal/ui/views/hostkey.go`，实现指纹确认对话框
- [x] 2.1.6 集成 HostKey 回调到 SSH 连接流程

### 2.2 加密安全改进
- [x] 2.2.1 创建 `internal/crypto/machine.go`，实现机器特征获取
- [x] 2.2.2 实现跨平台机器 UUID 获取 (Linux/macOS/Windows)
- [x] 2.2.3 更新 `internal/config/config.go`，使用机器特征派生密钥

## 3. 功能完善

### 3.1 StartupCommand 实现
- [x] 3.1.1 更新 `internal/ssh/terminal.go`，在会话建立后执行 StartupCommand
- [x] 3.1.2 更新 `internal/ui/views/form.go`，添加 StartupCommand 输入字段
- [x] 3.1.3 添加命令执行超时保护

### 3.2 SFTP 完善
- [x] 3.2.1 更新 `internal/sftp/sftp.go`，添加 `currentDir` 字段跟踪当前目录
- [x] 3.2.2 实现完整的 `cd` 命令支持
- [x] 3.2.3 添加文件传输进度回调接口
- [x] 3.2.4 更新 `internal/app/app.go`，显示传输进度

### 3.3 连接健康检查
- [x] 3.3.1 实现健康检查逻辑 (在 `internal/ssh/factory.go`)
- [x] 3.3.2 实现 `QuickCheck()` TCP 端口检查
- [x] 3.3.3 实现 `FullCheck()` SSH 握手检查
- [x] 3.3.4 更新 `internal/ui/views/list.go`，添加状态指示器显示
- [x] 3.3.5 添加快捷键 `t` 触发连接测试
- [x] 3.3.6 实现批量健康检查命令 (`gossh check`)

### 3.4 SSH Config 导入
- [x] 3.4.1 创建 `internal/sshconfig/parser.go`，实现 SSH 配置解析器
- [x] 3.4.2 支持 Host, HostName, User, Port, IdentityFile 指令
- [x] 3.4.3 更新 `internal/app/app.go`，添加 `import-ssh-config` 子命令
- [x] 3.4.4 实现重复检测和冲突处理

## 4. 用户体验

### 4.1 设置页面
- [x] 4.1.1 创建 `internal/ui/views/settings.go`，实现设置视图
- [x] 4.1.2 实现语言切换功能
- [x] 4.1.3 实现主密码启用/禁用功能
- [x] 4.1.4 实现主密码修改功能
- [x] 4.1.5 更新 `internal/ui/app.go`，添加设置视图状态
- [x] 4.1.6 添加快捷键 `s` 进入设置页面

### 4.2 UI 国际化
- [x] 4.2.1 替换 `internal/ui/views/list.go` 中的硬编码文本
- [x] 4.2.2 替换 `internal/ui/views/form.go` 中的硬编码文本
- [x] 4.2.3 替换 `internal/ui/views/setup.go` 中的硬编码文本
- [x] 4.2.4 替换 `internal/ui/views/unlock.go` 中的硬编码文本
- [x] 4.2.5 替换 `internal/ui/views/help.go` 中的硬编码文本
- [x] 4.2.6 替换 `internal/ui/views/confirm.go` 中的硬编码文本
- [x] 4.2.7 替换 `internal/app/app.go` 中的 CLI 输出文本

## 5. 测试

### 5.1 单元测试
- [x] 5.1.1 创建 `internal/crypto/encrypt_test.go`，测试加密解密
- [x] 5.1.2 创建 `internal/crypto/password_test.go`，测试密码哈希验证
- [x] 5.1.3 创建 `internal/crypto/machine_test.go`，测试机器特征获取
- [x] 5.1.4 创建 `internal/config/config_test.go`，测试配置读写
- [x] 5.1.5 创建 `internal/model/model_test.go`，测试模型验证
- [x] 5.1.6 创建 `internal/sshconfig/parser_test.go`，测试 SSH 配置解析
- [x] 5.1.7 创建 `internal/ssh/hostkey_test.go`，测试 HostKey 管理
- [x] 5.1.8 创建 `internal/i18n/i18n_test.go`，测试国际化

### 5.2 集成测试
- [x] 5.2.1 更新 `Makefile`，添加 test 目标
- [x] 5.2.2 配置测试覆盖率报告 (`make coverage`)

## 6. 文档更新

- [x] 6.1 更新 `README.md`，添加新功能说明
- [x] 6.2 更新 `README_CN.md`，添加新功能说明
- [x] 6.3 更新 `openspec/project.md`，更新版本号和功能列表

## Dependencies

任务依赖关系：
- 1.1 (SSH 工厂) → 2.1 (HostKey) → 3.1-3.4 (功能完善)
- 1.2 (i18n) → 4.2 (UI 国际化)
- 1.3 (版本号) 可并行
- 2.2 (加密) 可并行
- 5.x (测试) 依赖对应功能模块完成

## Parallelizable Work

以下任务可并行进行：
- 1.1, 1.2, 1.3 可同时进行
- 2.1, 2.2 可同时进行
- 3.1, 3.2, 3.3, 3.4 可同时进行（在 1.1 完成后）
- 4.1, 4.2 可同时进行（在 1.2 完成后）
- 5.x 测试在对应功能完成后可立即开始

## 完成状态

**已完成的核心功能：**
- ✅ SSH 连接工厂
- ✅ i18n 国际化模块
- ✅ 版本号管理
- ✅ SSH HostKey 验证
- ✅ 机器特征加密改进
- ✅ StartupCommand 实现
- ✅ SFTP cd 命令和进度显示
- ✅ 连接健康检查
- ✅ SSH Config 导入
- ✅ 设置页面核心功能
- ✅ 单元测试
- ✅ UI 国际化文本替换
- ✅ 设置页面集成到主 UI 流程
- ✅ 列表视图的状态指示器
- ✅ HostKey 确认对话框
- ✅ README 文档更新
