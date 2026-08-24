# MCP over MOQT Transport

[![CI](https://github.com/mcp-moqt/mcp-moqt-transport/actions/workflows/ci.yml/badge.svg)](https://github.com/mcp-moqt/mcp-moqt-transport/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mcp-moqt/mcp-moqt-transport)](https://goreportcard.com/report/github.com/mcp-moqt/mcp-moqt-transport)
[![Go Version](https://img.shields.io/badge/Go-1.25.13+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 项目简介

MCP over MOQT Transport 是一个用 **Go** 实现的 MCP（Model Context Protocol）传输层，面向需要在本地或分布式环境中高效承载 MCP 消息的场景。本项目基于 **QUIC** 与 **MOQT（Media over QUIC Transport）**，将 MCP 的 JSON-RPC 消息映射到控制轨道与数据轨道，并提供与官方 MCP Go SDK 兼容的 `Transport` / `Connection` 接口；上层应用只需调用 `server.Run(...)` / `client.Connect(...)` 即可建立会话并进行通信。

除高性能的 QUIC/MOQT 通路外，还支持 **stdio** 模式，便于接入 Cursor、Claude Desktop 等 MCP Host；同时内置工具、提示词、资源，以及 `mcp-moqt` CLI，方便安装、自检与本地演示。

当前版本：**v1.5.0** · 变更记录见 [CHANGELOG.md](./CHANGELOG.md)

## 传输模式说明

本项目提供两种 **互补** 的 MCP 接入方式，用途不同：

| 模式 | 命令 | 传输 | 典型场景 |
|------|------|------|----------|
| **stdio** | `mcp-moqt server -stdio` | 标准输入/输出（不经 QUIC） | Cursor、Claude Desktop、MCP 平台上架 |
| **QUIC/MOQT** | `mcp-moqt server -addr :8080` | QUIC + MOQT 控制轨/数据轨 | 服务间分布式 MCP、低延迟与多路复用 |

- **stdio**：面向 MCP Host 集成，满足平台审核（工具 / 提示词 / 资源），配置见 `configs/mcp/`。
- **QUIC/MOQT**：本项目的核心传输能力，适合自建部署与性能场景；生产配置见 [docs/deployment.md](./docs/deployment.md) 与 [docs/security.md](./docs/security.md)。

## 目录

- [项目简介](#项目简介)
- [传输模式说明](#传输模式说明)
- [特性](#特性)
- [草案兼容性](#草案兼容性)
- [快速开始](#快速开始)
- [安装](#安装)
- [MCP 服务能力](#mcp-服务能力tools--prompts--resources)
- [使用示例](#使用示例)
- [配置](#配置)
- [测试](#测试)
- [项目结构](#项目结构)
- [文档](#文档)
- [路线图](#路线图)
- [贡献](#贡献)
- [许可](#许可)

## 特性

- **高性能传输**：QUIC 多路复用、拥塞控制；可选 0-RTT（默认关闭）
- **可靠性**：ACK/NACK 跟踪、心跳 keepalive、Write 重试、连接级 Metrics
- **数据轨道**：resources / tools / notifications，支持 Publish / PublishStream / BatchPublish
- **MCP 能力完备**：内置工具、提示词、资源（见下方能力表）
- **双传输模式**：`stdio`（平台托管）与 QUIC/MOQT（高性能）
- **友好安装与 CLI**：安装脚本、`go install`、Make、Docker；`mcp-moqt`（server / client / doctor）
- **连接池**：客户端 QUIC 连接池（Get / Put）
- **灵活配置**：配置文件、环境变量、Option；默认配置可本地跑通
- **易于集成**：`server.Run(...)` / `client.Connect(...)` 即可连通 MCP SDK
- **多连接**：`Listen` + `Accept` / `Serve` 循环同时服务多个客户端
- **可观测性**：Prometheus 文本格式 `/metrics`（`WithMetricsAddr`）
- **配置热加载**：`config.WatchFile` 轮询检测配置变更

## 草案兼容性

- 基于 `moqtransport` **draft-11**（ALPN `moq-00`）
- 与 draft-16 的 stream/datagram 编码**不兼容**

## 快速开始

### stdio（Cursor / Claude Desktop / MCP Host）

```bash
mcp-moqt server -stdio
# 或
go run ./examples/server -stdio
```

配置样例：`configs/mcp/cursor.mcp.json`、`configs/mcp/claude_desktop.json`

### QUIC + MOQT（本地演示）

```bash
# 终端 A
mcp-moqt server -addr 127.0.0.1:8080
# 或：go run ./examples/server -addr 127.0.0.1:8080

# 终端 B
mcp-moqt client -addr 127.0.0.1:8080
# 或：go run ./examples/client -addr 127.0.0.1:8080
```

预期：client 输出 `connected ...; ping ok`；server 在 client 断开后返回 `context canceled`（单连接示例的正常行为）。

### 自检

```bash
mcp-moqt doctor
```

更完整的安装与 TLS 说明见英文文档 [docs/getting-started.md](./docs/getting-started.md)、[docs/security.md](./docs/security.md)（`docs/` 目录均为英文）。

## 安装

### 一键脚本（推荐）

```bash
# Linux / macOS（本地仓库）
bash scripts/install.sh

# Windows
powershell -File scripts/install.ps1
```

仓库已发布到 GitHub 时，也可：

```bash
curl -fsSL https://raw.githubusercontent.com/mcp-moqt/mcp-moqt-transport/main/scripts/install.sh | bash
```

### 使用 go install / Make / Docker

```bash
go install github.com/mcp-moqt/mcp-moqt-transport/cmd/mcp-moqt@latest
# 作为库依赖
go get github.com/mcp-moqt/mcp-moqt-transport

make install && make server   # 或 make stdio / make client
docker compose up --build
```

## MCP 服务能力（工具 / 提示词 / 资源）

示例服务与 `mcp-moqt server` 默认注册以下能力：

| 类型 | 名称 | 说明 |
|------|------|------|
| 工具 | `health_check` | 健康检查与运行时信息 |
| 工具 | `get_transport_info` | 传输层能力与草案信息 |
| 工具 | `list_data_tracks` | 列出数据轨道 |
| 工具 | `estimate_rtt` | 简易 RTT 估算演示 |
| 工具 | `validate_config` | 校验 YAML/JSON 配置文件 |
| 工具 | `export_metrics` | 导出传输层指标快照 |
| 提示词 | `debug_moqt_session` | 会话排查提示词 |
| 提示词 | `configure_transport` | 配置助手提示词 |
| 资源 | `moqt://docs/overview` | 服务概述 |
| 资源 | `moqt://config/example` | 示例 YAML 配置 |
| 资源 | `moqt://tracks/catalog` | 数据轨道目录 |
| 资源 | `moqt://install/quickstart` | 友好安装说明 |
| 资源 | `moqt://docs/deployment` | 生产部署指南 |
| 资源 | `moqt://docs/security` | 安全与 TLS 清单 |

```go
import "github.com/mcp-moqt/mcp-moqt-transport/pkg/mcpservice"

server := mcp.NewServer(&mcp.Implementation{
    Name:    mcpservice.ServiceName,
    Version: mcpservice.ServiceVersion,
}, nil)
mcpservice.RegisterCapabilities(server)
```

完整说明见 [docs/capabilities.md](./docs/capabilities.md)。

## 使用示例

### 服务端（QUIC + MOQT + MCP）

```go
package main

import (
    "context"
    "log"

    "github.com/mcp-moqt/mcp-moqt-transport/pkg/mcpservice"
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
        Name:    mcpservice.ServiceName,
        Version: mcpservice.ServiceVersion,
    }, nil)
    mcpservice.RegisterCapabilities(server)

    if err := server.Run(ctx, transport); err != nil {
        log.Fatalf("server run: %v", err)
    }
}
```

### 客户端

```go
package main

import (
    "context"
    "log"

    "github.com/mcp-moqt/mcp-moqt-transport/pkg/mcpservice"
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
        Version: mcpservice.ServiceVersion,
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

更多示例：

- `examples/server`、`examples/client`、`examples/data_tracks`、`examples/features`
- 配置文件 / 连接池 / 数据轨道等进阶用法见 [docs/api.md](./docs/api.md)

## 配置

| 来源 | 说明 |
|------|------|
| 命令行参数 | 最高优先级 |
| 环境变量 | 见 `.env.example`；`MCP_MOQT_METRICS_ADDR` 可开 `/metrics` |
| 配置文件 | `config.example.yaml` / `config.example.json`；可用 `config.WatchFile` 热加载 |
| 默认值 | 最低优先级 |

多连接服务示例：`mcp-moqt server -addr 127.0.0.1:8080 -multi -metrics 127.0.0.1:9090`

TLS 生产环境配置见 [docs/security.md](./docs/security.md)。

## 测试

```bash
go test ./test/...                          # 全部
go test -v ./test/unit/...                  # 单元
go test -v ./test/integration/...           # 集成
go test -bench=Benchmark -benchmem ./test/benchmark/...
go vet ./...
```

## 项目结构

```
mcp-moqt-transport/
├── cmd/mcp-moqt/          # CLI：server / client / doctor
├── pkg/
│   ├── config/            # 配置（YAML/JSON/环境变量）
│   ├── logger/            # 日志
│   ├── mcpservice/        # Tools / Prompts / Resources
│   ├── moqttransport/     # 核心传输实现
│   └── version/           # 统一版本号
├── scripts/               # install.sh / install.ps1
├── configs/mcp/           # MCP Host 配置样例
├── examples/              # server / client / data_tracks / features
├── test/                  # unit / integration / benchmark
├── docs/                  # 文档
├── Makefile
├── docker-compose.yml
└── CHANGELOG.md
```

## 文档

| 文档 | 说明 |
|------|------|
| [docs/getting-started.md](./docs/getting-started.md) | 快速开始（英文） |
| [docs/capabilities.md](./docs/capabilities.md) | 工具 / 提示词 / 资源（英文） |
| [docs/api.md](./docs/api.md) | 公共 API（英文） |
| [docs/design.md](./docs/design.md) | 设计与架构（英文） |
| [docs/deployment.md](./docs/deployment.md) | 生产部署（英文） |
| [docs/security.md](./docs/security.md) | 安全与 TLS 清单（英文） |
| [configs/grafana/](./configs/grafana/) | Grafana 仪表盘样例 |
| [docs/configuration.md](./docs/configuration.md) | 配置、热加载、指标、多连接（英文） |
| [configs/mcp/](./configs/mcp/) | Cursor / Claude Desktop 配置样例 |
| [CHANGELOG.md](./CHANGELOG.md) | 版本变更 |

## 路线图

**已完成**

- [x] MCP over MOQT 控制轨 + discovery
- [x] 数据轨道（resources / tools / notifications）与 e2e 订阅
- [x] ACK / 心跳 / 重试接入控制连接
- [x] 客户端 QUIC 连接池；stdio 与 MCP Host 配置
- [x] 工具 / 提示词 / 资源；安装脚本与 CLI
- [x] 多连接 Accept / Serve 循环
- [x] Prometheus `/metrics` HTTP 暴露
- [x] 配置文件热加载（`config.WatchFile`）
- [x] 生产 TLS 配置与部署文档；Grafana 样例；`doctor` 网络探测

**计划中**

- （无）近期 P0–P2 项已在 v1.5.0 完成

## 贡献

欢迎提交 Issue 与 Pull Request。开发前请运行：

```bash
go fmt ./...
go test ./test/...
```

详细流程与代码规范见 [CONTRIBUTING.md](./CONTRIBUTING.md)。

## 许可

[MIT 许可证](./LICENSE)
