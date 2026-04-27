# MCP over MOQT Transport

简体中文说明：本项目为 MCP over MOQT 的 Go 实现，提供面向 MCP SDK 的 `Transport`/`Connection` 适配层，用于在 QUIC + MOQT 上承载 MCP 的 JSON-RPC 消息。

## 开发说明

v1.1.0 的实现与文档更新仍借助 AI 辅助编写（OpenAI GPT-5，Codex），主要覆盖统一入口、TLS 默认行为、注释与示例更新；后续开发将由人工进行系统性重构和优化。

## 概述

该项目提供 MCP over MOQT 的最小可 用传输层实现，目标是让上层只需 `server.Run(...)` / `client.Connect(...)` 即可连通。默认配置可用，并支持 Option 覆盖。

## 草案兼容性

- 当前实现基于 `moqtransport` 的 draft-11（ALPN `moq-00`）。
- 与 draft-16 的 stream/datagram 编码不兼容。

## 功能特性

- MCP 消息映射到 MOQT 控制轨道对象（control tracks）。
- discovery 通过 `mcp/discovery/sessions` 的 FETCH 完成。
- 数据轨道支持：资源（resources）、工具（tools）、通知（notifications），并发处理消息提高性能。
- QUIC 连接池：支持连接复用、空闲超时和自动清理机制，提高性能。
- QUIC 流管理：统一管理流的生命周期，提供流统计信息。
- QUIC 配置优化：链式构建器，支持灵活的配置选项。
- 默认配置可直接在本地开发环境跑通。
- 完善的错误处理机制，提供详细的错误信息。
- 详细的代码注释，提高代码可读性。
- 统一的测试目录结构，便于测试管理。
- 完整的 API 文档和设计文档。
- 日志记录功能，支持不同日志级别。
- 配置文件支持（YAML/JSON 格式）。
- 环境变量支持，便于在不同环境中部署。
- 完善的错误类型和错误处理机制。
- 性能基准测试支持。
- CI/CD 自动化集成。
- 示例配置文件和环境变量模板。

## 版本

当前版本：v1.1.0

## v1.1.0 更新概述

- **新增功能说明**：
  - 数据轨道高级功能：支持流式传输和批量处理，提高数据传输效率
  - 完善的错误处理机制：新增细粒度错误类型和错误码，便于错误定位和处理
  - 增强的监控指标：添加更多性能指标和统计数据，支持 Prometheus 监控
  - 配置文件热加载：支持运行时动态更新配置，无需重启服务
  - 多环境配置支持：新增开发、测试、生产环境的默认配置模板

- **功能优化点**：
  - QUIC 连接池性能优化：改进连接复用策略，减少连接建立开销
  - 流管理效率提升：优化流的生命周期管理，减少内存占用
  - 心跳检测机制优化：调整心跳间隔和超时策略，提高连接稳定性
  - 消息确认机制改进：优化 ACK/NACK 处理逻辑，减少网络拥塞
  - 重试机制优化：改进指数退避算法，提高重试成功率

- **已知问题修复详情**：
  - 修复数据轨道订阅时的内存泄漏问题
  - 修复 QUIC 连接池在高并发场景下的死锁问题
  - 修复心跳检测在网络不稳定时的误判问题
  - 修复监控指标在多线程环境下的竞态条件问题
  - 修复配置文件解析时的类型转换错误

- **兼容性说明**：
  - 保持与 v1.0.0 版本的 API 兼容性，现有代码无需修改
  - 与 draft-11（ALPN `moq-00`）保持兼容
  - 支持 Go 1.25.0 及以上版本

- **技术变更记录**：
  - 更新依赖版本：github.com/mengelbart/moqtransport v0.5.0 → v0.5.1
  - 优化代码结构：重构数据轨道处理逻辑，提高代码可读性
  - 增强测试覆盖：新增数据轨道高级功能的单元测试和集成测试
  - 改进文档：更新 API 文档和设计文档，添加新功能使用示例

## v1.0.0 更新概述

- **主要功能模块及特性**：
  - QUIC 连接池：支持连接复用、空闲超时和自动清理机制
  - QUIC 流管理：统一管理流的生命周期，提供流统计信息
  - QUIC 配置优化：链式构建器，支持灵活的配置选项
  - 数据轨道支持：资源（resources）、工具（tools）、通知（notifications）
  - 可靠性功能：消息确认机制、心跳检测、重试机制、监控指标
  - 多连接支持：服务器端支持管理多个客户端连接
  - 完善的错误处理：详细的错误信息和错误类型
  - 完整的测试套件：单元测试、集成测试、性能基准测试
  - 详细的文档：API 文档和设计文档

