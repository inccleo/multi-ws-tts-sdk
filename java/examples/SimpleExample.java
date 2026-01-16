import com.inccleo.tts.TTSClient;
import com.inccleo.tts.TTSContext;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;

/**
 * 简单示例：单个 Context 的使用
 */
public class SimpleExample {
    public static void main(String[] args) {
        // 从环境变量读取配置
        String baseUrl = getEnv("TTS_BASE_URL", "ws://localhost:5001");
        String apiKey = getEnv("TTS_API_KEY", "your_api_key");
        String voiceId = getEnv("TTS_VOICE_ID", "your_voice_id");

        System.out.println("=== Multi-Context WebSocket TTS SDK - Simple Example ===\n");

        TTSClient client = null;
        try {
            // 1. 创建客户端
            client = new TTSClient(baseUrl, apiKey, voiceId);

            // 2. 连接到服务器
            Map<String, String> params = new HashMap<>();
            params.put("model_id", "flash_v2_5");
            params.put("format", "pcm_16000");
            params.put("language_code", "zh");

            System.out.println("🔌 连接到服务器...");
            client.connect(params);
            System.out.println("✅ 连接成功\n");

            // 3. 创建上下文
            TTSContext context = client.createContext("simple_context_001");
            System.out.println("📝 创建上下文: " + context.getContextId() + "\n");

            // 用于等待完成的 CountDownLatch
            CountDownLatch completeLatch = new CountDownLatch(1);

            // 4. 设置回调
            context.onAudio((audio, isFinal) -> {
                byte[] audioData = java.util.Base64.getDecoder().decode(audio);
                System.out.println("🎵 收到音频: " + audioData.length + " 字节" + 
                                 (isFinal ? " (最终帧)" : ""));
            })
            .onError((code, message) -> {
                System.err.println("❌ 错误: " + code + " - " + message);
                completeLatch.countDown();
            })
            .onComplete(() -> {
                System.out.println("\n✅ 上下文处理完成");
                completeLatch.countDown();
            });

            // 5. 发送文本
            String text = "你好，世界！这是一个测试。";
            System.out.println("📤 发送文本: '" + text + "'");
            context.sendText(text, true);

            // 6. 发送 EOS
            System.out.println("📤 发送 EOS (结束输入)\n");
            context.endInput();

            // 7. 等待处理完成（最多 10 秒）
            System.out.println("⏳ 等待 TTS 处理...\n");
            if (!completeLatch.await(10, TimeUnit.SECONDS)) {
                System.out.println("⚠️  等待超时");
            }

            // 8. 关闭上下文
            context.close();

            // 9. 显示统计信息
            System.out.println("\n============================================================");
            System.out.println("📊 统计信息:");
            System.out.println("============================================================");
            System.out.println("总音频块数: " + context.getAudioChunks().size());
            
            int totalBytes = 0;
            for (byte[] chunk : context.getAudioChunks()) {
                totalBytes += chunk.length;
            }
            System.out.println("总音频大小: " + totalBytes + " 字节");

            System.out.println("\n✅ 示例完成");

        } catch (Exception e) {
            System.err.println("❌ 错误: " + e.getMessage());
            e.printStackTrace();
        } finally {
            // 10. 断开连接
            if (client != null) {
                client.disconnect();
                System.out.println("🔌 已断开连接");
            }
        }
    }

    private static String getEnv(String name, String defaultValue) {
        String value = System.getenv(name);
        return (value != null && !value.isEmpty()) ? value : defaultValue;
    }
}
