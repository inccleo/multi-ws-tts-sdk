package tts

import (
	"testing"
)

func TestTTSContext_GetAllAudio(t *testing.T) {
	ctx := &TTSContext{
		ContextID:   "test-ctx",
		audioBuffer: make([][]byte, 0),
	}

	// 测试空缓冲区
	audio := ctx.GetAllAudio()
	if len(audio) != 0 {
		t.Errorf("Expected empty audio, got %d bytes", len(audio))
	}

	// 添加一些音频数据
	ctx.audioBuffer = append(ctx.audioBuffer, []byte{1, 2, 3})
	ctx.audioBuffer = append(ctx.audioBuffer, []byte{4, 5, 6})

	audio = ctx.GetAllAudio()
	expected := []byte{1, 2, 3, 4, 5, 6}

	if len(audio) != len(expected) {
		t.Errorf("Expected %d bytes, got %d bytes", len(expected), len(audio))
	}

	for i, b := range expected {
		if audio[i] != b {
			t.Errorf("Expected byte %d at position %d, got %d", b, i, audio[i])
		}
	}
}

func TestTTSContext_ClearAudioBuffer(t *testing.T) {
	ctx := &TTSContext{
		ContextID:   "test-ctx",
		audioBuffer: [][]byte{{1, 2, 3}, {4, 5, 6}},
	}

	ctx.ClearAudioBuffer()

	if len(ctx.audioBuffer) != 0 {
		t.Errorf("Expected buffer to be cleared, got %d chunks", len(ctx.audioBuffer))
	}
}

func TestTTSContext_HandleAudio(t *testing.T) {
	ctx := &TTSContext{
		ContextID:   "test-ctx",
		audioBuffer: make([][]byte, 0),
	}

	callbackCalled := false
	completeCalled := false

	ctx.OnAudio = func(audioData []byte, isFinal bool) {
		callbackCalled = true
		if len(audioData) != 3 {
			t.Errorf("Expected 3 bytes, got %d", len(audioData))
		}
	}

	ctx.OnComplete = func() {
		completeCalled = true
	}

	// 测试 base64 解码和回调
	// "AQID" 是 [1, 2, 3] 的 base64 编码
	ctx.handleAudio("AQID", false)

	if !callbackCalled {
		t.Error("Expected OnAudio callback to be called")
	}

	if completeCalled {
		t.Error("OnComplete should not be called when isFinal is false")
	}

	// 测试 isFinal
	ctx.handleAudio("AQID", true)

	if !completeCalled {
		t.Error("Expected OnComplete callback to be called when isFinal is true")
	}
}

func TestTTSContext_HandleError(t *testing.T) {
	ctx := &TTSContext{
		ContextID: "test-ctx",
	}

	callbackCalled := false
	var receivedCode, receivedMsg string

	ctx.OnError = func(errorCode, errorMessage string) {
		callbackCalled = true
		receivedCode = errorCode
		receivedMsg = errorMessage
	}

	ctx.handleError("TEST_ERROR", "Test error message")

	if !callbackCalled {
		t.Error("Expected OnError callback to be called")
	}

	if receivedCode != "TEST_ERROR" {
		t.Errorf("Expected error code TEST_ERROR, got %s", receivedCode)
	}

	if receivedMsg != "Test error message" {
		t.Errorf("Expected error message 'Test error message', got %s", receivedMsg)
	}
}