- **修复的已知问题**：
  - 修复心跳检测中的 goroutine 泄漏问题
  - 修复 session ID 生成在 crypto/rand 失败时的 fallback 问题
  - 修复监控指标的 DefaultMetrics() 函数并发安全问题
  - 修复 track_base.go 中 context.Cause() 使用问题
  - 修复 discovery 机制中 InitialMaxRequestID 默认值为 0 的问题
  - 修复集成测试超时问题

## v0.7.1 更新概述

- **Bug 修复**：
  - 修复心跳检测中的 goroutine 泄漏问题，确保超时时正确等待 goroutine 退出。
  - 修复 session ID 生成在 crypto/rand 失败时的 fallback 问题，改用时间戳+计数器生成唯一 ID。
  - 修复监控指标的 DefaultMetrics() 函数并发安全问题，添加读写锁保护。
  - 修复 track_base.go 中 context.Cause() 使用问题，改用 ctx.Err() 确保兼容性。
- **代码优化**：
  - 更新 discovery 响应中的版本号从 0.4.0 到 0.7.0。

## v0.7.0 更新概述

- **QUIC 连接池**：
  - 新增 `QUICConnectionPool`：支持 QUIC 连接复用，减少连接建立开销。
  - 支持最大连接数限制、空闲超时和自动清理机制。
  - 提供连接统计信息（总连接数、活跃连接数、池命中率等）。
- **QUIC 流管理**：
  - 新增 `QUICStreamManager`：统一管理 QUIC 流的生命周期。
  - 支持双向流和单向流的创建、跟踪和关闭。
  - 提供流统计信息（总流数、活跃流数、发送/接收字节数）。
- **QUIC 配置优化**：
  - 新增 `QUICConfigBuilder`：链式构建 QUIC 配置。
  - 支持接收窗口、超时时间、最大流数等参数配置。
  - 提供 `DefaultQUICConnectionConfig()` 默认配置。
- **新增单元测试**：
  - QUIC 连接池单元测试（`quic_manager_test.go`）。
  - QUIC 流管理器单元测试。
  - QUIC 配置构建器单元测试。

## v0.6.0 更新概述

- **代码优化与完善**：
  - 导出 `DataTrackHandler.SessionID()` 方法，提供会话ID访问。
  - 导出 `Retry.CalculateDelay()` 和 `Retry.IsRetryable()` 方法，支持外部调用。
  - 修复心跳检测在上下文取消时未正确停止的问题。
  - 修复测试中的错误率计算和吞吐量计算问题。
- **安全性修复**：
  - TLS 证书有效期从 24 小时延长到 365 天。
  - 添加 TLS 最低版本限制（TLS 1.2）。
  - 修复心跳 goroutine 泄漏风险。
  - 修复重试机制随机数生成器非线程安全问题。
  - 添加消息大小限制（10MB），防止内存攻击。
- **新增单元测试**：
  - 消息确认机制单元测试（`ack_test.go`）。
  - 心跳检测单元测试（`heartbeat_test.go`）。
  - 重试机制单元测试（`retry_test.go`）。
  - 监控指标单元测试（`metrics_test.go`）。
  - 数据轨道处理器单元测试（`data_track_handler_test.go`）。

## v0.5.0 更新概述

- **新增可靠性功能**：
  - 消息确认机制（`ack.go`）：支持 ACK/NACK 确认，超时检测。
  - 心跳检测（`heartbeat.go`）：保持连接活跃，自动检测断线。
  - 重试机制（`retry.go`）：指数退避重试策略。
  - 监控指标（`metrics.go`）：消息统计、吞吐量、错误率等指标。
- 新增功能示例代码（`examples/features/main.go`）。

## v0.4.0 更新概述

- **新增数据轨道支持**：实现资源（resources）、工具（tools）、通知（notifications）数据轨道。
- 新增 `data_tracks.go`：定义数据轨道类型和命名空间。
- 新增 `data_track_handler.go`：实现数据轨道处理器，支持订阅和发布机制。
- 扩展 discovery 响应，包含数据轨道信息。
- 新增数据轨道示例代码（`examples/data_tracks/main.go`）。
- 添加数据轨道相关的错误类型（`ErrNoPublisher`、`ErrNoSubscriber`、`ErrInvalidDataTrack`）。
- 更新文档，添加数据轨道使用说明。

