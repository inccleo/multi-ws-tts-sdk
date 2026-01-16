# Multi-Context WebSocket TTS SDK 开发指南

> 本文档为开发者提供基于 Multi-Context WebSocket 接口实现 SDK 的完整规范和示例代码（Java、Python、Go）

---

## 📋 目录

1. [接口概述](#接口概述)
2. [核心概念](#核心概念)
3. [连接规范](#连接规范)
4. [消息协议](#消息协议)
5. [SDK 架构设计](#sdk-架构设计)
6. [Java SDK 实现](#java-sdk-实现)
7. [Python SDK 实现](#python-sdk-实现)
8. [Go SDK 实现](#go-sdk-实现)
9. [错误处理](#错误处理)
10. [最佳实践](#最佳实践)
11. [测试建议](#测试建议)

---

## 接口概述

### 端点信息

```
wss://<your-domain>/enterprise/v1/tts/{voice_id}/websocket/multi
```

### 核心特性

- ✅ **多 Context 并发**：单连接支持最多 5 个并发 context
- ✅ **实时语音生成**：流式返回音频数据
- ✅ **用户打断支持**：可随时关闭 context
- ✅ **配额实时检查**：每条消息实时验证配额
- ✅ **错误标准化**：统一的错误码和错误消息

### 使用场景

- AI 电话外呼
- 实时语音 Agent
- 多轮对话系统
- 语音交互应用

---

## 核心概念

### Context（上下文）

每个 **context** 代表一个独立的语音生成流：

```
┌─────────────────────────────────────────────┐
│         WebSocket Connection                 │
├─────────────────────────────────────────────┤
│  Context 1: "你好，我是AI助手"              │
│  Context 2: "请问有什么可以帮您？"          │
│  Context 3: "好的，我明白了"                │
│  Context 4: "感谢您的咨询"                  │
│  Context 5: "再见"                          │
└─────────────────────────────────────────────┘
```

### Context 生命周期

```
创建 → 发送文本 → 接收音频 → 关闭
  ↑                              │
  └──────────── 可复用 ───────────┘
```

### 并发限制

- 单个 WebSocket 连接：**最多 5 个活跃 context**
- 超限后：返回 `MAX_CONTEXT_LIMIT_EXCEEDED` 错误
- 关闭 context 后：可创建新的 context

---

## 连接规范

### 1. WebSocket URL 构建

```
wss://<domain>/enterprise/v1/tts/{voice_id}/websocket/multi?<query_params>
```

#### 必需参数

| 参数 | 值 | 说明 |
|------|-----|------|
| `priority` | `dedicated_concurrency` | 必填，用于权限验证 |

#### 可选参数

| 参数 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `model_id` | string | 模型ID（支持简化名） | `flash_v2_5` |
| `format` | string | 音频格式（映射为 `output_format`） | `pcm_16000` |
| `language_code` | string | 语言代码 | `zh` |
| `enable_logging` | boolean | 启用上游日志 | `true` |

**参数映射关系**：

| 客户端参数 | 上游参数 | 说明 |
|-----------|----------|------|
| `format` | `output_format` | 音频格式 |
| `idleTimeout` | `inactivity_timeout` | 空闲超时 |
| `timestamps` | `sync_alignment` | 时间戳对齐 |
| `directStreaming` | `auto_mode` | 自动模式 |

### 2. Headers

```http
api-key: <your-api-key>
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Version: 13
```

### 3. 连接建立

```
Client                           Server
  │                                 │
  ├─────── WebSocket Upgrade ──────>│
  │                                 │
  │<────── 101 Switching Protocols ─┤
  │                                 │
  │        (连接建立成功)            │
  │                                 │
```

---

## 消息协议

### 客户端 → 服务器

#### 消息类型

所有消息均为 **JSON 格式**的 TextMessage。

#### 1. 初始化 Context

```json
{
  "context_id": "ctx_001",
  "text": "你好"
}
```

#### 2. 继续发送文本

```json
{
  "context_id": "ctx_001",
  "text": "，我是AI助手"
}
```

#### 3. 强制生成（Flush）

```json
{
  "context_id": "ctx_001",
  "text": "。",
  "flush": true
}
```

#### 4. 结束输入（EOS）

```json
{
  "context_id": "ctx_001",
  "text": ""
}
```

#### 5. 关闭 Context

```json
{
  "context_id": "ctx_001",
  "close_context": true
}
```

### 服务器 → 客户端

#### 1. 音频数据

```json
{
  "context_id": "ctx_001",
  "audio": "UklGRiQAAABXQVZF...",
  "is_final": false
}
```

- `audio`: Base64 编码的音频数据
- `is_final`: 是否为该 context 的最后一个音频块

#### 2. 最终音频

```json
{
  "context_id": "ctx_001",
  "audio": "...",
  "is_final": true
}
```

#### 3. 错误消息

```json
{
  "error": "ERROR_CODE",
  "message": "错误描述",
  "context_id": "ctx_001"
}
```

---

## SDK 架构设计

### 核心组件

```
┌─────────────────────────────────────────┐
│              TTSClient                   │
│  - connect()                             │
│  - disconnect()                          │
│  - createContext(contextId)              │
├─────────────────────────────────────────┤
│            TTSContext                    │
│  - sendText(text, flush)                 │
│  - close()                               │
│  - onAudio(callback)                     │
│  - onError(callback)                     │
│  - onComplete(callback)                  │
├─────────────────────────────────────────┤
│          WebSocketManager                │
│  - sendMessage(message)                  │
│  - receiveMessage()                      │
│  - handleReconnect()                     │
├─────────────────────────────────────────┤
│          AudioBufferManager              │
│  - bufferAudio(contextId, audio)         │
│  - getAudioData(contextId)               │
│  - clearBuffer(contextId)                │
└─────────────────────────────────────────┘
```

### 设计原则

1. **异步优先**：所有 I/O 操作应为异步
2. **事件驱动**：通过回调/事件通知用户
3. **线程安全**：支持多线程环境
4. **资源管理**：自动管理连接和 context 生命周期
5. **错误容错**：优雅处理网络异常和服务错误

---

## Java SDK 实现

### Maven 依赖

```xml
<dependencies>
    <!-- WebSocket -->
    <dependency>
        <groupId>org.java-websocket</groupId>
        <artifactId>Java-WebSocket</artifactId>
        <version>1.5.3</version>
    </dependency>
    
    <!-- JSON -->
    <dependency>
        <groupId>com.google.code.gson</groupId>
        <artifactId>gson</artifactId>
        <version>2.10.1</version>
    </dependency>
</dependencies>
```

### 核心类实现

#### 1. TTSClient.java

```java
package com.yourcompany.tts;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import org.java_websocket.client.WebSocketClient;
import org.java_websocket.handshake.ServerHandshake;

import java.net.URI;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;

public class TTSClient {
    private final String apiKey;
    private final String voiceId;
    private final String baseUrl;
    private WebSocketClient webSocketClient;
    private final Map<String, TTSContext> contexts = new ConcurrentHashMap<>();
    private final Gson gson = new Gson();
    
    public TTSClient(String baseUrl, String apiKey, String voiceId) {
        this.baseUrl = baseUrl;
        this.apiKey = apiKey;
        this.voiceId = voiceId;
    }
    
    /**
     * 连接到 WebSocket 服务器
     */
    public CompletableFuture<Void> connect(Map<String, String> queryParams) {
        CompletableFuture<Void> future = new CompletableFuture<>();
        
        try {
            // 构建 URL
            StringBuilder url = new StringBuilder(baseUrl)
                .append("/enterprise/v1/tts/")
                .append(voiceId)
                .append("/websocket/multi?priority=dedicated_concurrency");
            
            if (queryParams != null) {
                queryParams.forEach((key, value) -> 
                    url.append("&").append(key).append("=").append(value)
                );
            }
            
            // 创建 WebSocket 客户端
            Map<String, String> headers = new HashMap<>();
            headers.put("api-key", apiKey);
            
            webSocketClient = new WebSocketClient(new URI(url.toString()), headers) {
                @Override
                public void onOpen(ServerHandshake handshake) {
                    System.out.println("WebSocket connected");
                    future.complete(null);
                }
                
                @Override
                public void onMessage(String message) {
                    handleMessage(message);
                }
                
                @Override
                public void onClose(int code, String reason, boolean remote) {
                    System.out.println("WebSocket closed: " + reason);
                    contexts.values().forEach(ctx -> ctx.handleClose());
                }
                
                @Override
                public void onError(Exception ex) {
                    System.err.println("WebSocket error: " + ex.getMessage());
                    future.completeExceptionally(ex);
                }
            };
            
            webSocketClient.connect();
            
        } catch (Exception e) {
            future.completeExceptionally(e);
        }
        
        return future;
    }
    
    /**
     * 创建新的 Context
     */
    public TTSContext createContext(String contextId) {
        if (contexts.size() >= 5) {
            throw new IllegalStateException("Maximum 5 contexts allowed per connection");
        }
        
        TTSContext context = new TTSContext(contextId, this);
        contexts.put(contextId, context);
        return context;
    }
    
    /**
     * 发送消息到服务器
     */
    void sendMessage(JsonObject message) {
        if (webSocketClient != null && webSocketClient.isOpen()) {
            webSocketClient.send(gson.toJson(message));
        } else {
            throw new IllegalStateException("WebSocket is not connected");
        }
    }
    
    /**
     * 处理收到的消息
     */
    private void handleMessage(String message) {
        try {
            JsonObject json = gson.fromJson(message, JsonObject.class);
            
            // 错误消息
            if (json.has("error")) {
                String error = json.get("error").getAsString();
                String errorMessage = json.get("message").getAsString();
                String contextId = json.has("context_id") ? 
                    json.get("context_id").getAsString() : null;
                
                if (contextId != null && contexts.containsKey(contextId)) {
                    contexts.get(contextId).handleError(error, errorMessage);
                } else {
                    System.err.println("Error: " + error + " - " + errorMessage);
                }
                return;
            }
            
            // 音频数据
            if (json.has("context_id")) {
                String contextId = json.get("context_id").getAsString();
                TTSContext context = contexts.get(contextId);
                
                if (context != null) {
                    if (json.has("audio")) {
                        String audioData = json.get("audio").getAsString();
                        boolean isFinal = json.has("is_final") && 
                            json.get("is_final").getAsBoolean();
                        context.handleAudio(audioData, isFinal);
                    }
                }
            }
            
        } catch (Exception e) {
            System.err.println("Failed to parse message: " + e.getMessage());
        }
    }
    
    /**
     * 移除 Context
     */
    void removeContext(String contextId) {
        contexts.remove(contextId);
    }
    
    /**
     * 断开连接
     */
    public void disconnect() {
        if (webSocketClient != null) {
            webSocketClient.close();
        }
        contexts.clear();
    }
}
```

#### 2. TTSContext.java

```java
package com.yourcompany.tts;

import com.google.gson.JsonObject;

import java.util.ArrayList;
import java.util.Base64;
import java.util.List;
import java.util.function.BiConsumer;
import java.util.function.Consumer;

public class TTSContext {
    private final String contextId;
    private final TTSClient client;
    private final List<byte[]> audioBuffer = new ArrayList<>();
    
    // 回调函数
    private BiConsumer<byte[], Boolean> onAudioCallback;
    private BiConsumer<String, String> onErrorCallback;
    private Runnable onCompleteCallback;
    
    TTSContext(String contextId, TTSClient client) {
        this.contextId = contextId;
        this.client = client;
    }
    
    /**
     * 发送文本
     */
    public void sendText(String text, boolean flush) {
        JsonObject message = new JsonObject();
        message.addProperty("context_id", contextId);
        message.addProperty("text", text);
        if (flush) {
            message.addProperty("flush", true);
        }
        client.sendMessage(message);
    }
    
    /**
     * 发送文本（不立即 flush）
     */
    public void sendText(String text) {
        sendText(text, false);
    }
    
    /**
     * 结束输入（EOS）
     */
    public void endInput() {
        JsonObject message = new JsonObject();
        message.addProperty("context_id", contextId);
        message.addProperty("text", "");
        client.sendMessage(message);
    }
    
    /**
     * 关闭 Context
     */
    public void close() {
        JsonObject message = new JsonObject();
        message.addProperty("context_id", contextId);
        message.addProperty("close_context", true);
        client.sendMessage(message);
        client.removeContext(contextId);
    }
    
    /**
     * 设置音频回调
     */
    public void onAudio(BiConsumer<byte[], Boolean> callback) {
        this.onAudioCallback = callback;
    }
    
    /**
     * 设置错误回调
     */
    public void onError(BiConsumer<String, String> callback) {
        this.onErrorCallback = callback;
    }
    
    /**
     * 设置完成回调
     */
    public void onComplete(Runnable callback) {
        this.onCompleteCallback = callback;
    }
    
    /**
     * 处理音频数据
     */
    void handleAudio(String audioBase64, boolean isFinal) {
        byte[] audioData = Base64.getDecoder().decode(audioBase64);
        audioBuffer.add(audioData);
        
        if (onAudioCallback != null) {
            onAudioCallback.accept(audioData, isFinal);
        }
        
        if (isFinal && onCompleteCallback != null) {
            onCompleteCallback.run();
        }
    }
    
    /**
     * 处理错误
     */
    void handleError(String errorCode, String errorMessage) {
        if (onErrorCallback != null) {
            onErrorCallback.accept(errorCode, errorMessage);
        }
    }
    
    /**
     * 处理连接关闭
     */
    void handleClose() {
        if (onCompleteCallback != null) {
            onCompleteCallback.run();
        }
    }
    
    /**
     * 获取所有音频数据
     */
    public byte[] getAllAudio() {
        int totalLength = audioBuffer.stream().mapToInt(arr -> arr.length).sum();
        byte[] result = new byte[totalLength];
        int offset = 0;
        for (byte[] chunk : audioBuffer) {
            System.arraycopy(chunk, 0, result, offset, chunk.length);
            offset += chunk.length;
        }
        return result;
    }
}
```

#### 3. 使用示例

```java
import java.util.HashMap;
import java.util.Map;

public class Example {
    public static void main(String[] args) throws Exception {
        // 1. 创建客户端
        TTSClient client = new TTSClient(
            "wss://your-domain.com",
            "your-api-key",
            "your-voice-id"
        );
        
        // 2. 连接
        Map<String, String> params = new HashMap<>();
        params.put("model_id", "flash_v2_5");
        params.put("format", "pcm_16000");
        
        client.connect(params).get(); // 等待连接完成
        
        // 3. 创建 Context
        TTSContext context = client.createContext("ctx_001");
        
        // 4. 设置回调
        context.onAudio((audioData, isFinal) -> {
            System.out.println("Received audio: " + audioData.length + " bytes");
            if (isFinal) {
                System.out.println("Audio generation complete");
            }
        });
        
        context.onError((errorCode, errorMessage) -> {
            System.err.println("Error: " + errorCode + " - " + errorMessage);
        });
        
        context.onComplete(() -> {
            System.out.println("Context completed");
        });
        
        // 5. 发送文本
        context.sendText("你好，");
        context.sendText("我是AI助手。", true);
        
        // 6. 等待处理完成
        Thread.sleep(5000);
        
        // 7. 关闭
        context.close();
        client.disconnect();
    }
}
```

---

## Python SDK 实现

### 安装依赖

```bash
pip install websockets asyncio
```

### 核心类实现

#### 1. tts_client.py

```python
import asyncio
import json
import base64
from typing import Dict, Optional, Callable, List
import websockets
from websockets.client import WebSocketClientProtocol


class TTSContext:
    """TTS Context 上下文管理"""
    
    def __init__(self, context_id: str, websocket: WebSocketClientProtocol):
        self.context_id = context_id
        self.websocket = websocket
        self.audio_buffer: List[bytes] = []
        
        # 回调函数
        self.on_audio_callback: Optional[Callable[[bytes, bool], None]] = None
        self.on_error_callback: Optional[Callable[[str, str], None]] = None
        self.on_complete_callback: Optional[Callable[[], None]] = None
    
    async def send_text(self, text: str, flush: bool = False):
        """发送文本"""
        message = {
            "context_id": self.context_id,
            "text": text
        }
        if flush:
            message["flush"] = True
        
        await self.websocket.send(json.dumps(message))
    
    async def end_input(self):
        """结束输入（EOS）"""
        message = {
            "context_id": self.context_id,
            "text": ""
        }
        await self.websocket.send(json.dumps(message))
    
    async def close(self):
        """关闭 Context"""
        message = {
            "context_id": self.context_id,
            "close_context": True
        }
        await self.websocket.send(json.dumps(message))
    
    def on_audio(self, callback: Callable[[bytes, bool], None]):
        """设置音频回调"""
        self.on_audio_callback = callback
        return self
    
    def on_error(self, callback: Callable[[str, str], None]):
        """设置错误回调"""
        self.on_error_callback = callback
        return self
    
    def on_complete(self, callback: Callable[[], None]):
        """设置完成回调"""
        self.on_complete_callback = callback
        return self
    
    def handle_audio(self, audio_base64: str, is_final: bool):
        """处理音频数据"""
        audio_data = base64.b64decode(audio_base64)
        self.audio_buffer.append(audio_data)
        
        if self.on_audio_callback:
            self.on_audio_callback(audio_data, is_final)
        
        if is_final and self.on_complete_callback:
            self.on_complete_callback()
    
    def handle_error(self, error_code: str, error_message: str):
        """处理错误"""
        if self.on_error_callback:
            self.on_error_callback(error_code, error_message)
    
    def get_all_audio(self) -> bytes:
        """获取所有音频数据"""
        return b''.join(self.audio_buffer)


class TTSClient:
    """Multi-Context WebSocket TTS 客户端"""
    
    def __init__(self, base_url: str, api_key: str, voice_id: str):
        self.base_url = base_url
        self.api_key = api_key
        self.voice_id = voice_id
        self.websocket: Optional[WebSocketClientProtocol] = None
        self.contexts: Dict[str, TTSContext] = {}
        self._receive_task: Optional[asyncio.Task] = None
    
    async def connect(self, query_params: Optional[Dict[str, str]] = None):
        """连接到 WebSocket 服务器"""
        # 构建 URL
        url = f"{self.base_url}/enterprise/v1/tts/{self.voice_id}/websocket/multi"
        url += "?priority=dedicated_concurrency"
        
        if query_params:
            for key, value in query_params.items():
                url += f"&{key}={value}"
        
        # 连接
        self.websocket = await websockets.connect(
            url,
            additional_headers={
                "api-key": self.api_key
            }
        )
        
        print(f"✅ Connected to {url}")
        
        # 启动消息接收任务
        self._receive_task = asyncio.create_task(self._receive_messages())
    
    def create_context(self, context_id: str) -> TTSContext:
        """创建新的 Context"""
        if len(self.contexts) >= 5:
            raise ValueError("Maximum 5 contexts allowed per connection")
        
        if not self.websocket:
            raise RuntimeError("WebSocket is not connected")
        
        context = TTSContext(context_id, self.websocket)
        self.contexts[context_id] = context
        return context
    
    async def _receive_messages(self):
        """接收消息的后台任务"""
        try:
            async for message in self.websocket:
                await self._handle_message(message)
        except websockets.exceptions.ConnectionClosed:
            print("WebSocket connection closed")
        except Exception as e:
            print(f"Error receiving messages: {e}")
    
    async def _handle_message(self, message: str):
        """处理收到的消息"""
        try:
            data = json.loads(message)
            
            # 错误消息
            if "error" in data:
                error_code = data["error"]
                error_message = data.get("message", "Unknown error")
                context_id = data.get("context_id")
                
                if context_id and context_id in self.contexts:
                    self.contexts[context_id].handle_error(error_code, error_message)
                else:
                    print(f"Error: {error_code} - {error_message}")
                return
            
            # 音频数据
            if "context_id" in data:
                context_id = data["context_id"]
                context = self.contexts.get(context_id)
                
                if context and "audio" in data:
                    audio_data = data["audio"]
                    is_final = data.get("is_final", False)
                    context.handle_audio(audio_data, is_final)
        
        except Exception as e:
            print(f"Failed to parse message: {e}")
    
    async def disconnect(self):
        """断开连接"""
        if self._receive_task:
            self._receive_task.cancel()
            try:
                await self._receive_task
            except asyncio.CancelledError:
                pass
        
        if self.websocket:
            await self.websocket.close()
        
        self.contexts.clear()


# 使用示例
async def main():
    # 1. 创建客户端
    client = TTSClient(
        base_url="wss://your-domain.com",
        api_key="your-api-key",
        voice_id="your-voice-id"
    )
    
    # 2. 连接
    await client.connect({
        "model_id": "flash_v2_5",
        "format": "pcm_16000"
    })
    
    # 3. 创建 Context
    context = client.create_context("ctx_001")
    
    # 4. 设置回调
    context.on_audio(lambda audio_data, is_final: 
        print(f"Received audio: {len(audio_data)} bytes, is_final={is_final}")
    )
    
    context.on_error(lambda error_code, error_message:
        print(f"Error: {error_code} - {error_message}")
    )
    
    context.on_complete(lambda:
        print("Context completed")
    )
    
    # 5. 发送文本
    await context.send_text("你好，")
    await context.send_text("我是AI助手。", flush=True)
    
    # 6. 等待处理完成
    await asyncio.sleep(5)
    
    # 7. 关闭
    await context.close()
    await client.disconnect()


if __name__ == "__main__":
    asyncio.run(main())
```

---

## Go SDK 实现

### 安装依赖

```bash
go get github.com/gorilla/websocket
```

### 核心实现

#### 1. tts_client.go

```go
package tts

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

// TTSContext 表示一个 TTS 上下文
type TTSContext struct {
	ContextID   string
	ws          *websocket.Conn
	audioBuffer [][]byte
	mu          sync.Mutex

	// 回调函数
	OnAudio    func(audioData []byte, isFinal bool)
	OnError    func(errorCode, errorMessage string)
	OnComplete func()
}

// SendText 发送文本
func (ctx *TTSContext) SendText(text string, flush bool) error {
	message := map[string]interface{}{
		"context_id": ctx.ContextID,
		"text":       text,
	}
	if flush {
		message["flush"] = true
	}

	return ctx.sendMessage(message)
}

// EndInput 结束输入（EOS）
func (ctx *TTSContext) EndInput() error {
	message := map[string]interface{}{
		"context_id": ctx.ContextID,
		"text":       "",
	}
	return ctx.sendMessage(message)
}

// Close 关闭 Context
func (ctx *TTSContext) Close() error {
	message := map[string]interface{}{
		"context_id":    ctx.ContextID,
		"close_context": true,
	}
	return ctx.sendMessage(message)
}

func (ctx *TTSContext) sendMessage(message map[string]interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return ctx.ws.WriteMessage(websocket.TextMessage, data)
}

func (ctx *TTSContext) handleAudio(audioBase64 string, isFinal bool) {
	audioData, err := base64.StdEncoding.DecodeString(audioBase64)
	if err != nil {
		fmt.Printf("Failed to decode audio: %v\n", err)
		return
	}

	ctx.mu.Lock()
	ctx.audioBuffer = append(ctx.audioBuffer, audioData)
	ctx.mu.Unlock()

	if ctx.OnAudio != nil {
		ctx.OnAudio(audioData, isFinal)
	}

	if isFinal && ctx.OnComplete != nil {
		ctx.OnComplete()
	}
}

func (ctx *TTSContext) handleError(errorCode, errorMessage string) {
	if ctx.OnError != nil {
		ctx.OnError(errorCode, errorMessage)
	}
}

// GetAllAudio 获取所有音频数据
func (ctx *TTSContext) GetAllAudio() []byte {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	totalLength := 0
	for _, chunk := range ctx.audioBuffer {
		totalLength += len(chunk)
	}

	result := make([]byte, 0, totalLength)
	for _, chunk := range ctx.audioBuffer {
		result = append(result, chunk...)
	}

	return result
}

// TTSClient Multi-Context WebSocket TTS 客户端
type TTSClient struct {
	baseURL  string
	apiKey   string
	voiceID  string
	ws       *websocket.Conn
	contexts map[string]*TTSContext
	mu       sync.RWMutex
	done     chan struct{}
}

// NewTTSClient 创建新的 TTS 客户端
func NewTTSClient(baseURL, apiKey, voiceID string) *TTSClient {
	return &TTSClient{
		baseURL:  baseURL,
		apiKey:   apiKey,
		voiceID:  voiceID,
		contexts: make(map[string]*TTSContext),
		done:     make(chan struct{}),
	}
}

// Connect 连接到 WebSocket 服务器
func (c *TTSClient) Connect(queryParams map[string]string) error {
	// 构建 URL
	u, err := url.Parse(fmt.Sprintf("%s/enterprise/v1/tts/%s/websocket/multi", c.baseURL, c.voiceID))
	if err != nil {
		return err
	}

	q := u.Query()
	q.Set("priority", "dedicated_concurrency")
	for key, value := range queryParams {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	// 设置 headers
	header := http.Header{}
	header.Set("api-key", c.apiKey)

	// 连接
	c.ws, _, err = websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Connected to %s\n", u.String())

	// 启动消息接收 goroutine
	go c.receiveMessages()

	return nil
}

// CreateContext 创建新的 Context
func (c *TTSClient) CreateContext(contextID string) (*TTSContext, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.contexts) >= 5 {
		return nil, errors.New("maximum 5 contexts allowed per connection")
	}

	if c.ws == nil {
		return nil, errors.New("websocket is not connected")
	}

	context := &TTSContext{
		ContextID:   contextID,
		ws:          c.ws,
		audioBuffer: make([][]byte, 0),
	}

	c.contexts[contextID] = context
	return context, nil
}

func (c *TTSClient) receiveMessages() {
	defer close(c.done)

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			fmt.Printf("Error reading message: %v\n", err)
			return
		}

		c.handleMessage(message)
	}
}

func (c *TTSClient) handleMessage(message []byte) {
	var data map[string]interface{}
	if err := json.Unmarshal(message, &data); err != nil {
		fmt.Printf("Failed to parse message: %v\n", err)
		return
	}

	// 错误消息
	if errorCode, hasError := data["error"].(string); hasError {
		errorMessage := ""
		if msg, ok := data["message"].(string); ok {
			errorMessage = msg
		}

		contextID := ""
		if id, ok := data["context_id"].(string); ok {
			contextID = id
		}

		c.mu.RLock()
		context := c.contexts[contextID]
		c.mu.RUnlock()

		if context != nil {
			context.handleError(errorCode, errorMessage)
		} else {
			fmt.Printf("Error: %s - %s\n", errorCode, errorMessage)
		}
		return
	}

	// 音频数据
	if contextID, ok := data["context_id"].(string); ok {
		c.mu.RLock()
		context := c.contexts[contextID]
		c.mu.RUnlock()

		if context != nil {
			if audioData, ok := data["audio"].(string); ok {
				isFinal := false
				if final, ok := data["is_final"].(bool); ok {
					isFinal = final
				}
				context.handleAudio(audioData, isFinal)
			}
		}
	}
}

// Disconnect 断开连接
func (c *TTSClient) Disconnect() error {
	if c.ws != nil {
		err := c.ws.Close()
		<-c.done // 等待接收 goroutine 结束
		return err
	}
	return nil
}

// 使用示例
func Example() {
	// 1. 创建客户端
	client := NewTTSClient(
		"wss://your-domain.com",
		"your-api-key",
		"your-voice-id",
	)

	// 2. 连接
	params := map[string]string{
		"model_id": "flash_v2_5",
		"format":   "pcm_16000",
	}
	if err := client.Connect(params); err != nil {
		panic(err)
	}
	defer client.Disconnect()

	// 3. 创建 Context
	context, err := client.CreateContext("ctx_001")
	if err != nil {
		panic(err)
	}

	// 4. 设置回调
	context.OnAudio = func(audioData []byte, isFinal bool) {
		fmt.Printf("Received audio: %d bytes, is_final=%v\n", len(audioData), isFinal)
	}

	context.OnError = func(errorCode, errorMessage string) {
		fmt.Printf("Error: %s - %s\n", errorCode, errorMessage)
	}

	context.OnComplete = func() {
		fmt.Println("Context completed")
	}

	// 5. 发送文本
	context.SendText("你好，", false)
	context.SendText("我是AI助手。", true)

	// 6. 等待处理完成
	time.Sleep(5 * time.Second)

	// 7. 关闭
	context.Close()
}
```

---

## 错误处理

### 标准错误码

| 错误码 | 说明 | 处理建议 |
|--------|------|----------|
| `INVALID_PRIORITY` | priority 参数错误 | 确保 `priority=dedicated_concurrency` |
| `MAX_CONTEXT_LIMIT_EXCEEDED` | Context 超限 | 关闭现有 context 后重试 |
| `INSUFFICIENT_QUOTA` | 配额不足 | 提示用户充值 |
| `INVALID_REQUEST` | 参数非法 | 检查请求参数 |
| `STREAMING_SERVICE_ERROR` | 上游服务异常 | 重试或联系技术支持 |

### 错误响应示例

```json
{
  "error": "INSUFFICIENT_QUOTA",
  "message": "配额不足，请充值后重试",
  "context_id": "ctx_001"
}
```

### SDK 错误处理建议

1. **连接错误**：实现指数退避重连
2. **配额错误**：暂停发送，提示用户
3. **网络异常**：自动重连，保持状态
4. **超限错误**：清理旧 context，创建新的

---

## 最佳实践

### 1. Context 管理

✅ **推荐**：
```python
# 及时关闭不用的 context
await context.close()

# 复用 context 池
context_pool = [f"ctx_{i}" for i in range(5)]
```

❌ **避免**：
```python
# 不关闭 context，导致无法创建新的
for i in range(10):
    ctx = client.create_context(f"ctx_{i}")  # 超过5个会报错
```

### 2. 文本分段

✅ **推荐**：
```python
# 按标点符号分段
sentences = text.split("。")
for sentence in sentences:
    await context.send_text(sentence + "。", flush=True)
```

❌ **避免**：
```python
# 一次发送大量文本
await context.send_text(very_long_text, flush=True)
```

### 3. 音频处理

✅ **推荐**：
```python
# 流式处理音频
context.on_audio(lambda audio, is_final: 
    audio_player.play(audio)
)
```

❌ **避免**：
```python
# 等待所有音频后再处理
audio_data = context.get_all_audio()  # 延迟高
```

### 4. 错误恢复

✅ **推荐**：
```python
context.on_error(lambda code, msg:
    if code == "INSUFFICIENT_QUOTA":
        notify_user("配额不足")
    else:
        retry_with_backoff()
)
```

### 5. 资源清理

✅ **推荐**：
```python
try:
    await client.connect()
    # ... 使用 ...
finally:
    await client.disconnect()  # 确保清理
```

---

## 测试建议

### 单元测试

1. **连接测试**
   - 正常连接
   - 无效 API Key
   - 网络异常

2. **Context 测试**
   - 创建 context
   - 并发限制（5个）
   - 关闭 context

3. **消息测试**
   - 发送文本
   - 接收音频
   - 错误处理

### 集成测试

1. **多 Context 并发**
   ```python
   contexts = [client.create_context(f"ctx_{i}") for i in range(5)]
   await asyncio.gather(*[ctx.send_text("测试") for ctx in contexts])
   ```

2. **长时间运行**
   ```python
   # 测试连接稳定性
   for i in range(1000):
       await context.send_text(f"消息 {i}")
   ```

3. **错误恢复**
   ```python
   # 模拟网络中断后重连
   await client.disconnect()
   await asyncio.sleep(1)
   await client.connect()
   ```

### 压力测试

- 多连接并发
- 高频率消息发送
- 大量音频数据接收

---

## 附录

### A. 完整配置示例

```json
{
  "baseUrl": "wss://api.example.com",
  "apiKey": "sk_xxx",
  "voiceId": "voice_xxx",
  "queryParams": {
    "priority": "dedicated_concurrency",
    "model_id": "flash_v2_5",
    "format": "pcm_16000",
    "language_code": "zh",
    "enable_logging": "false"
  }
}
```

### B. Model ID 映射表

| 简化名称 | 上游模型 ID |
|---------|------------|
| `flash_v2` | `eleven_flash_v2` |
| `flash_v2_5` | `eleven_flash_v2_5` |
| `multilingual_v2` | `eleven_multilingual_v2` |

### C. 音频格式支持

| 格式 | 采样率 | 说明 |
|------|--------|------|
| `pcm_16000` | 16kHz | 推荐，低延迟 |
| `pcm_22050` | 22.05kHz | 平衡 |
| `pcm_24000` | 24kHz | 高质量 |
| `mp3_44100_128` | 44.1kHz | 压缩格式 |

---

## 📚 参考资料

- [ElevenLabs Multi-Stream API](https://elevenlabs.io/docs/api-reference/text-to-speech/v-1-text-to-speech-voice-id-multi-stream-input)
- [WebSocket RFC 6455](https://tools.ietf.org/html/rfc6455)
- 内部开发文档：`docs/multi-context-websocket-implementation.md`

---

## 📞 技术支持

如有问题，请联系：
- 技术支持邮箱：support@example.com
- 文档问题：docs@example.com

---

**最后更新**: 2026-01-16
**版本**: v1.0.0
