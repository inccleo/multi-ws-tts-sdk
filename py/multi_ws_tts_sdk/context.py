"""TTS Context 上下文管理"""

import base64
import json
from typing import List, Optional, Callable
from websockets.client import WebSocketClientProtocol


class TTSContext:
    """TTS 上下文，代表一个独立的 TTS 流"""
    
    def __init__(self, context_id: str, websocket: WebSocketClientProtocol):
        """
        初始化 TTS Context
        
        Args:
            context_id: Context 的唯一标识符
            websocket: WebSocket 连接对象
        """
        self.context_id = context_id
        self.websocket = websocket
        self.audio_buffer: List[bytes] = []
        
        # 回调函数
        self.on_audio_callback: Optional[Callable[[bytes, bool], None]] = None
        self.on_error_callback: Optional[Callable[[str, str], None]] = None
        self.on_complete_callback: Optional[Callable[[], None]] = None
    
    async def send_text(self, text: str, flush: bool = False):
        """
        发送文本到服务器进行 TTS
        
        Args:
            text: 要转换为语音的文本
            flush: 是否立即刷新（生成音频）
        """
        message = {
            "context_id": self.context_id,
            "text": text
        }
        if flush:
            message["flush"] = True
        
        await self.websocket.send(json.dumps(message))
    
    async def end_input(self):
        """发送 EOS (End Of Stream) 信号，表示输入结束"""
        message = {
            "context_id": self.context_id,
            "text": ""
        }
        await self.websocket.send(json.dumps(message))
    
    async def close(self):
        """关闭此 Context"""
        message = {
            "context_id": self.context_id,
            "close_context": True
        }
        await self.websocket.send(json.dumps(message))
    
    def on_audio(self, callback: Callable[[bytes, bool], None]) -> 'TTSContext':
        """
        设置音频数据回调函数
        
        Args:
            callback: 回调函数，接收 (audio_data: bytes, is_final: bool)
        
        Returns:
            self，支持链式调用
        """
        self.on_audio_callback = callback
        return self
    
    def on_error(self, callback: Callable[[str, str], None]) -> 'TTSContext':
        """
        设置错误回调函数
        
        Args:
            callback: 回调函数，接收 (error_code: str, error_message: str)
        
        Returns:
            self，支持链式调用
        """
        self.on_error_callback = callback
        return self
    
    def on_complete(self, callback: Callable[[], None]) -> 'TTSContext':
        """
        设置完成回调函数
        
        Args:
            callback: 回调函数，无参数
        
        Returns:
            self，支持链式调用
        """
        self.on_complete_callback = callback
        return self
    
    def handle_audio(self, audio_base64: str, is_final: bool):
        """
        处理接收到的音频数据（内部使用）
        
        Args:
            audio_base64: Base64 编码的音频数据
            is_final: 是否为最后一帧
        """
        audio_data = base64.b64decode(audio_base64)
        self.audio_buffer.append(audio_data)
        
        if self.on_audio_callback:
            self.on_audio_callback(audio_data, is_final)
        
        if is_final and self.on_complete_callback:
            self.on_complete_callback()
    
    def handle_error(self, error_code: str, error_message: str):
        """
        处理错误（内部使用）
        
        Args:
            error_code: 错误代码
            error_message: 错误消息
        """
        if self.on_error_callback:
            self.on_error_callback(error_code, error_message)
    
    def get_all_audio(self) -> bytes:
        """
        获取所有累积的音频数据
        
        Returns:
            合并后的音频数据
        """
        return b''.join(self.audio_buffer)
    
    def clear_audio_buffer(self):
        """清空音频缓冲区"""
        self.audio_buffer.clear()
