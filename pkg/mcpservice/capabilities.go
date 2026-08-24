// Package mcpservice registers practical MCP tools, prompts, and resources
// for the MCP-over-MOQT transport demo server.
package mcpservice

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	ServiceName    = "mcp-moqt-transport"
	ServiceVersion = "1.3.0"
)

// RegisterCapabilities adds tools, prompts, and resources required for a
// full-featured MCP service listing (tools / prompts / resources).
func RegisterCapabilities(server *mcp.Server) {
	registerTools(server)
	registerPrompts(server)
	registerResources(server)
}

func registerTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "health_check",
		Description: "检查 MCP over MOQT 传输层健康状态，返回运行时与协议信息",
	}, healthCheck)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_transport_info",
		Description: "获取传输层能力说明：ALPN、数据轨道、可靠性特性等",
	}, getTransportInfo)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_data_tracks",
		Description: "列出可用的 MOQT 数据轨道类型（resources / tools / notifications）",
	}, listDataTracks)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "estimate_rtt",
		Description: "根据往返样本估算简易 RTT（演示工具，不依赖真实网络探测）",
	}, estimateRTT)
}

type emptyInput struct{}

type healthOutput struct {
	Status    string `json:"status" jsonschema:"服务状态：ok / degraded"`
	Service   string `json:"service" jsonschema:"服务名称"`
	Version   string `json:"version" jsonschema:"服务版本"`
	GoVersion string `json:"go_version" jsonschema:"Go 运行时版本"`
	OS        string `json:"os" jsonschema:"操作系统"`
	Arch      string `json:"arch" jsonschema:"CPU 架构"`
	Timestamp string `json:"timestamp" jsonschema:"检查时间（RFC3339）"`
}

func healthCheck(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, healthOutput, error) {
	_ = ctx
	out := healthOutput{
		Status:    "ok",
		Service:   ServiceName,
		Version:   ServiceVersion,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	return nil, out, nil
}

type transportInfoInput struct {
	IncludeTracks bool `json:"include_tracks,omitempty" jsonschema:"是否在结果中附带数据轨道列表"`
}

type transportInfoOutput struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	ALPN        string   `json:"alpn"`
	Draft       string   `json:"draft"`
	Features    []string `json:"features"`
	DataTracks  []string `json:"data_tracks,omitempty"`
	Description string   `json:"description"`
}

func getTransportInfo(ctx context.Context, _ *mcp.CallToolRequest, in transportInfoInput) (*mcp.CallToolResult, transportInfoOutput, error) {
	_ = ctx
	out := transportInfoOutput{
		Name:    ServiceName,
		Version: ServiceVersion,
		ALPN:    "moq-00",
		Draft:   "moqtransport draft-11 / draft-jennings-mcp-over-moqt-00",
		Features: []string{
			"quic-multiplex",
			"0-rtt",
			"ack-nack",
			"heartbeat",
			"retry",
			"connection-pool",
			"data-tracks",
		},
		Description: "高性能 MCP 传输层，基于 QUIC + MOQT，支持控制轨道与数据轨道。",
	}
	if in.IncludeTracks {
		out.DataTracks = []string{"resources", "tools", "notifications"}
	}
	return nil, out, nil
}

type listTracksOutput struct {
	Tracks []trackInfo `json:"tracks"`
}

type trackInfo struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
}

func listDataTracks(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, listTracksOutput, error) {
	_ = ctx
	return nil, listTracksOutput{
		Tracks: []trackInfo{
			{Name: "resources", Namespace: "mcp/data/resources", Description: "资源数据轨道，用于上下文与静态内容传输"},
			{Name: "tools", Namespace: "mcp/data/tools", Description: "工具数据轨道，用于工具调用相关载荷"},
			{Name: "notifications", Namespace: "mcp/data/notifications", Description: "通知数据轨道，用于服务端主动推送"},
		},
	}, nil
}

type rttInput struct {
	SamplesMS []int `json:"samples_ms" jsonschema:"RTT 样本（毫秒），至少 1 个"`
}

type rttOutput struct {
	Count   int     `json:"count"`
	MinMS   int     `json:"min_ms"`
	MaxMS   int     `json:"max_ms"`
	AvgMS   float64 `json:"avg_ms"`
	Summary string  `json:"summary"`
}

func estimateRTT(ctx context.Context, _ *mcp.CallToolRequest, in rttInput) (*mcp.CallToolResult, rttOutput, error) {
	_ = ctx
	if len(in.SamplesMS) == 0 {
		return nil, rttOutput{}, fmt.Errorf("samples_ms 不能为空")
	}
	minV, maxV := in.SamplesMS[0], in.SamplesMS[0]
	sum := 0
	for _, v := range in.SamplesMS {
		if v < 0 {
			return nil, rttOutput{}, fmt.Errorf("RTT 样本不能为负数: %d", v)
		}
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
		sum += v
	}
	avg := float64(sum) / float64(len(in.SamplesMS))
	out := rttOutput{
		Count:   len(in.SamplesMS),
		MinMS:   minV,
		MaxMS:   maxV,
		AvgMS:   avg,
		Summary: fmt.Sprintf("基于 %d 个样本：min=%dms max=%dms avg=%.1fms", len(in.SamplesMS), minV, maxV, avg),
	}
	return nil, out, nil
}

