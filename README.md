# Multi-Context WebSocket TTS SDK (Go)

基于 WebSocket 的多上下文 TTS（文本转语音）Go SDK。

## 安装

```bash
go get github.com/inccleo/multi-ws-tts-sdk@latest
```

## 目录结构

```
go/
├── tts/              # SDK 核心代码
│   ├── client.go     # WebSocket 客户端
│   ├── context.go    # TTS 上下文
│   ├── client_test.go
│   └── context_test.go
├── examples/         # 示例代码
│   ├── simple/       # 单上下文示例
│   └── multi_context/ # 多上下文并发示例
├── go.mod
└── go.sum
```

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 运行示例

```bash
# 设置环境变量
export TTS_BASE_URL="wss://your-domain.com"
export TTS_API_KEY="your_api_key"
export TTS_VOICE_ID="your_voice_id"

# 运行单上下文示例
go run examples/simple/main.go

# 运行多上下文示例
go run examples/multi_context/main.go
```

### 3. 编译示例

```bash
# 编译 simple 示例
go build -o bin/simple ./examples/simple

# 编译 multi_context 示例
go build -o bin/multi_context ./examples/multi_context

# 运行编译后的程序
./bin/simple
./bin/multi_context
```

## 基本使用

```go
package main

import (
    "fmt"
    "time"
    "github.com/inccleo/multi-ws-tts-sdk/tts"
)

func main() {
    // 创建客户端
    client := tts.NewTTSClient(
        "wss://your-domain.com",
        "your_api_key",
        "your_voice_id",
    )

    // 连接
    params := map[string]string{
        "model_id": "flash_v2_5",
        "format": "pcm_16000",
        "language_code": "zh",
    }
    
    if err := client.Connect(params); err != nil {
        panic(err)
    }
    defer client.Disconnect()

    // 创建上下文
    ctx, _ := client.CreateContext("ctx_001")
    
    // 设置回调
    ctx.OnAudio = func(audio []byte, isFinal bool) {
        fmt.Printf("收到音频: %d 字节\n", len(audio))
    }
    
    ctx.OnError = func(code, msg string) {
        fmt.Printf("错误: %s - %s\n", code, msg)
    }
    
    ctx.OnComplete = func() {
        fmt.Println("完成")
    }
    
    // 发送文本
    ctx.SendText("你好，世界", true)
    
    time.Sleep(5 * time.Second)
    ctx.Close()
}
```

## 调试模式

设置 `TTS_DEBUG=1` 可查看详细的消息日志：

```bash
export TTS_DEBUG=1
go run examples/simple/main.go
```

## 运行测试

```bash
go test ./tts/...
```

## API 兼容性

SDK 支持服务器返回的 camelCase 和 snake_case 两种字段格式：
- `contextId` / `context_id`
- `isFinal` / `is_final`

---

## 📦 发布到仓库

### 快速发布（推荐）

使用提供的发布脚本：

```bash
./publish.sh
```

脚本会自动：
1. 更新模块路径和导入路径
2. 运行测试
3. 初始化 Git 仓库
4. 推送到远程仓库
5. 创建版本标签

### 用户安装

发布后，用户可以通过以下方式安装：

```bash
# 安装最新版本
go get github.com/inccleo/multi-ws-tts-sdk@latest

# 安装指定版本
go get github.com/inccleo/multi-ws-tts-sdk@v1.0.0
```

## 📚 文档

完整的 API 文档会自动发布到：
```
https://pkg.go.dev/github.com/inccleo/multi-ws-tts-sdk
```
