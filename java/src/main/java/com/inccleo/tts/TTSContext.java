package com.inccleo.tts;

import com.inccleo.tts.callback.AudioCallback;
import com.inccleo.tts.callback.CompleteCallback;
import com.inccleo.tts.callback.ErrorCallback;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.Base64;
import java.util.List;
import java.util.concurrent.atomic.AtomicBoolean;

/**
 * TTS 上下文类
 * 表示一个独立的文本转语音流
 */
public class TTSContext {
    private final String contextId;
    private final TTSClient client;
    private AudioCallback audioCallback;
    private ErrorCallback errorCallback;
    private CompleteCallback completeCallback;
    private final List<byte[]> audioChunks;
    private final AtomicBoolean closed;

    /**
     * 构造函数（内部使用）
     * 
     * @param contextId 上下文 ID
     * @param client TTS 客户端
     */
    TTSContext(String contextId, TTSClient client) {
        this.contextId = contextId;
        this.client = client;
        this.audioChunks = new ArrayList<>();
        this.closed = new AtomicBoolean(false);
    }

    /**
     * 获取上下文 ID
     */
    public String getContextId() {
        return contextId;
    }

    /**
     * 设置音频回调
     */
    public TTSContext onAudio(AudioCallback callback) {
        this.audioCallback = callback;
        return this;
    }

    /**
     * 设置错误回调
     */
    public TTSContext onError(ErrorCallback callback) {
        this.errorCallback = callback;
        return this;
    }

    /**
     * 设置完成回调
     */
    public TTSContext onComplete(CompleteCallback callback) {
        this.completeCallback = callback;
        return this;
    }

    /**
     * 发送文本
     * 
     * @param text 要转换的文本
     * @param flush 是否立即刷新
     */
    public void sendText(String text, boolean flush) {
        if (closed.get()) {
            throw new IllegalStateException("Context is closed");
        }

        JSONObject message = new JSONObject();
        message.put("type", "text");
        message.put("context_id", contextId);
        message.put("text", text);
        message.put("flush", flush);

        client.sendMessage(message);
    }

    /**
     * 结束输入（发送 EOS）
     */
    public void endInput() {
        if (closed.get()) {
            return;
        }

        JSONObject message = new JSONObject();
        message.put("type", "eos");
        message.put("context_id", contextId);

        client.sendMessage(message);
    }

    /**
     * 关闭上下文
     */
    public void close() {
        if (closed.getAndSet(true)) {
            return;
        }

        JSONObject message = new JSONObject();
        message.put("type", "close");
        message.put("context_id", contextId);

        client.sendMessage(message);
        client.removeContext(contextId);
    }

    /**
     * 获取所有音频数据
     */
    public List<byte[]> getAudioChunks() {
        return new ArrayList<>(audioChunks);
    }

    /**
     * 处理音频数据（内部使用）
     */
    void handleAudio(String audioBase64, boolean isFinal) {
        try {
            byte[] audioData = Base64.getDecoder().decode(audioBase64);
            audioChunks.add(audioData);

            if (audioCallback != null) {
                audioCallback.onAudio(audioBase64, isFinal);
            }

            if (isFinal && completeCallback != null) {
                completeCallback.onComplete();
            }
        } catch (Exception e) {
            if (errorCallback != null) {
                errorCallback.onError("DECODE_ERROR", "Failed to decode audio: " + e.getMessage());
            }
        }
    }

    /**
     * 处理错误（内部使用）
     */
    void handleError(String errorCode, String message) {
        if (errorCallback != null) {
            errorCallback.onError(errorCode, message);
        }
    }

    /**
     * 是否已关闭
     */
    public boolean isClosed() {
        return closed.get();
    }
}