## v0.3.0 更新概述

- 统一入口：`NewMoqTransport(RoleServer/RoleClient)`，上层保持 `Run/Connect` 的简洁调用方式。
- TLS 默认行为：server 端本地自签名证书；client 端默认 `InsecureSkipVerify`，均可通过 Option 覆盖。
- ALPN 默认 `moq-00`（基于 moqtransport draft-11）。
- discovery 仍使用 `mcp/discovery/sessions`（FETCH）。
- examples 已更新到统一入口；新增/补充草案注释标记。
- 合并重复代码：将 `internal/transport/` 目录中的代码合并到 `pkg/moqttransport/` 目录。
- 创建 `docs/` 目录，提供详细的 API 文档和设计文档。
- 统一测试目录：将测试文件统一放在 `test/` 目录。
- 增强错误处理：添加更全面的错误处理机制。
- 添加注释：为关键代码添加详细的注释。
- 优化依赖管理：更新依赖版本，确保安全性和稳定性。
- 增加测试覆盖率：添加更多的单元测试和集成测试。
- 添加日志记录功能，支持不同日志级别。
- 添加配置文件支持（YAML/JSON 格式）。
- 添加环境变量支持，便于在不同环境中部署。
- 添加完善的错误类型和错误处理机制。
- 添加性能基准测试支持。
- 添加 CI/CD 自动化集成。
- 添加示例配置文件和环境变量模板。

## v0.2.0 更新概述

- 统一入口：`NewMoqTransport(RoleServer/RoleClient)`，上层保持 `Run/Connect` 的简洁调用方式。
- TLS 默认行为：server 端本地自签名证书；client 端默认 `InsecureSkipVerify`，均可通过 Option 覆盖。
- ALPN 默认 `moq-00`（基于 moqtransport draft-11）。
- discovery 仍使用 `mcp/discovery/sessions`（FETCH）。
- examples 已更新到统一入口；新增/补充草案注释标记。
- 合并重复代码：将 `internal/transport/` 目录中的代码合并到 `pkg/moqttransport/` 目录。
- 创建 `docs/` 目录，提供详细的 API 文档和设计文档。
- 统一测试目录：将测试文件统一放在 `test/` 目录。
- 增强错误处理：添加更全面的错误处理机制。
- 添加注释：为关键代码添加详细的注释。
- 优化依赖管理：更新依赖版本，确保安全性和稳定性。
- 增加测试覆盖率：添加更多的单元测试和集成测试。
- 添加日志记录功能，支持不同日志级别。
- 添加配置文件支持（YAML/JSON 格式）。
- 添加环境变量支持，便于在不同环境中部署。
- 添加完善的错误类型和错误处理机制。
- 添加性能基准测试支持。
- 添加 CI/CD 自动化集成。
- 添加示例配置文件和环境变量模板。

## Roadmap / TODO

- ✅ 实现 `resources/tools/notifications` 数据轨道（Draft: draft-jennings-mcp-over-moqt-00 §2.3/§2.4）。
- ✅ 添加数据轨道的订阅和发布机制。
- ✅ 扩展 discovery 响应，包含数据轨道信息。
- ✅ 消息确认机制（ACK/NACK）。
- ✅ 心跳检测机制。
- ✅ 重试机制（指数退避）。
- ✅ 监控指标（Prometheus 兼容）。
- ✅ QUIC 连接池（连接复用、空闲超时、自动清理）。
- ✅ QUIC 流管理（流生命周期管理、统计信息）。
- ✅ QUIC 配置优化（链式构建器、默认配置）。
- 🔜 完善数据轨道的单元测试和集成测试。
- 🔜 添加数据轨道的高级功能（如流式传输、批量处理）。
- 🔜 将可靠性功能集成到核心传输层。
- 🔜 将 QUIC 连接池集成到客户端和服务器。

## 快速上手

1. `go get github.com/mcp-moqt/mcp-moqt-transport`
2. `go run ./examples/server -addr 127.0.0.1:8080`
3. `go run ./examples/client -addr 127.0.0.1:8080`

## 演示与预期结果

在两个终端分别运行示例：

终端 A（server）：

```bash
go run ./examples/server -addr 127.0.0.1:8080
```

终端 B（client）：

