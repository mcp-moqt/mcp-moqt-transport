package mcpservice_test

import (
	"context"
	"testing"

	"github.com/mcp-moqt/mcp-moqt-transport/pkg/mcpservice"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestRegisterCapabilities_ExposesToolsPromptsResources(t *testing.T) {
	ctx := context.Background()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    mcpservice.ServiceName,
		Version: mcpservice.ServiceVersion,
	}, nil)
	mcpservice.RegisterCapabilities(server)

	t1, t2 := mcp.NewInMemoryTransports()
	_, err := server.Connect(ctx, t1, nil)
	require.NoError(t, err)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v1"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	require.NoError(t, err)
	defer session.Close()

	tools, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(tools.Tools), 4)

	prompts, err := session.ListPrompts(ctx, nil)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(prompts.Prompts), 2)

	resources, err := session.ListResources(ctx, nil)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(resources.Resources), 4)

	health, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "health_check"})
	require.NoError(t, err)
	require.False(t, health.IsError)

	prompt, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name:      "debug_moqt_session",
		Arguments: map[string]string{"symptom": "timeout"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, prompt.Messages)

	res, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: "moqt://docs/overview"})
	require.NoError(t, err)
	require.NotEmpty(t, res.Contents)
	require.Contains(t, res.Contents[0].Text, "MCP over MOQT")
}
