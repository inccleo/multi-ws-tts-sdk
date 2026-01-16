package main

import (
	"fmt"
	"os"
	"time"

	"github.com/inccleo/multi-ws-tts-sdk/tts"
)

func main() {
	// 从环境变量读取配置
	baseURL := os.Getenv("TTS_BASE_URL")
	if baseURL == "" {
		baseURL = "wss://your-domain.com"
	}

	apiKey := os.Getenv("TTS_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ Please set TTS_API_KEY environment variable")
		return
	}

	voiceID := os.Getenv("TTS_VOICE_ID")
	if voiceID == "" {
		fmt.Println("❌ Please set TTS_VOICE_ID environment variable")
		return
	}

	// 1. 创建客户端
	client := tts.NewTTSClient(baseURL, apiKey, voiceID)

	// 设置全局回调
	client.OnConnected = func() {
		fmt.Println("✅ WebSocket 连接成功")
	}

	client.OnDisconnected = func() {
		fmt.Println("🔌 WebSocket 连接已断开")
	}

	client.OnGlobalError = func(err error) {
		fmt.Printf("⚠️ 全局错误: %v\n", err)
	}

	// 2. 连接到服务器
	params := map[string]string{
		"model_id":      "flash_v2_5",
		"format":        "pcm_16000",
		"language_code": "zh",
	}

	if err := client.Connect(params); err != nil {
		fmt.Printf("❌ 连接失败: %v\n", err)
		return
	}
	defer client.Disconnect()

	// 3. 创建 Context
	context, err := client.CreateContext("ctx_001")
	if err != nil {
		fmt.Printf("❌ 创建 context 失败: %v\n", err)
		return
	}

	// 4. 设置回调函数
	totalBytes := 0
	context.OnAudio = func(audioData []byte, isFinal bool) {
		totalBytes += len(audioData)
		fmt.Printf("🎵 收到音频: %d 字节, is_final=%v, 累计: %d 字节\n",
			len(audioData), isFinal, totalBytes)

		// 这里可以实时播放音频
		// audioPlayer.Play(audioData)
	}

	context.OnError = func(errorCode, errorMessage string) {
		fmt.Printf("❌ Context 错误: %s - %s\n", errorCode, errorMessage)
	}

	completed := make(chan bool, 1)
	context.OnComplete = func() {
		fmt.Println("✅ Context 处理完成")
		completed <- true
	}

	// 5. 发送文本
	fmt.Println("\n📤 开始发送文本...")

	if err := context.SendText("你好，", false); err != nil {
		fmt.Printf("❌ 发送文本失败: %v\n", err)
		return
	}

	if err := context.SendText("我是AI助手。", true); err != nil {
		fmt.Printf("❌ 发送文本失败: %v\n", err)
		return
	}

	if err := context.SendText("很高兴为您服务。", true); err != nil {
		fmt.Printf("❌ 发送文本失败: %v\n", err)
		return
	}

	// 标记输入结束
	if err := context.EndInput(); err != nil {
		fmt.Printf("❌ 发送 EOS 失败: %v\n", err)
		return
	}

	// 6. 等待处理完成（最多等待 30 秒）
	select {
	case <-completed:
		fmt.Println("\n✅ 所有音频已接收")
	case <-time.After(30 * time.Second):
		fmt.Println("\n⏱️ 等待超时")
	}

	// 7. 获取完整音频（可选）
	allAudio := context.GetAllAudio()
	fmt.Printf("\n📊 音频统计:\n")
	fmt.Printf("  - 总大小: %d 字节\n", len(allAudio))
	fmt.Printf("  - 时长(估算): %.2f 秒\n", float64(len(allAudio))/(16000*2))

	// 这里可以保存音频到文件
	// os.WriteFile("output.pcm", allAudio, 0644)

	// 8. 关闭 Context
	if err := context.Close(); err != nil {
		fmt.Printf("❌ 关闭 context 失败: %v\n", err)
	}

	fmt.Println("\n👋 示例运行完成")
}
