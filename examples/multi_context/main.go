package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/inccleo/multi-ws-tts-sdk/tts"
)

func main() {
	// 读取配置
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

	// 创建客户端
	client := tts.NewTTSClient(baseURL, apiKey, voiceID)

	// 连接
	params := map[string]string{
		"model_id": "flash_v2_5",
		"format":   "pcm_16000",
	}

	if err := client.Connect(params); err != nil {
		fmt.Printf("❌ 连接失败: %v\n", err)
		return
	}
	defer client.Disconnect()

	fmt.Println("✅ 连接成功，开始多 Context 并发测试...")

	// 准备多个文本
	texts := []string{
		"你好，我是AI助手。",
		"今天天气真不错。",
		"很高兴为您服务。",
		"欢迎使用我们的产品。",
		"感谢您的支持。",
	}

	// 创建 WaitGroup 等待所有 context 完成
	var wg sync.WaitGroup

	// 创建多个 context 并发处理
	for i := 0; i < len(texts); i++ {
		contextID := fmt.Sprintf("ctx_%03d", i+1)
		text := texts[i]

		wg.Add(1)

		// 为每个 context 启动独立的 goroutine
		go func(id string, content string) {
			defer wg.Done()

			// 创建 context
			ctx, err := client.CreateContext(id)
			if err != nil {
				fmt.Printf("❌ [%s] 创建失败: %v\n", id, err)
				return
			}

			// 设置回调
			ctx.OnAudio = func(audioData []byte, isFinal bool) {
				status := "中间"
				if isFinal {
					status = "最终"
				}
				fmt.Printf("🎵 [%s] 收到%s音频: %d 字节\n", id, status, len(audioData))
			}

			ctx.OnError = func(errorCode, errorMessage string) {
				fmt.Printf("❌ [%s] 错误: %s - %s\n", id, errorCode, errorMessage)
			}

			done := make(chan bool, 1)
			ctx.OnComplete = func() {
				fmt.Printf("✅ [%s] 完成\n", id)
				done <- true
			}

			// 发送文本
			fmt.Printf("📤 [%s] 发送: %s\n", id, content)
			if err := ctx.SendText(content, true); err != nil {
				fmt.Printf("❌ [%s] 发送失败: %v\n", id, err)
				return
			}

			// 结束输入
			if err := ctx.EndInput(); err != nil {
				fmt.Printf("❌ [%s] EOS 失败: %v\n", id, err)
				return
			}

			// 等待完成
			select {
			case <-done:
				// 完成
			case <-time.After(30 * time.Second):
				fmt.Printf("⏱️ [%s] 超时\n", id)
			}

			// 关闭 context
			ctx.Close()

			// 显示统计
			allAudio := ctx.GetAllAudio()
			fmt.Printf("📊 [%s] 总计: %d 字节\n", id, len(allAudio))

		}(contextID, text)

		// 稍微错开启动时间
		time.Sleep(200 * time.Millisecond)
	}

	// 等待所有 context 完成
	fmt.Println("\n⏳ 等待所有 context 完成...")
	wg.Wait()

	fmt.Printf("\n✅ 所有任务完成！\n")
	fmt.Printf("📊 活跃 Context 数量: %d\n", client.GetActiveContextCount())
}
