import com.inccleo.tts.TTSClient;
import com.inccleo.tts.TTSContext;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

/**
 * 多上下文示例：演示 5 个并发 Context 的使用
 */
public class MultiContextExample {
    public static void main(String[] args) {
        // 从环境变量读取配置
        String baseUrl = getEnv("TTS_BASE_URL", "ws://localhost:5001");
        String apiKey = getEnv("TTS_API_KEY", "your_api_key");
        String voiceId = getEnv("TTS_VOICE_ID", "your_voice_id");

        System.out.println("=== Multi-Context WebSocket TTS SDK - Multi-Context Example ===\n");

        TTSClient client = null;
        try {
            // 1. 创建客户端并连接
            client = new TTSClient(baseUrl, apiKey, voiceId);

            Map<String, String> params = new HashMap<>();
            params.put("model_id", "flash_v2_5");
            params.put("format", "pcm_16000");
            params.put("language_code", "zh");

            System.out.println("🔌 连接到服务器...");
            client.connect(params);
            System.out.println("✅ 连接成功，开始多 Context 并发测试...\n");

            // 2. 创建 5 个并发上下文
            final int NUM_CONTEXTS = 5;
            List<TTSContext> contexts = new ArrayList<>();
            CountDownLatch completeLatch = new CountDownLatch(NUM_CONTEXTS);

            // 用于统计的 Map
            Map<String, AtomicInteger> audioCountMap = new ConcurrentHashMap<>();
            Map<String, AtomicInteger> totalBytesMap = new ConcurrentHashMap<>();

            for (int i = 1; i <= NUM_CONTEXTS; i++) {
                String contextId = "ctx_" + String.format("%03d", i);
                TTSContext context = client.createContext(contextId);
                contexts.add(context);

                audioCountMap.put(contextId, new AtomicInteger(0));
                totalBytesMap.put(contextId, new AtomicInteger(0));

                System.out.println("📝 创建 Context: " + contextId);

                // 设置回调
                final String ctxId = contextId;
                context.onAudio((audio, isFinal) -> {
                    byte[] audioData = Base64.getDecoder().decode(audio);
                    int count = audioCountMap.get(ctxId).incrementAndGet();
                    totalBytesMap.get(ctxId).addAndGet(audioData.length);
                    System.out.println("🎵 [" + ctxId + "] 收到音频块 #" + count + 
                                     ": " + audioData.length + " 字节");
                })
                .onError((code, message) -> {
                    System.err.println("❌ [" + ctxId + "] 错误: " + code + " - " + message);
                    completeLatch.countDown();
                })
                .onComplete(() -> {
                    System.out.println("✅ [" + ctxId + "] 处理完成");
                    completeLatch.countDown();
                });
            }

            System.out.println("\n✅ 已创建 " + NUM_CONTEXTS + " 个并发 Context");
            System.out.println("📊 活跃 Context 数量: " + client.getActiveContextCount());

            // 3. 并发发送不同的文本
            System.out.println("\n📤 开始并发发送文本...\n");
            String[] texts = {
                "第一个上下文的测试文本。",
                "第二个上下文的测试文本。",
                "第三个上下文的测试文本。",
                "第四个上下文的测试文本。",
                "第五个上下文的测试文本。"
            };

            for (int i = 0; i < contexts.size(); i++) {
                TTSContext context = contexts.get(i);
                String text = texts[i];
                System.out.println("📤 [" + context.getContextId() + "] 发送: '" + text + "'");
                context.sendText(text, true);
                context.endInput();
            }

            // 4. 等待所有上下文完成（最多 15 秒）
            System.out.println("\n⏳ 等待所有 Context 处理完成...\n");
            if (!completeLatch.await(15, TimeUnit.SECONDS)) {
                System.out.println("⚠️  部分 Context 处理超时");
            }

            // 5. 关闭所有上下文
            for (TTSContext context : contexts) {
                context.close();
            }

            // 6. 显示统计信息
            System.out.println("\n============================================================");
            System.out.println("📊 统计信息:");
            System.out.println("============================================================");

            int grandTotalChunks = 0;
            int grandTotalBytes = 0;

            for (String contextId : audioCountMap.keySet()) {
                int chunks = audioCountMap.get(contextId).get();
                int bytes = totalBytesMap.get(contextId).get();
                System.out.println(contextId + ": " + chunks + " 个音频块, 总大小: " + 
                                 String.format("%,d", bytes) + " 字节");
                grandTotalChunks += chunks;
                grandTotalBytes += bytes;
            }

            System.out.println("\n总计: " + grandTotalChunks + " 个音频块, " + 
                             String.format("%,d", grandTotalBytes) + " 字节");

            System.out.println("\n✅ 示例完成");

        } catch (Exception e) {
            System.err.println("❌ 错误: " + e.getMessage());
            e.printStackTrace();
        } finally {
            // 7. 断开连接
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