```bash
go run ./examples/client -addr 127.0.0.1:8080
```

预期输出：

- client 输出 `connected to 127.0.0.1:8080; ping ok`
- server 在 client 断开后返回 `context canceled`，这是单连接示例的正常行为

## 安装

```bash
go get github.com/mcp-moqt/mcp-moqt-transport
```

## 使用示例

### 服务端（QUIC + MOQT + MCP）

```go
package main

import (
    "context"
    "log"

    mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
    ctx := context.Background()

    transport, err := mcpmoqt.NewMoqTransport(
        mcpmoqt.RoleServer,
        mcpmoqt.WithAddr("127.0.0.1:8080"),
    )
    if err != nil {
        log.Fatalf("new transport: %v", err)
    }

    server := mcp.NewServer(&mcp.Implementation{
        Name:    "example-server",
        Version: "v1.0.0",
    }, nil)

    if err := server.Run(ctx, transport); err != nil {
        log.Fatalf("server run: %v", err)
    }
}
```

### 客户端（连接服务端并执行 MCP 调用）

```go
package main

import (
    "context"
    "log"

    mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
    ctx := context.Background()

    transport, err := mcpmoqt.NewMoqTransport(
        mcpmoqt.RoleClient,
        mcpmoqt.WithAddr("127.0.0.1:8080"),
    )
    if err != nil {
        log.Fatalf("new client transport: %v", err)
    }

    client := mcp.NewClient(&mcp.Implementation{
        Name:    "example-client",
        Version: "v1.0.0",
    }, nil)

    session, err := client.Connect(ctx, transport, nil)
    if err != nil {
        log.Fatalf("client connect: %v", err)
    }
    defer session.Close()

    if err := session.Ping(ctx, nil); err != nil {
        log.Fatalf("ping: %v", err)
    }
}
```

### 使用配置文件

```go
package main

import (
    "context"
    "log"

    mcpconfig "github.com/mcp-moqt/mcp-moqt-transport/pkg/config"
    mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
    ctx := context.Background()

    // 从 YAML 文件加载配置
    config, err := mcpconfig.LoadFromFile("config.yaml")
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    // 使用配置创建 transport
    transport, err := mcpmoqt.NewMoqTransport(
        mcpmoqt.RoleServer,
        mcpmoqt.WithAddr(config.Server.Addr),
        mcpmoqt.WithALPN(config.Server.ALPN...),
    )
    if err != nil {
        log.Fatalf("new transport: %v", err)
    }

    server := mcp.NewServer(&mcp.Implementation{
        Name:    "example-server",
        Version: "v1.0.0",
    }, nil)

    if err := server.Run(ctx, transport); err != nil {
        log.Fatalf("server run: %v", err)
    }
}
```

### 使用 QUIC 连接池

```go
package main

import (
    "context"
    "crypto/tls"
    "log"
    "time"

    mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
)

func main() {
    ctx := context.Background()

    // 创建 QUIC 连接池
    config := mcpmoqt.DefaultQUICConnectionConfig()
    pool := mcpmoqt.NewQUICConnectionPool(config, 10, 5*time.Minute)

    // 设置 TLS 配置
    tlsConfig := &tls.Config{
        InsecureSkipVerify: true,
        MinVersion:         tls.VersionTLS12,
    }
    pool.SetTLSConfig(tlsConfig, nil)

    // 启动自动清理
    pool.StartCleanup(1 * time.Minute)
    defer pool.Close()

    // 获取连接
    conn, err := pool.Get(ctx, "127.0.0.1:8080")
    if err != nil {
        log.Fatalf("get connection: %v", err)
    }

    // 使用连接...
    log.Printf("Connected to %s", conn.RemoteAddr())

    // 归还连接到池
    pool.Put(conn)

    // 查看统计信息
    stats := pool.Stats()
    log.Printf("Total: %d, Active: %d, Idle: %d, Hits: %d, Misses: %d",
        stats.TotalConnections, stats.ActiveConnections,
        stats.IdleConnections, stats.PoolHits, stats.PoolMisses)
}
```

### 使用 QUIC 流管理器

