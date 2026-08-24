# 变更日志

本文件记录项目的所有重要变更。

格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## v1.5.0 更新概述

### 新增

- 生产部署文档 `docs/deployment.md`；扩展 `docs/security.md` 生产清单
- 生产 TLS 配置：`tls.cert_file` / `tls.key_file` / `tls.ca_file`（`pkg/config` + `WithTLSCertFiles` 等）
- `mcp-moqt server -config` 从 YAML/JSON 启动；`configs/config.production.yaml` 样例
- `doctor` 增强：QUIC 探测、metrics `/healthz`、TLS 证书校验、UDP 端口检查
- MCP 工具：`validate_config`、`export_metrics`
- MCP 资源：`moqt://docs/deployment`、`moqt://docs/security`
- Prometheus 指标扩充：数据轨消息/字节、吞吐率、`tracks_total`
- Grafana 样例仪表盘 `configs/grafana/dashboard.json`
- 示例 `examples/multi_client`（多客户端并发）；压测 benchmark 扩充

### 变更

- README 明确 stdio 与 QUIC/MOQT 定位
- `MOQTTransport` 标注为 legacy，推荐使用 `NewMoqTransport`
- 文档语言规范：**README 中文**、**`docs/` 全英文**、**CHANGELOG 中文**

## v1.4.0 更新概述

### 新增

- 服务端多连接：`Listen` + `Accept` 循环；`Serve` / `PreconnectedTransport` 支持并发客户端
- Prometheus 文本格式 `/metrics` HTTP 服务（`WithMetricsAddr` / `MetricsHTTPServer`）
- 配置文件热加载（`config.WatchFile` 轮询）

### 变更

- 路线图标记多连接、`/metrics`、配置热加载为已完成
- 统一 `gofmt` 代码风格
- Go 工具链最低版本提升至 **1.25.13**（CI / Docker / `go.mod`）

## v1.3.0 更新概述

### 新增

- MCP 服务能力（满足平台上架要求）：
  - 工具：`health_check`、`get_transport_info`、`list_data_tracks`、`estimate_rtt`
  - 提示词：`debug_moqt_session`、`configure_transport`
  - 资源：`moqt://docs/overview`、`moqt://config/example`、`moqt://tracks/catalog`、`moqt://install/quickstart`
- 友好安装方式：`scripts/install.sh`、`scripts/install.ps1`、`Makefile`（支持本地 checkout 安装）
- CLI 工具：`cmd/mcp-moqt`（`server` / `client` / `doctor` / `version`）
  - **stdio 模式**：`mcp-moqt server -stdio`
  - doctor 增加工具/提示词/资源冒烟检查
- MCP Host 配置样例：`configs/mcp/*.json`
- 文档：`docs/getting-started.md`、`docs/capabilities.md`、`docs/security.md`
- 控制连接可靠性接线：ACK 跟踪、心跳 keepalive 通知、Write 重试计数、连接级 Metrics
- 客户端连接池：`WithQUICConfig` 覆盖池参数；会话结束 `Put` 归还连接
- 服务端 composite subscribe（控制轨 + 数据轨）；`SubscribeDataTrack` / `AttachSessionSubscribe`
- 数据轨道 e2e 集成测试；CI Go 1.25、coverage 软门槛、`govulncheck`、Release 工作流

### 变更

- 示例 server 支持 `-stdio`
- 示例 client 演示 list/call/read 能力
- Dockerfiles 更新到 Go 1.25

## v1.2.0 更新概述

### 变更

- 项目结构优化：调整目录结构以符合成熟 GitHub 项目规范
- 文档增强：补充项目简介、主要特性与贡献指南
- 测试覆盖改进：确保单元测试与集成测试全部通过
- 代码质量优化：通过 `go vet` 检查
- 更新所有示例文件与文档中的版本号

## v1.1.1 更新概述

### 变更

- 依赖升级：
  - github.com/modelcontextprotocol/go-sdk：v1.4.0 → v1.5.0
  - github.com/quic-go/quic-go：v0.53.0 → v0.59.0
- 提升协议兼容性与稳定性

## v1.1.0 更新概述

### 新增

- 数据轨道高级特性：流式与批处理支持
- 完善的错误处理机制：细粒度错误类型与错误码
- 增强监控指标：Prometheus 支持
- 配置热加载：运行时更新无需重启服务
- 多环境配置支持：dev/test/prod 默认模板

### 变更

- 优化 QUIC 连接池性能与连接复用策略
- 增强流生命周期管理效率
- 改进心跳检测机制：调整间隔与超时策略
- 优化消息确认机制：改进 ACK/NACK 处理
- 增强重试机制：改进指数退避算法

### 修复

