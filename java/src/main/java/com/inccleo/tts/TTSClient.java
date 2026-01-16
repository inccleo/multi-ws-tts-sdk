package com.inccleo.tts;

import org.java_websocket.client.WebSocketClient;
import org.java_websocket.handshake.ServerHandshake;
import org.json.JSONObject;

import java.net.URI;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;

/**
 * 多上下文 WebSocket TTS 客户端
 * 支持单连接管理多个独立的 TTS 流（最多 5 个）
 */
public class TTSClient {
    private static final int MAX_CONTEXTS = 5;
    private static final boolean DEBUG = Boolean.parseBoolean(System.getenv("TTS_DEBUG"));

    private final String baseUrl;
    private final String apiKey;
    private final String voiceId;
    private WebSocketClient wsClient;
    private final Map<String, TTSContext> contexts;
    private volatile boolean connected;
    private CountDownLatch connectLatch;

    /**
     * 构造函数
     * 
     * @param baseUrl WebSocket 基础 URL (例如: ws://localhost:5001)
     * @param apiKey API 密钥
     * @param voiceId 音色 ID
     */
    public TTSClient(String baseUrl, String apiKey, String voiceId) {
        this.baseUrl = baseUrl;
        this.apiKey = apiKey;
        this.voiceId = voiceId;
        this.contexts = new ConcurrentHashMap<>();
        this.connected = false;
    }

    /**
     * 连接到服务器
     * 
     * @param params 连接参数（model_id, format, language_code 等）
     * @throws Exception 连接失败时抛出异常
     */
    public void connect(Map<String, String> params) throws Exception {
        if (connected) {
            throw new IllegalStateException("Already connected");
        }

        // 构建 WebSocket URL
        StringBuilder urlBuilder = new StringBuilder(baseUrl);
        if (!baseUrl.endsWith("/")) {
            urlBuilder.append("/");
        }
        urlBuilder.append("enterprise/v1/tts/")
                  .append(voiceId)
                  .append("/websocket/multi");

        // 添加查询参数
        urlBuilder.append("?priority=dedicated_concurrency");
        if (params != null) {
            for (Map.Entry<String, String> entry : params.entrySet()) {
                urlBuilder.append("&")
                          .append(URLEncoder.encode(entry.getKey(), StandardCharsets.UTF_8))
                          .append("=")
                          .append(URLEncoder.encode(entry.getValue(), StandardCharsets.UTF_8));
            }
        }

        String wsUrl = urlBuilder.toString();
        if (DEBUG) {
            System.out.println("[DEBUG] Connecting to: " + wsUrl);
        }

        URI uri = new URI(wsUrl);
        connectLatch = new CountDownLatch(1);

        wsClient = new WebSocketClient(uri) {
            @Override
            public void onOpen(ServerHandshake handshake) {
                if (DEBUG) {
                    System.out.println("[DEBUG] WebSocket opened");
                }
                connected = true;
                connectLatch.countDown();
            }

            @Override
            public void onMessage(String message) {
                if (DEBUG) {
                    System.out.println("[DEBUG] Received: " + 
                        (message.length() > 200 ? message.substring(0, 200) + "..." : message));
                }
                handleMessage(message);
            }

            @Override
            public void onClose(int code, String reason, boolean remote) {
                if (DEBUG) {
                    System.out.println("[DEBUG] WebSocket closed: " + reason);
                }
                connected = false;
            }

            @Override
            public void onError(Exception ex) {
                System.err.println("[ERROR] WebSocket error: " + ex.getMessage());
                if (DEBUG) {
                    ex.printStackTrace();
                }
                connected = false;
                connectLatch.countDown();
            }
        };

        // 添加认证头
        wsClient.addHeader("api-key", apiKey);
        wsClient.connect();

        // 等待连接完成（最多 10 秒）
        if (!connectLatch.await(10, TimeUnit.SECONDS)) {
            throw new Exception("Connection timeout");
        }

        if (!connected) {
            throw new Exception("Failed to connect");
        }
    }

