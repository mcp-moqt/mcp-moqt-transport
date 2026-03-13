# MCP over MOQT Transport

简体中文说明：本项目为 MCP over MOQT 的 Go 实现，提供面向 MCP SDK 的 `Transport`/`Connection` 适配层，用于在 QUIC + MOQT 上承载 MCP 的 JSON-RPC 消息。

## 开发说明

v0.2.0 的实现与文档更新仍由 AI 辅助编写（OpenAI GPT-5，Codex），主要覆盖统一入口、TLS 默认行为、注释与示例更新；后续开发将由人工进行系统性重构和优化。

## 概述

该项目提供 MCP over MOQT 的最小可用传输层实现，目标是让上层只需 `server.Run(...)` / `client.Connect(...)` 即可连通。默认配置可用，并支持 Option 覆盖。

## 草案兼容性

- 当前实现基于 `moqtransport` 的 draft-11（ALPN `moq-00`）。
- 与 draft-16 的 stream/datagram 编码不兼容。

## 功能特性

- MCP 消息映射到 MOQT 控制轨道对象（control tracks）。
- discovery 通过 `mcp/discovery/sessions` 的 FETCH 完成。
- 默认配置可直接在本地开发环境跑通。
- 完善的错误处理机制，提供详细的错误信息。
- 详细的代码注释，提高代码可读性。
- 统一的测试目录结构，便于测试管理。
- 完整的 API 文档和设计文档。
- 日志记录功能，支持不同日志级别。
- 配置文件支持（YAML/JSON 格式）。
- 环境变量支持，便于在不同环境中部署。

## 版本

当前版本：v0.2.0

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

## Roadmap / TODO

- 预留 `resources/tools/notifications` 轨道（Draft: draft-jennings-mcp-over-moqt-00 §2.3/§2.4）。
- v0.2.0 仅保留轨道命名与 TODO 注释，不实现具体业务语义；后续将补充订阅/发布与消息结构。
- 添加配置文件和环境变量支持，提高灵活性。

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
        Version: "v0.2.0",
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
        Version: "v0.2.0",
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

## 项目结构

```
mcp-moqt-transport/
pkg/
  config/                # 配置管理（支持 YAML/JSON/环境变量）
  logger/                # 日志记录
  moqttransport/         # 核心实现
    client.go            # 客户端连接/会话封装
    control_conn.go      # 控制轨道的 JSON-RPC 读写
    new_transport.go     # 统一构造函数
    options.go           # 可配置选项（addr/TLS/ALPN/QUIC）
    quic_moq.go          # QUIC <-> moqtransport 适配
    server.go            # 服务端连接/会话封装
    session_handlers.go  # discovery / subscribe handler
    session_id.go        # MCP Session ID 生成
    tls.go               # TLS 默认行为与封装
    transport.go         # MCP over MOQT 的 Transport/Connection 实现

test/
  integration/           # 集成测试
    connectivity_test.go # 端到端 Run + Ping 测试
  unit/                  # 单元测试
    config/              # 配置管理测试
    configfile/          # 配置文件测试
    logger/              # 日志记录测试
    mcpmoqt/             # 核心功能测试
    options_test.go      # 配置选项测试

docs/
  api.md                # API 文档
  design.md             # 设计文档

examples/               # 示例：server/client
README.md               # 项目说明
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
- [ ] resources/tools/notifications 轨道与语义

## 许可

MIT License

## 贡献

欢迎提交 Issue 或 Pull Request。
