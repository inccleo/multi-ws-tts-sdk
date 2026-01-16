# Multi-Context WebSocket TTS SDK

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Python](https://img.shields.io/badge/Python-3.8+-3776AB?style=flat&logo=python)](https://www.python.org/)
[![Java](https://img.shields.io/badge/Java-8+-007396?style=flat&logo=openjdk)](https://www.java.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> 🎯 **企业级多上下文 WebSocket TTS SDK**  
> 单连接管理多个独立的文本转语音流，支持最多 5 个并发上下文

## 🌟 核心特性

- ✅ **多上下文并发**：单个 WebSocket 连接支持最多 5 个并发 TTS 流
- ✅ **实时流式输出**：边接收边播放，低延迟
- ✅ **独立生命周期**：每个上下文独立管理，互不干扰
- ✅ **完整错误处理**：标准化错误码和回调机制
- ✅ **格式兼容**：自动支持 camelCase 和 snake_case 字段格式
- ✅ **生产就绪**：完整测试，可直接用于生产环境

## 📦 支持的语言

### [Go SDK](./go/) 

```bash
go get github.com/inccleo/multi-ws-tts-sdk/go/tts
```

**特点：**
- 🚀 高性能，原生支持并发
- 📝 类型安全，完整的类型定义
- 🔧 简洁的 API 设计
- ✅ 包含单元测试

[📖 Go SDK 文档](./go/README.md) | [查看示例](./go/examples/)

---

### [Python SDK](./py/) 

```bash
pip install git+https://github.com/inccleo/multi-ws-tts-sdk.git#subdirectory=py
```

**特点：**
- ⚡ 基于 asyncio，异步高效
- 🔗 链式调用，优雅的 API
- 🐍 Pythonic 设计
- ✅ 完整的类型提示

[📖 Python SDK 文档](./py/README.md) | [查看示例](./py/examples/)

---

### [Java SDK](./java/) 

```xml
<!-- Maven -->
<dependency>
    <groupId>com.inccleo</groupId>
    <artifactId>multi-ws-tts-sdk</artifactId>
    <version>1.0.0</version>
</dependency>
```

**特点：**
- ☕ Java 8+ 兼容
- 🔒 线程安全设计
- 🎯 函数式回调 (Lambda)
- 🔗 链式调用 API

[📖 Java SDK 文档](./java/README.md) | [查看示例](./java/examples/)

---

## 🚀 快速开始

### Go 示例

```go
package main

import (
    "github.com/inccleo/multi-ws-tts-sdk/go/tts"
)

func main() {
    client := tts.NewTTSClient(baseURL, apiKey, voiceID)
    client.Connect(map[string]interface{}{
        "model_id": "flash_v2_5",
        "format":   "pcm_16000",
    })
    
    ctx := client.CreateContext("ctx_001")
    ctx.OnAudio(func(audio []byte, isFinal bool) {
        // 处理音频数据
    })
    
    ctx.SendText("你好，世界", true)
}
```

### Python 示例

```python
import asyncio
from multi_ws_tts_sdk import TTSClient

async def main():
    client = TTSClient(base_url, api_key, voice_id)
    await client.connect({"model_id": "flash_v2_5", "format": "pcm_16000"})
    
    ctx = client.create_context("ctx_001")
    ctx.on_audio(lambda audio, is_final: print(f"收到 {len(audio)} 字节"))
    
    await ctx.send_text("你好，世界", flush=True)
    await asyncio.sleep(3)
    await ctx.close()

asyncio.run(main())
```

### Java 示例

```java
import com.inccleo.tts.TTSClient;
import com.inccleo.tts.TTSContext;

public class QuickStart {
    public static void main(String[] args) throws Exception {
        TTSClient client = new TTSClient(baseUrl, apiKey, voiceID);
        client.connect(Map.of(
            "model_id", "flash_v2_5",
            "format", "pcm_16000"
        ));
        
        TTSContext ctx = client.createContext("ctx_001");
        ctx.onAudio((audio, isFinal) -> {
            // 处理音频数据
        });
        
        ctx.sendText("你好，世界", true);
        ctx.endInput();
        Thread.sleep(3000);
        ctx.close();
        client.disconnect();
    }
}
```

## 📊 架构设计

```
                   单个 WebSocket 连接
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
    Context 1         Context 2         Context 3
        │                 │                 │
    "你好世界"         "How are you"      "Bonjour"
        │                 │                 │
     音频流 1           音频流 2           音频流 3
```

### 核心概念

- **TTSClient**：管理 WebSocket 连接和多个 Context
- **TTSContext**：独立的 TTS 流，支持发送文本、接收音频、错误处理
- **生命周期**：创建 → 发送文本 → 接收音频 → 关闭
- **并发限制**：单连接最多 5 个活跃 Context

## 🔧 配置选项

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `model_id` | 模型 ID | `flash_v2_5` |
| `format` | 音频格式 | `pcm_16000` |
| `language_code` | 语言代码 | `zh` |
| `priority` | 优先级 | `dedicated_concurrency` |

## 📖 API 文档

完整的 API 规范请参考：[multi-context-websocket-sdk-guide.md](./multi-context-websocket-sdk-guide.md)

## 🧪 测试

### Go SDK

```bash
cd go
go test ./...
```

### Python SDK

```bash
cd py
pip install -e .
python examples/simple_example.py
```

### Java SDK

```bash
cd java
mvn clean compile
javac -cp "..." examples/SimpleExample.java
java -cp "..." SimpleExample
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

---

## 📞 支持

如有问题，请提交 [Issue](https://github.com/inccleo/multi-ws-tts-sdk/issues)

---

<div align="center">

**[Go SDK](./go/)** · **[Python SDK](./py/)** · **[Java SDK](./java/)**

Made with ❤️ for developers

</div>