    /**
     * 创建新的上下文
     * 
     * @param contextId 上下文 ID
     * @return TTSContext 对象
     * @throws IllegalStateException 如果未连接或超过最大上下文数
     */
    public TTSContext createContext(String contextId) {
        if (!connected) {
            throw new IllegalStateException("Not connected");
        }

        if (contexts.size() >= MAX_CONTEXTS) {
            throw new IllegalStateException("Maximum contexts (" + MAX_CONTEXTS + ") reached");
        }

        if (contexts.containsKey(contextId)) {
            throw new IllegalArgumentException("Context already exists: " + contextId);
        }

        TTSContext context = new TTSContext(contextId, this);
        contexts.put(contextId, context);

        if (DEBUG) {
            System.out.println("[DEBUG] Created context: " + contextId + 
                             " (total: " + contexts.size() + ")");
        }

        return context;
    }

    /**
     * 获取指定的上下文
     */
    public TTSContext getContext(String contextId) {
        return contexts.get(contextId);
    }

    /**
     * 移除上下文（内部使用）
     */
    void removeContext(String contextId) {
        contexts.remove(contextId);
        if (DEBUG) {
            System.out.println("[DEBUG] Removed context: " + contextId + 
                             " (remaining: " + contexts.size() + ")");
        }
    }

    /**
     * 发送消息（内部使用）
     */
    void sendMessage(JSONObject message) {
        if (!connected || wsClient == null) {
            throw new IllegalStateException("Not connected");
        }

        String messageStr = message.toString();
        if (DEBUG) {
            System.out.println("[DEBUG] Sending: " + messageStr);
        }

        wsClient.send(messageStr);
    }

    /**
     * 处理接收到的消息
     */
    private void handleMessage(String message) {
        try {
            JSONObject data = new JSONObject(message);

            // 提取 contextId（支持 snake_case 和 camelCase）
            String contextId = null;
            if (data.has("context_id")) {
                contextId = data.getString("context_id");
            } else if (data.has("contextId")) {
                contextId = data.getString("contextId");
            }

            // 处理错误消息
            if (data.has("error")) {
                String errorCode = data.getString("error");
                String errorMsg = data.optString("message", "Unknown error");
                
                if (contextId != null) {
                    TTSContext context = contexts.get(contextId);
                    if (context != null) {
                        context.handleError(errorCode, errorMsg);
                    }
                }
                return;
            }

            // 处理音频数据
            if (data.has("audio") && contextId != null) {
                TTSContext context = contexts.get(contextId);
                if (context != null) {
                    String audioData = data.getString("audio");
                    
                    // 提取 isFinal（支持 snake_case 和 camelCase）
                    boolean isFinal = false;
                    if (data.has("is_final") && !data.isNull("is_final")) {
                        isFinal = data.getBoolean("is_final");
                    } else if (data.has("isFinal") && !data.isNull("isFinal")) {
                        isFinal = data.getBoolean("isFinal");
                    }
                    
                    context.handleAudio(audioData, isFinal);
                }
            }

        } catch (Exception e) {
            System.err.println("[ERROR] Failed to handle message: " + e.getMessage());
            if (DEBUG) {
                e.printStackTrace();
            }
        }
    }

    /**
     * 断开连接
     */
    public void disconnect() {
        if (wsClient != null) {
            // 关闭所有上下文
            for (TTSContext context : contexts.values()) {
                if (!context.isClosed()) {
                    context.close();
                }
            }
            contexts.clear();

            wsClient.close();
            connected = false;

            if (DEBUG) {
                System.out.println("[DEBUG] Disconnected");
            }
        }
    }

    /**
     * 是否已连接
     */
    public boolean isConnected() {
        return connected;
    }

    /**
     * 获取活跃的上下文数量
     */
    public int getActiveContextCount() {
        return contexts.size();
    }
}
