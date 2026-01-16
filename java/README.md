# Multi-Context WebSocket TTS SDK for Java

[![Java](https://img.shields.io/badge/Java-8+-007396?style=flat&logo=java)](https://www.java.com/)
[![Maven Central](https://img.shields.io/badge/Maven-1.0.0-C71A36?style=flat&logo=apache-maven)](https://maven.apache.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../LICENSE)

> 🎯 **企业级多上下文 WebSocket TTS SDK for Java**  
> 单连接管理多个独立的文本转语音流，支持最多 5 个并发上下文

## ✨ 核心特性

- ✅ **多上下文并发**：单个 WebSocket 连接支持最多 5 个并发 TTS 流
- ✅ **线程安全**：使用 ConcurrentHashMap 和原子操作保证线程安全
- ✅ **函数式回调**：支持 Lambda 表达式和方法引用
- ✅ **链式调用**：流畅的 API 设计
- ✅ **格式兼容**：自动支持 camelCase 和 snake_case 字段格式
- ✅ **Java 8+**：兼容 Java 8 及以上版本

## 📦 依赖管理

### Maven

```xml
<dependency>
    <groupId>com.inccleo</groupId>
    <artifactId>multi-ws-tts-sdk</artifactId>
    <version>1.0.0</version>
</dependency>
```

### Gradle

```gradle
implementation 'com.inccleo:multi-ws-tts-sdk:1.0.0'
```

### 手动安装

```bash
git clone https://github.com/inccleo/multi-ws-tts-sdk.git
cd multi-ws-tts-sdk/java
mvn clean install
```

## 🚀 快速开始

### 基础示例

```java
import com.inccleo.tts.TTSClient;
import com.inccleo.tts.TTSContext;

import java.util.HashMap;
import java.util.Map;

public class QuickStart {
    public static void main(String[] args) throws Exception {
        // 1. 创建客户端
        TTSClient client = new TTSClient(
            "ws://your-server.com",
            "your_api_key",
            "your_voice_id"
        );

        // 2. 连接到服务器
        Map<String, String> params = new HashMap<>();
        params.put("model_id", "flash_v2_5");
        params.put("format", "pcm_16000");
        
        client.connect(params);

        // 3. 创建上下文
        TTSContext context = client.createContext("ctx_001");

        // 4. 设置回调（支持链式调用）
        context.onAudio((audio, isFinal) -> {
            byte[] audioData = java.util.Base64.getDecoder().decode(audio);
            System.out.println("收到音频: " + audioData.length + " 字节");
        })
        .onError((code, message) -> {
            System.err.println("错误: " + code + " - " + message);
        })
        .onComplete(() -> {
            System.out.println("完成");
        });

        // 5. 发送文本
        context.sendText("你好，世界", true);
        context.endInput();

        // 6. 清理
        Thread.sleep(3000);
        context.close();
        client.disconnect();
    }
}
```

## 📚 API 文档

### TTSClient

WebSocket 客户端，管理连接和多个上下文。

#### 构造函数

```java
public TTSClient(String baseUrl, String apiKey, String voiceId)
```

**参数：**
- `baseUrl`: WebSocket 服务器地址（例如：`ws://localhost:5001`）
- `apiKey`: API 密钥
- `voiceId`: 音色 ID

#### 主要方法

##### connect()

```java
public void connect(Map<String, String> params) throws Exception
```

连接到 WebSocket 服务器。

**参数：**
- `params`: 连接参数
  - `model_id`: 模型 ID（默认：`flash_v2_5`）
  - `format`: 音频格式（默认：`pcm_16000`）
  - `language_code`: 语言代码（默认：`zh`）

**异常：**
- `Exception`: 连接失败时抛出

##### createContext()

```java
public TTSContext createContext(String contextId)
```

创建新的 TTS 上下文。

**参数：**
- `contextId`: 上下文 ID（必须唯一）

**返回：**
- `TTSContext`: 新创建的上下文对象

**异常：**
- `IllegalStateException`: 未连接或超过最大上下文数（5个）
- `IllegalArgumentException`: 上下文 ID 已存在

##### disconnect()

```java
public void disconnect()
```

断开 WebSocket 连接，关闭所有上下文。

##### getActiveContextCount()

```java
public int getActiveContextCount()
```

获取当前活跃的上下文数量。

---

### TTSContext

TTS 上下文类，表示一个独立的文本转语音流。

#### 回调设置（链式调用）

##### onAudio()

```java
public TTSContext onAudio(AudioCallback callback)
```

设置音频数据回调。

**回调参数：**
- `audio`: Base64 编码的音频数据
- `isFinal`: 是否为最后一帧

**示例：**
```java
context.onAudio((audio, isFinal) -> {
    byte[] audioData = Base64.getDecoder().decode(audio);
    // 处理音频数据
});
```

##### onError()

```java
public TTSContext onError(ErrorCallback callback)
```

设置错误回调。

**回调参数：**
- `errorCode`: 错误码
- `message`: 错误消息

**示例：**
```java
context.onError((code, message) -> {
    System.err.println("错误: " + code + " - " + message);
});
```

##### onComplete()

```java
public TTSContext onComplete(CompleteCallback callback)
```

设置完成回调。

**示例：**
```java
context.onComplete(() -> {
    System.out.println("TTS 完成");
});
```

#### 文本处理

##### sendText()

```java
public void sendText(String text, boolean flush)
```

发送文本进行 TTS 转换。

**参数：**
- `text`: 要转换的文本
- `flush`: 是否立即刷新（通常设为 `true`）

##### endInput()

```java
public void endInput()
```

发送 EOS（End of Stream）信号，表示输入结束。

##### close()

```java
public void close()
```

关闭上下文，释放资源。

#### 其他方法

##### getAudioChunks()

```java
public List<byte[]> getAudioChunks()
```

获取所有接收到的音频数据块。

##### isClosed()

```java
public boolean isClosed()
```

检查上下文是否已关闭。

---

## 📋 完整示例

### 单上下文示例

参见：[`examples/SimpleExample.java`](examples/SimpleExample.java)

```bash
# 运行示例
export TTS_BASE_URL="ws://localhost:5001"
export TTS_API_KEY="your_api_key"
export TTS_VOICE_ID="your_voice_id"

mvn compile exec:java -Dexec.mainClass="SimpleExample"
```

### 多上下文并发示例

参见：[`examples/MultiContextExample.java`](examples/MultiContextExample.java)

演示如何同时管理 5 个独立的 TTS 流。

```bash
mvn compile exec:java -Dexec.mainClass="MultiContextExample"
```

---

## 🔧 构建和测试

### 构建项目

```bash
mvn clean package
```

### 运行测试

```bash
mvn test
```

### 生成 Javadoc

```bash
mvn javadoc:javadoc
```

文档将生成在 `target/site/apidocs/` 目录。

---

## 🐛 调试模式

启用调试输出：

```bash
export TTS_DEBUG=true
mvn compile exec:java -Dexec.mainClass="SimpleExample"
```

或在代码中：

```java
System.setProperty("TTS_DEBUG", "true");
```

调试模式会输出：
- WebSocket 连接详情
- 发送/接收的消息
- 上下文创建/销毁事件

---

## ⚠️ 注意事项

1. **连接限制**：单个连接最多支持 5 个并发上下文
2. **线程安全**：所有公开 API 都是线程安全的
3. **资源管理**：使用完毕后务必调用 `disconnect()` 释放资源
4. **错误处理**：建议为每个上下文设置 `onError()` 回调

---

## 📊 常见错误码

| 错误码 | 说明 | 处理建议 |
|--------|------|----------|
| `INSUFFICIENT_QUOTA` | 配额不足 | 充值或联系客服 |
| `INVALID_CONTEXT` | 无效的上下文 | 检查 contextId |
| `CONNECTION_ERROR` | 连接错误 | 检查网络和服务器地址 |
| `DECODE_ERROR` | 音频解码失败 | 检查数据格式 |

---

## 🔗 相关链接

- [GitHub 仓库](https://github.com/inccleo/multi-ws-tts-sdk)
- [Go SDK](../go/)
- [Python SDK](../py/)
- [问题反馈](https://github.com/inccleo/multi-ws-tts-sdk/issues)

---

## 📄 许可证

MIT License - 详见 [LICENSE](../LICENSE)

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

<div align="center">

Made with ❤️ for Java developers

</div>
