package com.inccleo.tts.callback;

/**
 * 完成回调接口
 */
@FunctionalInterface
public interface CompleteCallback {
    /**
     * 当上下文完成时调用
     */
    void onComplete();
}