- 数据轨道订阅内存泄漏
- 高并发下 QUIC 连接池死锁
- 网络不稳定时心跳检测误报
- 多线程环境下监控指标竞态
- 配置文件解析类型转换错误

## v1.0.0 更新概述

### 新增

- QUIC 连接池：连接复用、空闲超时、自动清理
- QUIC 流管理：生命周期跟踪与统计
- QUIC 配置优化：流式构建器模式
- 数据轨道支持：resources、tools、notifications
- 可靠性特性：消息确认、心跳检测、重试机制、监控指标
- 服务端多连接支持
- 完善的错误处理与详细信息
- 完整测试套件：单元测试、集成测试、性能基准
- 详细文档：API 文档与设计文档

### 修复

- 心跳检测 goroutine 泄漏
- `crypto/rand` 失败时会话 ID 生成回退问题
- `DefaultMetrics()` 并发安全问题
- `track_base.go` 中 `context.Cause()` 使用问题
- discovery 机制中 `InitialMaxRequestID` 默认值问题
- 集成测试超时问题

## v0.7.1 更新概述

### 修复

- 心跳检测 goroutine 泄漏：超时时正确等待 goroutine 退出
- 会话 ID 生成回退：使用时间戳+计数器保证唯一性
- `DefaultMetrics()` 并发安全：读写锁保护
- `track_base.go` 兼容性：使用 `ctx.Err()` 替代 `context.Cause()`

### 变更

- discovery 响应版本号从 0.4.0 更新为 0.7.0

## v0.7.0 更新概述

### 新增

- QUIC 连接池：连接复用、空闲超时、自动清理
- QUIC 流管理器：统一流生命周期管理
- QUIC 配置构建器：流式接口
- QUIC 连接池、流管理器、配置构建器单元测试

## v0.6.0 更新概述

### 变更

- 导出 `DataTrackHandler.SessionID()` 供外部访问会话 ID
- 导出 `Retry.CalculateDelay()` 与 `Retry.IsRetryable()` 供外部调用
- 修复 context 取消时心跳检测未正确停止
- 修复测试中错误率与吞吐量计算问题

### 安全

- TLS 证书有效期从 24 小时延长至 365 天
- 限制 TLS 最低版本为 TLS 1.2
- 修复心跳 goroutine 泄漏风险
- 修复重试机制中非线程安全的随机数生成器
- 增加消息大小限制（10MB）防止内存攻击

### 新增

- 消息确认、心跳检测、重试机制、监控指标、数据轨道处理器的单元测试

## v0.5.0 更新概述

### 新增

- 可靠性特性：
  - 消息确认机制（ACK/NACK）与超时检测
  - 心跳检测保持连接活跃并发现断连
  - 指数退避重试机制
  - 监控指标：消息统计、吞吐量、错误率等
- 特性演示示例代码

## v0.4.0 更新概述

### 新增

- 数据轨道支持：resources、tools、notifications
- `data_tracks.go`：数据轨道类型与命名空间定义
- `data_track_handler.go`：订阅与发布机制实现
- 扩展 discovery 响应包含数据轨道信息
- 数据轨道示例代码
- 数据轨道相关错误类型
- 更新文档说明数据轨道用法

## v0.3.0 更新概述

### 变更

- 统一入口：`NewMoqTransport(RoleServer/RoleClient)`
- TLS 默认行为：服务端自签名证书，客户端 `InsecureSkipVerify`
- ALPN 默认为 `moq-00`（基于 moqtransport draft-11）
- 合并 `internal/transport/` 重复代码至 `pkg/moqttransport/`
- 创建 `docs/` 目录与 API、设计文档
- 统一测试目录结构至 `test/`

### 新增

- 增强错误处理
- 详细代码注释
- 优化依赖管理
- 提高测试覆盖率
- 日志功能与多级别支持
- 配置文件支持（YAML/JSON）
- 环境变量支持
- 完善的错误类型与处理机制
- 性能基准测试支持
- CI/CD 自动化集成
- 示例配置文件与环境变量模板

## v0.2.0 更新概述

### 变更

- 统一入口：`NewMoqTransport(RoleServer/RoleClient)`
- TLS 默认行为：服务端自签名证书，客户端 `InsecureSkipVerify`
- ALPN 默认为 `moq-00`（基于 moqtransport draft-11）
- 合并 `internal/transport/` 重复代码至 `pkg/moqttransport/`
- 创建 `docs/` 目录与 API、设计文档
- 统一测试目录结构至 `test/`

### 新增

- 增强错误处理
- 详细代码注释
- 优化依赖管理
- 提高测试覆盖率
- 日志功能与多级别支持
- 配置文件支持（YAML/JSON）
- 环境变量支持
- 完善的错误类型与处理机制
- 性能基准测试支持
- CI/CD 自动化集成
- 示例配置文件与环境变量模板
