"""
简单示例：单个 Context 的 TTS 使用
"""

import asyncio
import os
from multi_ws_tts_sdk import TTSClient


async def main():
    # 从环境变量获取配置
    base_url = os.getenv("TTS_BASE_URL", "wss://your-domain.com")
    api_key = os.getenv("TTS_API_KEY")
    voice_id = os.getenv("TTS_VOICE_ID")
    
    if not api_key or not voice_id:
        print("❌ Please set TTS_API_KEY and TTS_VOICE_ID environment variables")
        return
    
    print("=== Multi-Context WebSocket TTS SDK - Simple Example ===\n")
    
    # 1. 创建客户端
    client = TTSClient(base_url, api_key, voice_id)
    
    # 2. 连接到服务器
    params = {
        "model_id": "flash_v2_5",
        "format": "pcm_16000",
        "language_code": "zh"
    }
    
    try:
        await client.connect(params)
        print("✅ 连接成功\n")
    except Exception as e:
        print(f"❌ 连接失败: {e}")
        return
    
    # 3. 创建 Context
    context = client.create_context("ctx_001")
    
    # 4. 设置回调
    total_audio_size = 0
    
    def on_audio(audio_data: bytes, is_final: bool):
        nonlocal total_audio_size
        total_audio_size += len(audio_data)
        final_text = " (最终)" if is_final else ""
        print(f"🎵 收到音频: {len(audio_data)} 字节, is_final={is_final}{final_text}, 累计: {total_audio_size} 字节")
    
    def on_error(error_code: str, error_message: str):
        print(f"❌ Context 错误: {error_code} - {error_message}")
    
    def on_complete():
        print("✅ Context 完成")
    
    context.on_audio(on_audio).on_error(on_error).on_complete(on_complete)
    
    # 5. 发送文本
    print("📤 发送文本: '你好，世界'")
    await context.send_text("你好，", flush=False)
    await context.send_text("世界", flush=True)
    
    # 6. 等待处理
    print("⏳ 等待 TTS 处理...\n")
    await asyncio.sleep(5)
    
    # 7. 发送 EOS
    print("📤 发送 EOS (结束输入)")
    await context.end_input()
    
    await asyncio.sleep(2)
    
    # 8. 获取所有音频
    all_audio = context.get_all_audio()
    print(f"\n📊 总音频大小: {len(all_audio)} 字节")
    
    # 可选：保存音频文件
    # with open("output.pcm", "wb") as f:
    #     f.write(all_audio)
    # print("💾 音频已保存到 output.pcm")
    
    # 9. 关闭
    await context.close()
    await client.disconnect()
    
    print("\n✅ 示例完成")


if __name__ == "__main__":
    asyncio.run(main())
