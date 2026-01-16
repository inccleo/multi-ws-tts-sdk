# Multi-Context WebSocket TTS SDK (Python)

基于 WebSocket 的多上下文 TTS（文本转语音）Python SDK。

## 特性

- ✅ 支持单个 WebSocket 连接管理多个独立的 TTS 上下文（最多 5 个）
- ✅ 异步 I/O，高性能并发处理
- ✅ 简洁的 API 设计，支持链式调用
- ✅ 完整的错误处理和回调机制
- ✅ 兼容 camelCase 和 snake_case 字段格式
- ✅ Python 3.8+ 支持

## 安装

### 从 PyPI 安装（推荐）

```bash
pip install multi-ws-tts-sdk
```

### 从源码安装

```bash
git clone https://github.com/inccleo/multi-ws-tts-sdk.git
cd multi-ws-tts-sdk/py
pip install -e .
```

## 快速开始

### 基本用法

```python
import asyncio
from multi_ws_tts_sdk import TTSClient

async def main():
    # 创建客户端
    client = TTSClient(
        base_url="wss://your-domain.com",
        api_key="your_api_key",
        voice_id="your_voice_id"
    )
    
    # 连接到服务器
    await client.connect({
        "model_id": "flash_v2_5",
        "format": "pcm_16000",
        "language_code": "zh"
    })
    
    # 创建上下文
    ctx = client.create_context("ctx_001")
    
    # 设置回调
    ctx.on_audio(lambda audio, is_final: 
        print(f"收到音频: {len(audio)} 字节"))
    
    # 发送文本
    await ctx.send_text("你好，世界", flush=True)
    await asyncio.sleep(3)
    
    # 清理
    await ctx.close()
    await client.disconnect()

asyncio.run(main())
```

### 使用环境变量

```bash
export TTS_BASE_URL="wss://your-domain.com"
export TTS_API_KEY="your_api_key"
export TTS_VOICE_ID="your_voice_id"

# 运行示例
python examples/simple_example.py
```

## 示例

### 单上下文示例

```bash
python examples/simple_example.py
```

### 多上下文并发示例

```bash
python examples/multi_context_example.py
```

## API 文档

### TTSClient

主客户端类，管理 WebSocket 连接和多个上下文。

#### 方法

**`__init__(base_url: str, api_key: str, voice_id: str)`**

创建客户端实例。

- `base_url`: WebSocket 服务器地址
- `api_key`: API 密钥
- `voice_id`: 语音 ID

**`async connect(query_params: Optional[Dict[str, str]] = None)`**

连接到服务器。

- `query_params`: 查询参数，如 `{"model_id": "flash_v2_5", "format": "pcm_16000"}`

**`create_context(context_id: str) -> TTSContext`**

创建新的 TTS 上下文。

- `context_id`: 唯一标识符
- 返回: TTSContext 实例

**`get_context(context_id: str) -> Optional[TTSContext]`**

获取已存在的上下文。

**`remove_context(context_id: str)`**

移除上下文。

**`get_active_context_count() -> int`**

获取活跃上下文数量。

**`is_connected() -> bool`**

检查是否已连接。

**`async disconnect()`**

断开连接并清理资源。

### TTSContext

TTS 上下文类，代表一个独立的 TTS 流。

#### 方法

**`async send_text(text: str, flush: bool = False)`**

发送文本进行 TTS。

- `text`: 要转换的文本
- `flush`: 是否立即生成音频

**`async end_input()`**

发送 EOS（End Of Stream）信号。

**`async close()`**

关闭此上下文。

**`on_audio(callback: Callable[[bytes, bool], None]) -> TTSContext`**

设置音频回调。

- `callback`: 接收 `(audio_data: bytes, is_final: bool)`

**`on_error(callback: Callable[[str, str], None]) -> TTSContext`**

设置错误回调。

- `callback`: 接收 `(error_code: str, error_message: str)`

**`on_complete(callback: Callable[[], None]) -> TTSContext`**

设置完成回调。

**`get_all_audio() -> bytes`**

获取所有累积的音频数据。

**`clear_audio_buffer()`**

清空音频缓冲区。

## 高级用法

### 多上下文并发

```python
import asyncio
from multi_ws_tts_sdk import TTSClient

async def main():
    client = TTSClient(base_url, api_key, voice_id)
    await client.connect({"model_id": "flash_v2_5", "format": "pcm_16000"})
    
    # 创建多个上下文
    contexts = [
        client.create_context(f"ctx_{i:03d}")
        for i in range(5)
    ]
    
    # 设置回调
    for i, ctx in enumerate(contexts):
        ctx.on_audio(lambda audio, is_final, i=i: 
            print(f"[ctx_{i:03d}] {len(audio)} 字节"))
    
    # 并发发送
    await asyncio.gather(*[
        ctx.send_text(f"文本 {i}", flush=True)
        for i, ctx in enumerate(contexts)
    ])
    
    await asyncio.sleep(5)
    
    # 清理
    await asyncio.gather(*[ctx.close() for ctx in contexts])
    await client.disconnect()

asyncio.run(main())
```

### 保存音频文件

```python
async def main():
    client = TTSClient(base_url, api_key, voice_id)
    await client.connect({"model_id": "flash_v2_5", "format": "pcm_16000"})
    
    ctx = client.create_context("ctx_001")
    await ctx.send_text("你好，世界", flush=True)
    await asyncio.sleep(3)
    
    # 获取所有音频并保存
    audio_data = ctx.get_all_audio()
    with open("output.pcm", "wb") as f:
        f.write(audio_data)
    
    print(f"音频已保存，大小: {len(audio_data)} 字节")
    
    await ctx.close()
    await client.disconnect()
```

### 调试模式

设置环境变量 `TTS_DEBUG=1` 启用调试输出：

```bash
export TTS_DEBUG=1
python your_script.py
```

会输出详细的连接和消息日志。

## 错误处理

```python
async def main():
    client = TTSClient(base_url, api_key, voice_id)
    
    try:
        await client.connect(params)
    except ConnectionError as e:
        print(f"连接失败: {e}")
        return
    
    ctx = client.create_context("ctx_001")
    
    def on_error(error_code: str, error_message: str):
        if error_code == "INSUFFICIENT_QUOTA":
            print("配额不足，请充值")
        elif error_code == "INVALID_API_KEY":
            print("API Key 无效")
        else:
            print(f"错误: {error_code} - {error_message}")
    
    ctx.on_error(on_error)
    
    await ctx.send_text("测试文本", flush=True)
    await asyncio.sleep(3)
    
    await ctx.close()
    await client.disconnect()
```

## API 兼容性

SDK 自动兼容服务器返回的两种字段格式：

| 用途 | snake_case | camelCase |
|------|------------|-----------|
| Context ID | `context_id` | `contextId` |
| 是否最终音频 | `is_final` | `isFinal` |

## 依赖

- Python 3.8+
- websockets >= 12.0

## 许可证

MIT License

## 链接

- **GitHub**: https://github.com/inccleo/multi-ws-tts-sdk
- **PyPI**: https://pypi.org/project/multi-ws-tts-sdk/
- **Issues**: https://github.com/inccleo/multi-ws-tts-sdk/issues
