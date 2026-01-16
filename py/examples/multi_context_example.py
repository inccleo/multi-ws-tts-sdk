"""
多上下文示例：演示并发处理多个 TTS 流
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
    
    print("=== Multi-Context WebSocket TTS SDK - Multi-Context Example ===\n")
    
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
        print("✅ 连接成功，开始多 Context 并发测试...\n")
    except Exception as e:
        print(f"❌ 连接失败: {e}")
        return
    
    # 3. 创建多个 Context
    contexts = []
    context_stats = {}  # 统计每个 context 的音频数据
    
    texts = [
        "第一个上下文的测试文本",
        "第二个上下文的测试文本",
        "第三个上下文的测试文本",
        "第四个上下文的测试文本",
        "第五个上下文的测试文本"
    ]
    
    for i in range(5):
        context_id = f"ctx_{i+1:03d}"
        context = client.create_context(context_id)
        contexts.append(context)
        context_stats[context_id] = {"total_size": 0, "chunk_count": 0}
        
        # 为每个 context 设置回调
        def make_callbacks(cid):
            def on_audio(audio_data: bytes, is_final: bool):
                context_stats[cid]["total_size"] += len(audio_data)
                context_stats[cid]["chunk_count"] += 1
                final_text = " (最终)" if is_final else ""
                print(f"🎵 [{cid}] 收到音频: {len(audio_data)} 字节{final_text}")
            
            def on_error(error_code: str, error_message: str):
                print(f"❌ [{cid}] 错误: {error_code} - {error_message}")
            
            def on_complete():
                print(f"✅ [{cid}] 完成")
            
            return on_audio, on_error, on_complete
        
        on_audio_cb, on_error_cb, on_complete_cb = make_callbacks(context_id)
        context.on_audio(on_audio_cb).on_error(on_error_cb).on_complete(on_complete_cb)
        
        print(f"📝 创建 Context: {context_id}")
    
    print(f"\n✅ 已创建 {len(contexts)} 个并发 Context")
    print(f"📊 活跃 Context 数量: {client.get_active_context_count()}\n")
    
    # 4. 并发发送文本
    print("📤 开始并发发送文本...\n")
    
    async def send_text_to_context(ctx, text):
        """向单个 context 发送文本"""
        await ctx.send_text(text, flush=True)
        await asyncio.sleep(0.1)  # 小延迟避免太快
        await ctx.end_input()
    
    # 使用 asyncio.gather 并发发送
    send_tasks = [
        send_text_to_context(contexts[i], texts[i])
        for i in range(len(contexts))
    ]
    await asyncio.gather(*send_tasks)
    
    # 5. 等待所有处理完成
    print("\n⏳ 等待所有 Context 处理完成...\n")
    await asyncio.sleep(8)
    
    # 6. 显示统计信息
    print("\n" + "="*60)
    print("📊 统计信息:")
    print("="*60)
    
    for context_id, stats in context_stats.items():
        print(f"{context_id}: {stats['chunk_count']} 个音频块, 总大小: {stats['total_size']} 字节")
    
    total_audio = sum(s["total_size"] for s in context_stats.values())
    total_chunks = sum(s["chunk_count"] for s in context_stats.values())
    print(f"\n总计: {total_chunks} 个音频块, {total_audio} 字节")
    print("="*60)
    
    # 7. 关闭所有 Context
    print("\n🔒 关闭所有 Context...")
    close_tasks = [ctx.close() for ctx in contexts]
    await asyncio.gather(*close_tasks)
    
    # 8. 断开连接
    await client.disconnect()
    
    print("\n✅ 多 Context 并发示例完成")


if __name__ == "__main__":
    asyncio.run(main())
