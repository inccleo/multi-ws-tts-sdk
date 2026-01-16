"""
Multi-Context WebSocket TTS SDK for Python
==========================================

基于 WebSocket 的多上下文 TTS（文本转语音）Python SDK。

基本用法：
    >>> from multi_ws_tts_sdk import TTSClient
    >>> 
    >>> async def main():
    ...     client = TTSClient(base_url, api_key, voice_id)
    ...     await client.connect({"model_id": "flash_v2_5", "format": "pcm_16000"})
    ...     
    ...     ctx = client.create_context("ctx_001")
    ...     ctx.on_audio(lambda audio, is_final: print(f"收到音频: {len(audio)} 字节"))
    ...     
    ...     await ctx.send_text("你好，世界", flush=True)
    ...     await asyncio.sleep(3)
    ...     
    ...     await ctx.close()
    ...     await client.disconnect()
"""

from .client import TTSClient
from .context import TTSContext

__version__ = "1.0.0"
__all__ = ["TTSClient", "TTSContext"]
