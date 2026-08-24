# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v1.4.0 更新概述

### Added

- 服务端多连接：`Listen` + `Accept` 循环；`Serve` / `PreconnectedTransport` 支持并发客户端
- Prometheus 文本格式 `/metrics` HTTP 服务（`WithMetricsAddr` / `MetricsHTTPServer`）
- 配置文件热加载（`config.WatchFile` 轮询）

### Changed

- Roadmap 标记多连接、`/metrics`、配置热加载为已完成
- 统一 `gofmt` 代码风格
- Go 工具链最低版本提升至 **1.25.13**（CI / Docker / `go.mod`）

## v1.3.0 更新概述

### Added

- MCP service capabilities for marketplace completeness:
  - Tools: `health_check`, `get_transport_info`, `list_data_tracks`, `estimate_rtt`
  - Prompts: `debug_moqt_session`, `configure_transport`
  - Resources: `moqt://docs/overview`, `moqt://config/example`, `moqt://tracks/catalog`, `moqt://install/quickstart`
- Friendly install methods: `scripts/install.sh`, `scripts/install.ps1`, `Makefile`（支持本地 checkout 安装）
- CLI utility: `cmd/mcp-moqt`（`server` / `client` / `doctor` / `version`）
  - **stdio 模式**：`mcp-moqt server -stdio`
  - doctor 增加 tools/prompts/resources 冒烟检查
- MCP host 配置样例：`configs/mcp/*.json`
- 文档：`docs/getting-started.md`、`docs/capabilities.md`、`docs/security.md`
- 控制连接可靠性接线：ACK 跟踪、心跳 keepalive notification、Write 重试计数、连接级 Metrics
- 客户端连接池：`WithQUICConfig` 覆盖池参数；会话结束 `Put` 归还连接
- 服务端 composite subscribe（控制轨 + 数据轨）；`SubscribeDataTrack` / `AttachSessionSubscribe`
- 数据轨道 e2e 集成测试；CI Go 1.25、coverage soft gate、`govulncheck`、Release 工作流

### Changed

- Example server 支持 `-stdio`
- Example client 演示 list/call/read capabilities
- Dockerfiles 更新到 Go 1.25

## v1.2.0 更新概述

### Changed

- Project structure optimization: improved directory structure to meet mature GitHub project standards
- Documentation enhancement: added project introduction, main features, and contribution guidelines
- Test coverage improvement: ensured all unit and integration tests pass
- Code quality optimization: passed go vet code quality check
- Updated version numbers in all example files and documentation

## v1.1.1 更新概述

### Changed

- Updated dependencies:
  - github.com/modelcontextprotocol/go-sdk: v1.4.0 → v1.5.0
  - github.com/quic-go/quic-go: v0.53.0 → v0.59.0
- Improved protocol compatibility and stability

## v1.1.0 更新概述

### Added

- Advanced data track features: support for streaming and batch processing
- Comprehensive error handling mechanism with fine-grained error types and codes
- Enhanced monitoring metrics with Prometheus support
- Configuration hot-reloading for runtime updates without service restart
- Multi-environment configuration support with default templates for dev/test/prod

### Changed

- Optimized QUIC connection pool performance with improved connection reuse strategy
- Enhanced stream management efficiency with optimized lifecycle management
- Improved heartbeat detection mechanism with adjusted intervals and timeout strategies
- Optimized message acknowledgment mechanism with improved ACK/NACK processing
- Enhanced retry mechanism with improved exponential backoff algorithm

### Fixed

- Memory leak issue in data track subscription
- Deadlock issue in QUIC connection pool under high concurrency
- False positive issue in heartbeat detection during network instability
- Race condition issue in monitoring metrics under multi-threaded environment
- Type conversion error in configuration file parsing

## v1.0.0 更新概述

### Added

- QUIC connection pool with connection reuse, idle timeout, and automatic cleanup
- QUIC stream management with lifecycle tracking and statistics
- QUIC configuration optimization with fluent builder pattern
- Data track support for resources, tools, and notifications
- Reliability features: message acknowledgment, heartbeat detection, retry mechanism, monitoring metrics
- Multi-connection support on server-side
- Comprehensive error handling with detailed error information
- Complete test suite: unit tests, integration tests, performance benchmarks
- Detailed documentation: API docs and design docs

