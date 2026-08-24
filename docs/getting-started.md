# Getting Started / 快速开始

## English

### Install

```bash
# From a local checkout
bash scripts/install.sh
# or
go install github.com/mcp-moqt/mcp-moqt-transport/cmd/mcp-moqt@latest
```

Windows:

```powershell
powershell -File scripts/install.ps1
```

### Run (stdio — Cursor / Claude Desktop / MCP hosts)

```bash
mcp-moqt server -stdio
```

Sample host configs:

- `configs/mcp/cursor.mcp.json`
- `configs/mcp/claude_desktop.json`
- `configs/mcp/generic.mcp.json`

### Run (QUIC + MOQT)

```bash
mcp-moqt server -addr 127.0.0.1:8080
mcp-moqt client -addr 127.0.0.1:8080
```

### Doctor

```bash
mcp-moqt doctor
```

Validates Go toolchain presence and runs an in-memory smoke test for tools / prompts / resources.

### TLS note

Local defaults use a self-signed server certificate and client `InsecureSkipVerify`.
For production, pass real TLS configs via `WithTLSServerConfig` / `WithTLSClientConfig`. See [security.md](./security.md).

---

## 简体中文

### 安装

```bash
bash scripts/install.sh
# 或
go install github.com/mcp-moqt/mcp-moqt-transport/cmd/mcp-moqt@latest
```

Windows：

```powershell
powershell -File scripts/install.ps1
```

### 运行（stdio，适合 Cursor / 平台托管）

```bash
mcp-moqt server -stdio
```

配置样例见 `configs/mcp/`。

### 运行（QUIC + MOQT）

```bash
mcp-moqt server -addr 127.0.0.1:8080
mcp-moqt client -addr 127.0.0.1:8080
```

### 自检

```bash
mcp-moqt doctor
```

### TLS 说明

本地默认：服务端自签名证书，客户端跳过校验。生产环境请使用正式证书，详见 [security.md](./security.md)。