func registerPrompts(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "debug_moqt_session",
		Description: "生成用于排查 MCP over MOQT 会话问题的调试提示词",
		Arguments: []*mcp.PromptArgument{
			{Name: "symptom", Description: "观察到的异常现象（如超时、断连、ACK 丢失）", Required: true},
			{Name: "addr", Description: "服务地址，例如 127.0.0.1:8080", Required: false},
		},
	}, debugMOQTSessionPrompt)

	server.AddPrompt(&mcp.Prompt{
		Name:        "configure_transport",
		Description: "生成配置 QUIC/MOQT 传输参数的助手提示词",
		Arguments: []*mcp.PromptArgument{
			{Name: "environment", Description: "部署环境：dev / test / prod", Required: true},
			{Name: "goal", Description: "优化目标：latency / throughput / reliability", Required: false},
		},
	}, configureTransportPrompt)
}

func debugMOQTSessionPrompt(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	symptom := req.Params.Arguments["symptom"]
	addr := req.Params.Arguments["addr"]
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	text := fmt.Sprintf(`你是 MCP over MOQT 传输层专家。请帮助排查会话问题。

目标地址: %s
异常现象: %s

请按以下步骤分析：
1. 确认 QUIC 握手与 ALPN（moq-00）是否成功
2. 检查 discovery（FETCH mcp/discovery/sessions）是否返回会话
3. 检查控制轨道 JSON-RPC 读写是否超时
4. 如使用数据轨道，确认 resources/tools/notifications 订阅是否建立
5. 给出可执行的复现命令与修复建议`, addr, symptom)

	return &mcp.GetPromptResult{
		Description: "MOQT 会话调试提示词",
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: text}},
		},
	}, nil
}

func configureTransportPrompt(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	env := req.Params.Arguments["environment"]
	goal := req.Params.Arguments["goal"]
	if goal == "" {
		goal = "reliability"
	}
	text := fmt.Sprintf(`请为 MCP over MOQT 传输层生成 %s 环境配置建议，优化目标为 %s。

需要覆盖：
- WithAddr / TLS（本地自签名或正式证书）
- QUIC：MaxIdleTimeout、KeepAlive、流窗口、0-RTT
- 可靠性：heartbeat、ACK 超时、重试退避
- 数据轨道：是否启用 resources/tools/notifications
- 监控：关键 metrics 与日志级别

输出一份可直接落地的 YAML 配置草案与简要说明。`, env, goal)

	return &mcp.GetPromptResult{
		Description: "传输层配置助手提示词",
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: text}},
		},
	}, nil
}

func registerResources(server *mcp.Server) {
	handler := resourceHandler()

	server.AddResource(&mcp.Resource{
		URI:         "moqt://docs/overview",
		Name:        "Service Overview",
		Description: "MCP over MOQT 服务概述与能力说明",
		MIMEType:    "text/markdown",
	}, handler)

	server.AddResource(&mcp.Resource{
		URI:         "moqt://config/example",
		Name:        "Example Config",
		Description: "示例传输层 YAML 配置",
		MIMEType:    "application/yaml",
	}, handler)

	server.AddResource(&mcp.Resource{
		URI:         "moqt://tracks/catalog",
		Name:        "Data Track Catalog",
		Description: "数据轨道目录（JSON）",
		MIMEType:    "application/json",
	}, handler)

	server.AddResource(&mcp.Resource{
		URI:         "moqt://install/quickstart",
		Name:        "Install Quickstart",
		Description: "友好安装与一键启动说明",
		MIMEType:    "text/markdown",
	}, handler)
}

func resourceHandler() mcp.ResourceHandler {
	catalog, _ := json.MarshalIndent(map[string]any{
		"tracks": []map[string]string{
			{"name": "resources", "namespace": "mcp/data/resources"},
			{"name": "tools", "namespace": "mcp/data/tools"},
			{"name": "notifications", "namespace": "mcp/data/notifications"},
		},
	}, "", "  ")

	contents := map[string]string{
		"moqt://docs/overview": `# MCP over MOQT Transport

高性能 MCP 传输层，基于 QUIC + MOQT。

## 能力
- Tools: health_check / get_transport_info / list_data_tracks / estimate_rtt
- Prompts: debug_moqt_session / configure_transport
- Resources: overview / example config / track catalog / install quickstart

## 协议
- ALPN: moq-00
- Draft: moqtransport draft-11
`,
		"moqt://config/example": `server:
  addr: "127.0.0.1:8080"
  alpn:
    - "moq-00"
client:
  addr: "127.0.0.1:8080"
  insecure_skip_verify: true
logging:
  level: info
`,
		"moqt://tracks/catalog": string(catalog),
		"moqt://install/quickstart": `# 友好安装

## 一键脚本
- Linux/macOS: curl -fsSL https://raw.githubusercontent.com/mcp-moqt/mcp-moqt-transport/main/scripts/install.sh | bash
- Windows: irm https://raw.githubusercontent.com/mcp-moqt/mcp-moqt-transport/main/scripts/install.ps1 | iex

## Go 安装
go install github.com/mcp-moqt/mcp-moqt-transport/cmd/mcp-moqt@latest

## Docker
docker compose up --build
`,
	}

	return func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		uri := req.Params.URI
		text, ok := contents[uri]
		if !ok {
			return nil, mcp.ResourceNotFoundError(uri)
		}
		mime := "text/plain"
		switch uri {
		case "moqt://docs/overview", "moqt://install/quickstart":
			mime = "text/markdown"
		case "moqt://config/example":
			mime = "application/yaml"
		case "moqt://tracks/catalog":
			mime = "application/json"
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      uri,
				MIMEType: mime,
				Text:     text,
			}},
		}, nil
	}
}
