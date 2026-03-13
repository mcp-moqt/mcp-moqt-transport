package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 包含 MCP over MOQT 传输的配置选项
type Config struct {
	// Addr 是服务器监听地址或客户端连接地址
	Addr string `json:"addr" yaml:"addr"`
	// ALPN 是 TLS ALPN 协议
	ALPN []string `json:"alpn" yaml:"alpn"`
	// EnableDatagrams 是否启用 QUIC 数据报
	EnableDatagrams bool `json:"enable_datagrams" yaml:"enable_datagrams"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Addr:            "localhost:0",
		ALPN:            []string{"moq-00"},
		EnableDatagrams: true,
	}
}

// LoadFromEnv 从环境变量加载配置
func LoadFromEnv() *Config {
	config := DefaultConfig()

	// 从环境变量加载地址
	if addr := os.Getenv("MCP_MOQT_ADDR"); addr != "" {
		config.Addr = addr
	}

	// 从环境变量加载 ALPN
	if alpn := os.Getenv("MCP_MOQT_ALPN"); alpn != "" {
		config.ALPN = strings.Split(alpn, ",")
	}

	// 从环境变量加载是否启用数据报
	if enableDatagrams := os.Getenv("MCP_MOQT_ENABLE_DATAGRAMS"); enableDatagrams != "" {
		config.EnableDatagrams = enableDatagrams == "true"
	}

	return config
}

// LoadFromFile 从文件加载配置，支持 YAML 和 JSON 格式
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()

	// 根据文件扩展名判断格式
	switch {
	case strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml"):
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case strings.HasSuffix(path, ".json"):
		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		// 默认尝试 YAML
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
	}

	return config, nil
}

// SaveToFile 保存配置到文件，支持 YAML 和 JSON 格式
func (c *Config) SaveToFile(path string) error {
	var data []byte
	var err error

	switch {
	case strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml"):
		data, err = yaml.Marshal(c)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
	case strings.HasSuffix(path, ".json"):
		data, err = json.MarshalIndent(c, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
	default:
		data, err = yaml.Marshal(c)
		if err != nil {
			return fmt.Errorf("failed to marshal: %w", err)
		}
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if c.Addr == "" {
		return fmt.Errorf("addr must not be empty")
	}

	if len(c.ALPN) == 0 {
		return fmt.Errorf("alpn must not be empty")
	}

	return nil
}
