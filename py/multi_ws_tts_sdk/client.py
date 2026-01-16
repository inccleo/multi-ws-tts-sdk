"""Multi-Context WebSocket TTS 客户端"""

import asyncio
import json
import os
from typing import Dict, Optional
from urllib.parse import urlencode

import websockets
from websockets.client import WebSocketClientProtocol

from .context import TTSContext


class TTSClient:
    """Multi-Context WebSocket TTS 客户端"""
    
    MAX_CONTEXTS = 5  # 每个连接最多支持 5 个并发上下文
    
    def __init__(self, base_url: str, api_key: str, voice_id: str):
        """
        初始化 TTS 客户端
        
        Args:
            base_url: WebSocket 服务器地址（如 wss://your-domain.com）
            api_key: API 密钥
            voice_id: 语音 ID
        """
        self.base_url = base_url.rstrip('/')
        self.api_key = api_key
        self.voice_id = voice_id
        self.websocket: Optional[WebSocketClientProtocol] = None
        self.contexts: Dict[str, TTSContext] = {}
        self._receive_task: Optional[asyncio.Task] = None
        self._is_connected = False
    
    async def connect(self, query_params: Optional[Dict[str, str]] = None):
        """
        连接到 WebSocket 服务器
        
        Args:
            query_params: 查询参数字典，如 {"model_id": "flash_v2_5", "format": "pcm_16000"}
        
        Raises:
            ConnectionError: 连接失败时抛出
        """
        # 构建 URL
        url = f"{self.base_url}/enterprise/v1/tts/{self.voice_id}/websocket/multi"
        url += "?priority=dedicated_concurrency"
        
        if query_params:
            url += "&" + urlencode(query_params)
        
        # 连接
        try:
            self.websocket = await websockets.connect(
                url,
                additional_headers={
                    "api-key": self.api_key
                }
            )
            self._is_connected = True
            
            # 调试输出
            if os.getenv("TTS_DEBUG") == "1":
                print(f"✅ Connected to {url}")
            
            # 启动消息接收任务
            self._receive_task = asyncio.create_task(self._receive_messages())
            
        except Exception as e:
            raise ConnectionError(f"Failed to connect: {e}")
    
    def create_context(self, context_id: str) -> TTSContext:
        """
        创建新的 TTS Context
        
        Args:
            context_id: Context 的唯一标识符
        
        Returns:
            TTSContext 实例
        
        Raises:
            ValueError: 如果超过最大上下文数量
            RuntimeError: 如果 WebSocket 未连接
        """
        if len(self.contexts) >= self.MAX_CONTEXTS:
            raise ValueError(f"Maximum {self.MAX_CONTEXTS} contexts allowed per connection")
        
        if not self.websocket or not self._is_connected:
            raise RuntimeError("WebSocket is not connected. Call connect() first")
        
        if context_id in self.contexts:
            raise ValueError(f"Context '{context_id}' already exists")
        
        context = TTSContext(context_id, self.websocket)
        self.contexts[context_id] = context
        return context
    
    def get_context(self, context_id: str) -> Optional[TTSContext]:
        """
        获取已存在的 Context
        
        Args:
            context_id: Context 的唯一标识符
        
        Returns:
            TTSContext 实例，不存在时返回 None
        """
        return self.contexts.get(context_id)
    
    def remove_context(self, context_id: str):
        """
        移除 Context
        
        Args:
            context_id: Context 的唯一标识符
        """
        if context_id in self.contexts:
            del self.contexts[context_id]
    
    def get_active_context_count(self) -> int:
        """
        获取活跃的 Context 数量
        
        Returns:
            当前活跃的 Context 数量
        """
        return len(self.contexts)
    
    def is_connected(self) -> bool:
        """
        检查是否已连接
        
        Returns:
            是否已连接
        """
        return self._is_connected and self.websocket is not None
    
    async def _receive_messages(self):
        """接收消息的后台任务（内部使用）"""
        try:
            async for message in self.websocket:
                await self._handle_message(message)
        except websockets.exceptions.ConnectionClosed:
            self._is_connected = False
            if os.getenv("TTS_DEBUG") == "1":
                print("🔌 WebSocket connection closed")
        except Exception as e:
            self._is_connected = False
            print(f"❌ Error receiving messages: {e}")
    
    async def _handle_message(self, message: str):
        """
        处理收到的消息（内部使用）
        
        Args:
            message: 收到的 JSON 消息
        """
        try:
            # 调试输出
            if os.getenv("TTS_DEBUG") == "1":
                print(f"📥 [收到消息] {message[:200]}...")
            
            data = json.loads(message)
            
            # 处理错误消息
            if "error" in data:
                error_code = data["error"]
                error_message = data.get("message", "Unknown error")
                
                # 支持 snake_case 和 camelCase
                context_id = data.get("context_id") or data.get("contextId")
                
                if context_id and context_id in self.contexts:
                    self.contexts[context_id].handle_error(error_code, error_message)
                else:
                    print(f"❌ Error: {error_code} - {error_message}")
                return
            
            # 处理音频数据
            # 支持 snake_case 和 camelCase
            context_id = data.get("context_id") or data.get("contextId")
            
            if context_id:
                context = self.contexts.get(context_id)
                
                if context and "audio" in data:
                    audio_data = data["audio"]
                    # 支持 snake_case 和 camelCase
                    is_final = data.get("is_final", data.get("isFinal", False))
                    context.handle_audio(audio_data, is_final)
        
        except json.JSONDecodeError as e:
            print(f"❌ Failed to parse JSON message: {e}")
        except Exception as e:
            print(f"❌ Failed to handle message: {e}")
    
    async def disconnect(self):
        """断开连接并清理资源"""
        # 取消接收任务
        if self._receive_task and not self._receive_task.done():
            self._receive_task.cancel()
            try:
                await self._receive_task
            except asyncio.CancelledError:
                pass
        
        # 关闭 WebSocket
        if self.websocket:
            await self.websocket.close()
            self.websocket = None
        
        self._is_connected = False
        self.contexts.clear()
        
        if os.getenv("TTS_DEBUG") == "1":
            print("🔌 Disconnected from server")