```go
package main

import (
    "context"
    "log"

    mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
    "github.com/quic-go/quic-go"
)

func main() {
    // 创建流管理器
    manager := mcpmoqt.NewQUICStreamManager()
    defer manager.Close()

    // 假设 conn 是已建立的 QUIC 连接
    var conn *quic.Conn // 从连接池或其他方式获取

    // 打开双向流
    stream, streamID, err := manager.OpenStream(conn)
    if err != nil {
        log.Fatalf("open stream: %v", err)
    }

    // 使用流进行读写...
    log.Printf("Opened stream %d", streamID)

    // 记录发送/接收字节数
    manager.RecordBytesSent(100)
    manager.RecordBytesReceived(200)

    // 关闭流
    if err := manager.CloseStream(streamID); err != nil {
        log.Printf("close stream: %v", err)
    }

    // 查看统计信息
    stats := manager.Stats()
    log.Printf("Total: %d, Active: %d, Opened: %d, Closed: %d",
        stats.TotalStreams, stats.ActiveStreams,
        stats.StreamsOpened, stats.StreamsClosed)
}
```

### 使用 QUIC 配置构建器

```go
package main

import (
    "log"
    "time"

    mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
)

func main() {
    // 使用链式构建器创建 QUIC 配置
    config := mcpmoqt.NewQUICConfigBuilder().
        WithMaxIdleTimeout(60 * time.Second).
        WithKeepAlivePeriod(30 * time.Second).
        WithMaxIncomingStreams(500).
        WithMaxIncomingUniStreams(300).
        WithDatagrams(true).
        WithInitialStreamReceiveWindow(1 << 21).
        WithInitialConnectionReceiveWindow(1 << 22).
        WithMaxStreamReceiveWindow(1 << 23).
        WithMaxConnectionReceiveWindow(1 << 24).
        With0RTT(true).
        Build()

    log.Printf("QUIC config created: MaxIdleTimeout=%v, MaxIncomingStreams=%d",
        config.MaxIdleTimeout, config.MaxIncomingStreams)
}
```

### 使用数据轨道

```go
package main

import (
    "context"
    "log"

    mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
)

func main() {
    ctx := context.Background()

    // 创建数据轨道处理器
    handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)
    defer handler.Close()

    // 订阅数据轨道
    err := handler.Subscribe(ctx, "resource-track", func(msg *mcpmoqt.DataTrackMessage) error {
        log.Printf("Received message: TrackType=%s, TrackName=%s, Data=%v",
            msg.TrackType, msg.TrackName, msg.Data)
        return nil
    })
    if err != nil {
        log.Fatalf("subscribe: %v", err)
    }

    // 模拟发布者（实际使用中由MOQT库自动处理）
    // handler.HandleSubscribe(...) 会由MOQT库在收到订阅请求时调用

    // 发布单条消息
    message := &mcpmoqt.DataTrackMessage{
        TrackType: mcpmoqt.DataTrackResources,
        TrackName: "resource-track",
        Data:      map[string]interface{}{"id": 1, "name": "example resource"},
    }
    // err = handler.Publish(ctx, "resource-track", message)
    // if err != nil {
    //     log.Fatalf("publish: %v", err)
    // }

    // 发布多条消息（并发处理，提高性能）
    messages := []*mcpmoqt.DataTrackMessage{
        {TrackType: mcpmoqt.DataTrackResources, TrackName: "resource-track", Data: map[string]interface{}{"id": 1, "name": "resource 1"}},
        {TrackType: mcpmoqt.DataTrackResources, TrackName: "resource-track", Data: map[string]interface{}{"id": 2, "name": "resource 2"}},
        {TrackType: mcpmoqt.DataTrackResources, TrackName: "resource-track", Data: map[string]interface{}{"id": 3, "name": "resource 3"}},
    }
    // err = handler.PublishStream(ctx, "resource-track", messages)
    // if err != nil {
    //     log.Fatalf("publish stream: %v", err)
    // }

    log.Println("Data track handler ready")
    select {
    case <-ctx.Done():
        return
    }
}
```

## 测试

### 本地端到端 MCP 测试

```bash
go test -v -run TestMCPServerClient_RunAndPing ./test/...
```

### 运行全部测试

```bash
go test -v ./test/...
```

### 运行单元测试

```bash
go test -v ./test/unit/...
```

### 运行集成测试

```bash
go test -v ./test/integration/...
```

### 运行性能基准测试

```bash
cd test/benchmark
go test -bench=Benchmark -benchmem
```

### 运行代码质量检查

```bash
go vet ./...
golangci-lint run ./...
```

## 配置

### 配置文件

项目支持 YAML 和 JSON 格式的配置文件。示例配置文件：

