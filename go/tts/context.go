package tts

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
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
// flush: 是否强制生成音频
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

// EndInput 结束输入（EOS - End of Stream）
// 告诉服务器这个 context 不会再发送更多文本
func (ctx *TTSContext) EndInput() error {
	message := map[string]interface{}{
		"context_id": ctx.ContextID,
		"text":       "",
	}
	return ctx.sendMessage(message)
}

// Close 关闭 Context
// 会立即停止音频生成并释放服务器资源
func (ctx *TTSContext) Close() error {
	message := map[string]interface{}{
		"context_id":    ctx.ContextID,
		"close_context": true,
	}
	return ctx.sendMessage(message)
}

// sendMessage 发送消息到服务器
func (ctx *TTSContext) sendMessage(message map[string]interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	// 调试：打印发送的原始消息
	if os.Getenv("TTS_DEBUG") == "1" {
		fmt.Printf("📤 [发送消息] %s\n", string(data))
	}
	
	return ctx.ws.WriteMessage(websocket.TextMessage, data)
}

// handleAudio 处理接收到的音频数据
func (ctx *TTSContext) handleAudio(audioBase64 string, isFinal bool) {
	audioData, err := base64.StdEncoding.DecodeString(audioBase64)
	if err != nil {
		if ctx.OnError != nil {
			ctx.OnError("DECODE_ERROR", "Failed to decode audio: "+err.Error())
		}
		return
	}

	// 缓存音频数据
	ctx.mu.Lock()
	ctx.audioBuffer = append(ctx.audioBuffer, audioData)
	ctx.mu.Unlock()

	// 触发音频回调
	if ctx.OnAudio != nil {
		ctx.OnAudio(audioData, isFinal)
	}

	// 如果是最后一个音频块，触发完成回调
	if isFinal && ctx.OnComplete != nil {
		ctx.OnComplete()
	}
}

// handleError 处理错误
func (ctx *TTSContext) handleError(errorCode, errorMessage string) {
	if ctx.OnError != nil {
		ctx.OnError(errorCode, errorMessage)
	}
}

// GetAllAudio 获取所有已缓存的音频数据
// 返回合并后的完整音频字节数组
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

// ClearAudioBuffer 清空音频缓存
func (ctx *TTSContext) ClearAudioBuffer() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.audioBuffer = nil
}
