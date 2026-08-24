# Contributing

感谢关注 MCP over MOQT Transport。欢迎通过 Issue 与 Pull Request 参与改进。

## 提交 Issue

- 报告 bug（请附上复现步骤、Go 版本、操作系统）
- 提出新功能建议
- 讨论现有功能的改进

## 提交 Pull Request

1. Fork 本仓库
2. 克隆到本地：`git clone https://github.com/<your-username>/mcp-moqt-transport.git`
3. 创建分支：`git checkout -b feature/your-feature-name`
4. 实现功能并补充测试
5. 运行检查：

```bash
go fmt ./...
go vet ./...
go test ./test/...
```

6. 提交：`git commit -m "Describe your change"`
7. 推送并创建 Pull Request，简要说明动机与改动

## 代码规范

- 遵循 Go 标准风格，使用 `go fmt`
- 新功能尽量附带单元或集成测试
- 确保相关测试通过
- 公共 API 变更请同步更新 `docs/api.md` 与 `CHANGELOG.md`
- 版本号以 `pkg/version/version.go` 为准，勿在多处硬编码

## 本地常用命令

```bash
make doctor    # 环境与 capabilities 冒烟
make test      # 运行测试
make stdio     # stdio 模式服务
make server    # QUIC 模式服务
```