### Fixed

- Goroutine leak in heartbeat detection
- Session ID generation fallback issue when crypto/rand fails
- Concurrent safety issue in DefaultMetrics() function
- context.Cause() usage issue in track_base.go
- InitialMaxRequestID default value issue in discovery mechanism
- Integration test timeout issue

## v0.7.1 更新概述

### Fixed

- Goroutine leak in heartbeat detection with proper goroutine wait on timeout
- Session ID generation fallback issue using timestamp+counter for unique ID
- Concurrent safety issue in DefaultMetrics() function with read-write lock
- context.Cause() usage issue in track_base.go using ctx.Err() for compatibility

### Changed

- Updated version number in discovery response from 0.4.0 to 0.7.0

## v0.7.0 更新概述

### Added

- QUIC connection pool with connection reuse, idle timeout, and automatic cleanup
- QUIC stream manager for unified stream lifecycle management
- QUIC configuration builder with fluent interface
- Unit tests for QUIC connection pool, stream manager, and configuration builder

## v0.6.0 更新概述

### Changed

- Exported DataTrackHandler.SessionID() method for session ID access
- Exported Retry.CalculateDelay() and Retry.IsRetryable() methods for external calls
- Fixed heartbeat detection not stopping correctly on context cancellation
- Fixed error rate and throughput calculation issues in tests

### Security

- Extended TLS certificate validity from 24 hours to 365 days
- Added TLS minimum version restriction (TLS 1.2)
- Fixed heartbeat goroutine leak risk
- Fixed non-thread-safe random number generator in retry mechanism
- Added message size limit (10MB) to prevent memory attacks

### Added

- Unit tests for message acknowledgment, heartbeat detection, retry mechanism, monitoring metrics, and data track handler

## v0.5.0 更新概述

### Added

- Reliability features:
  - Message acknowledgment mechanism (ACK/NACK) with timeout detection
  - Heartbeat detection to keep connections active and detect disconnections
  - Retry mechanism with exponential backoff strategy
  - Monitoring metrics for message statistics, throughput, error rates, etc.
- Feature example code

## v0.4.0 更新概述

### Added

- Data track support for resources, tools, and notifications
- data_tracks.go defining data track types and namespaces
- data_track_handler.go implementing data track handler with subscribe and publish mechanisms
- Extended discovery response to include data track information
- Data track example code
- Data track related error types
- Updated documentation with data track usage instructions

## v0.3.0 更新概述

### Changed

- Unified entry point: NewMoqTransport(RoleServer/RoleClient)
- TLS default behavior: self-signed certificate on server, InsecureSkipVerify on client
- ALPN default to moq-00 (based on moqtransport draft-11)
- Merged duplicate code from internal/transport/ to pkg/moqttransport/
- Created docs/ directory with detailed API and design documentation
- Unified test directory structure under test/

### Added

- Enhanced error handling
- Detailed code comments
- Optimized dependency management
- Increased test coverage
- Logging functionality with different log levels
- Configuration file support (YAML/JSON format)
- Environment variable support
- Comprehensive error types and error handling mechanism
- Performance benchmark support
- CI/CD automation integration
- Example configuration files and environment variable templates

## v0.2.0 更新概述

### Changed

- Unified entry point: NewMoqTransport(RoleServer/RoleClient)
- TLS default behavior: self-signed certificate on server, InsecureSkipVerify on client
- ALPN default to moq-00 (based on moqtransport draft-11)
- Merged duplicate code from internal/transport/ to pkg/moqttransport/
- Created docs/ directory with detailed API and design documentation
- Unified test directory structure under test/

### Added

- Enhanced error handling
- Detailed code comments
- Optimized dependency management
- Increased test coverage
- Logging functionality with different log levels
- Configuration file support (YAML/JSON format)
- Environment variable support
- Comprehensive error types and error handling mechanism
- Performance benchmark support
- CI/CD automation integration
- Example configuration files and environment variable templates