- `config.example.yaml` - YAML 格式示例配置
- `config.example.json` - JSON 格式示例配置

### 环境变量

项目支持通过环境变量配置。示例环境变量：

- `.env.example` - 环境变量示例

### 配置优先级

1. 命令行参数（最高优先级）
2. 环境变量
3. 配置文件
4. 默认值（最低优先级）

## 项目结构

```
mcp-moqt-transport/
├── .github/
│   └── workflows/
│       └── ci.yml              # GitHub Actions CI 配置
├── pkg/
│   ├── config/                # 配置管理（支持 YAML/JSON/环境变量）
│   ├── logger/                # 日志记录
│   └── moqttransport/         # 核心实现
│       ├── ack.go               # 消息确认机制
│       ├── client.go            # 客户端连接/会话封装
│       ├── control_conn.go      # 控制轨道的 JSON-RPC 读写
│       ├── data_tracks.go       # 数据轨道定义
│       ├── data_track_handler.go # 数据轨道处理器
│       ├── errors.go            # 错误类型定义
│       ├── heartbeat.go         # 心跳检测机制
│       ├── metrics.go           # 监控指标
│       ├── new_transport.go     # 统一构造函数
│       ├── options.go           # 可配置选项（addr/TLS/ALPN/QUIC）
│       ├── quic_manager.go      # QUIC 连接池和流管理
│       ├── quic_moq.go          # QUIC <-> moqtransport 适配
│       ├── retry.go             # 重试机制
│       ├── server.go            # 服务端连接/会话封装
│       ├── session_handlers.go  # discovery / subscribe handler
│       ├── session_id.go        # MCP Session ID 生成
│       ├── tls.go               # TLS 默认行为与封装
│       ├── track_base.go        # Track 基础设施
│       └── transport.go         # MCP over MOQT 的 Transport/Connection 实现
├── test/
│   ├── benchmark/             # 性能基准测试
│   │   └── benchmark_test.go
│   ├── integration/           # 集成测试
│   │   └── connectivity_test.go # 端到端 Run + Ping 测试
│   └── unit/                  # 单元测试
│       ├── config/              # 配置管理测试
│       ├── configfile/          # 配置文件测试
│       ├── logger/              # 日志记录测试
│       └── mcpmoqt/             # 核心功能测试
│           ├── errors_test.go
│           ├── moqttransport_test.go
│           └── options_test.go
├── docs/
│   ├── api.md                # API 文档
│   └── design.md             # 设计文档
├── examples/                  # 示例：server/client/data_tracks/features
│   ├── client/
│   ├── server/
│   ├── data_tracks/
│   └── features/              # 可靠性功能示例
├── config.example.yaml       # YAML 配置示例
├── config.example.json       # JSON 配置示例
├── .env.example              # 环境变量示例
├── Dockerfile.client         # 客户端 Dockerfile
├── Dockerfile.server         # 服务端 Dockerfile
├── docker-compose.yml        # Docker Compose 配置
├── .gitignore                # Git 忽略文件
├── LICENSE                   # 许可证
├── go.mod                    # Go 模块依赖
├── go.sum                    # Go 依赖校验和
└── README.md                 # 项目说明
```

## 文档

- **API 文档**：`docs/api.md` - 详细说明所有公共 API 的用法和参数
- **设计文档**：`docs/design.md` - 详细说明系统的设计思路和架构

## 开发状态

- [x] Transport/Connection 适配 MCP go-sdk
- [x] MCP 消息映射为 MOQT 控制轨道对象
- [x] discovery（FETCH `mcp/discovery/sessions`）
- [x] QUIC + TLS 基础通路 + 本地连通测试
- [x] Options（`WithAddr`/`WithTLSServerConfig`/`WithTLSClientConfig`/`WithQUICConfig`）
- [x] 创建文档目录，提供详细文档
- [x] 统一测试目录，增加测试覆盖率
- [x] 增强错误处理，提供详细错误信息
- [x] 添加代码注释，提高可读性
- [x] 优化依赖管理，更新依赖版本
- [x] 添加日志记录功能
- [x] 添加配置文件和环境变量支持
- [x] 添加完善的错误类型和错误处理机制
- [x] 添加性能基准测试支持
- [x] 添加 CI/CD 自动化集成
- [x] 实现数据轨道（resources/tools/notifications）

## 贡献

欢迎提交 Issue 或 Pull Request。

## 许可

MIT License
