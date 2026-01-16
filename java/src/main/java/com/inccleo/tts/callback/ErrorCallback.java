package com.inccleo.tts.callback;

/**
 * 错误回调接口
 */
@FunctionalInterface
public interface ErrorCallback {
    /**
     * 接收错误信息
     * 
     * @param errorCode 错误码
     * @param message 错误消息
     */
    void onError(String errorCode, String message);
}
