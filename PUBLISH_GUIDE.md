# Go SDK 发布指南

## 📦 准备发布

### 1. 创建 GitHub/GitLab 仓库

首先在 GitHub 或 GitLab 创建一个新仓库，例如：
- GitHub: `https://github.com/your-username/multi-ws-tts-sdk`
- GitLab: `https://gitlab.com/your-username/multi-ws-tts-sdk`

### 2. 修改 go.mod 模块路径

将 `go.mod` 中的模块路径改为你的实际仓库地址：

```go
// 修改前
module github.com/yourcompany/multi-ws-tts-sdk

// 修改后（GitHub）
module github.com/your-username/multi-ws-tts-sdk

// 或（GitLab）
module gitlab.com/your-username/multi-ws-tts-sdk
```

### 3. 更新示例代码的导入路径

修改 `examples/simple/main.go` 和 `examples/multi_context/main.go`：

```go
// 修改前
import "github.com/yourcompany/multi-ws-tts-sdk/tts"

// 修改后
import "github.com/your-username/multi-ws-tts-sdk/tts"
```

## 🚀 发布到 Git 仓库

### 初始化并推送

```bash
cd /Users/leo/Desktop/202601/multi-ws-sdk/go

# 初始化 Git 仓库
git init

# 创建 .gitignore
cat > .gitignore << 'EOF'
# 编译产物
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# 测试产物
*.test
*.out

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# macOS
.DS_Store
EOF

# 添加所有文件
git add .

# 提交
git commit -m "feat: 初始化 Multi-Context WebSocket TTS Go SDK

- 实现 WebSocket 客户端
- 支持多上下文并发
- 提供 simple 和 multi_context 示例
- 兼容 camelCase 和 snake_case 字段格式"

# 添加远程仓库（替换为你的实际地址）
git remote add origin https://github.com/your-username/multi-ws-tts-sdk.git

# 推送到 main 分支
git branch -M main
git push -u origin main
```

## 🏷️ 发布版本

### 打标签发布

```bash
# 打版本标签（Go 模块使用语义化版本）
git tag v1.0.0

# 推送标签
git push origin v1.0.0
```

### 版本号规范

遵循语义化版本 (SemVer)：
- `v1.0.0` - 主版本.次版本.修订号
- `v1.0.1` - 修复 bug
- `v1.1.0` - 新增功能（向后兼容）
- `v2.0.0` - 破坏性更新

## 📥 用户如何使用

### 方式一：直接使用（推荐）

用户创建项目：

```bash
mkdir my-tts-project
cd my-tts-project
go mod init my-tts-project

# 安装你的 SDK
go get github.com/your-username/multi-ws-tts-sdk@latest
```

示例代码：

```go
package main

import (
    "fmt"
    "time"
    "github.com/your-username/multi-ws-tts-sdk/tts"
)

func main() {
    client := tts.NewTTSClient(
        "wss://your-domain.com",
        "your_api_key",
        "your_voice_id",
    )
    
    params := map[string]string{
        "model_id": "flash_v2_5",
        "format": "pcm_16000",
        "language_code": "zh",
    }
    
    if err := client.Connect(params); err != nil {
        panic(err)
    }
    defer client.Disconnect()
    
    ctx, _ := client.CreateContext("ctx_001")
    ctx.OnAudio = func(audio []byte, isFinal bool) {
        fmt.Printf("收到音频: %d 字节\n", len(audio))
    }
    
    ctx.SendText("你好，世界", true)
    time.Sleep(3 * time.Second)
}
```

### 方式二：克隆仓库使用示例

```bash
# 克隆仓库
git clone https://github.com/your-username/multi-ws-tts-sdk.git
cd multi-ws-tts-sdk/go

# 安装依赖
go mod tidy

# 设置环境变量
export TTS_BASE_URL="wss://your-domain.com"
export TTS_API_KEY="your_api_key"
export TTS_VOICE_ID="your_voice_id"

# 运行示例
go run examples/simple/main.go
```

## 📚 发布到 pkg.go.dev

当你推送标签后，pkg.go.dev 会自动抓取你的模块（可能需要几分钟）。

用户可以访问：
```
https://pkg.go.dev/github.com/your-username/multi-ws-tts-sdk
```

查看完整的 API 文档。

## 🔄 更新版本

当你修复 bug 或添加新功能后：

```bash
# 提交更改
git add .
git commit -m "fix: 修复音频缓冲区问题"
git push

# 发布新版本
git tag v1.0.1
git push origin v1.0.1
```

用户更新：

```bash
go get -u github.com/your-username/multi-ws-tts-sdk@latest
```

## 📝 最佳实践

### 1. 添加 LICENSE 文件

```bash
# 选择许可证，例如 MIT License
cat > LICENSE << 'EOF'
MIT License

Copyright (c) 2026 Your Name

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction...
EOF
```

### 2. 完善 README.md

确保包含：
- ✅ 清晰的项目描述
- ✅ 安装说明
- ✅ 快速开始示例
- ✅ API 文档链接
- ✅ 贡献指南

### 3. 添加 GitHub Actions CI

创建 `.github/workflows/test.yml`：

```yaml
name: Test

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - run: go test -v ./...
```

## 🎯 完整发布检查清单

- [ ] 修改 `go.mod` 模块路径
- [ ] 更新示例代码导入路径
- [ ] 创建 `.gitignore`
- [ ] 添加 `LICENSE` 文件
- [ ] 完善 `README.md`
- [ ] 运行 `go test ./...` 确保测试通过
- [ ] 运行 `go mod tidy` 清理依赖
- [ ] 初始化 Git 并提交
- [ ] 推送到远程仓库
- [ ] 打标签发布版本
- [ ] 在 pkg.go.dev 验证文档
- [ ] 编写使用示例

## 🌐 私有仓库

如果是私有仓库，用户需要配置：

```bash
# 配置 Git 凭据
git config --global url."https://username:token@github.com/".insteadOf "https://github.com/"

# 或使用 SSH
go env -w GOPRIVATE=github.com/your-username/*
```

## 📞 支持

在仓库的 README.md 中提供：
- Issues 链接
- 讨论区链接
- 联系方式
