package com.inccleo.tts.callback;

/**
 * 音频数据回调接口
 */
@FunctionalInterface
public interface AudioCallback {
    /**
     * 接收音频数据
     * 
     * @param audio Base64 编码的音频数据
     * @param isFinal 是否为最后一帧
     */
    void onAudio(String audio, boolean isFinal);
}
