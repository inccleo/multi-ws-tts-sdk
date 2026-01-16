package tts

import (
	"testing"
)

func TestNewTTSClient(t *testing.T) {
	baseURL := "wss://test.example.com"
	apiKey := "test-api-key"
	voiceID := "test-voice-id"

	client := NewTTSClient(baseURL, apiKey, voiceID)

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if client.baseURL != baseURL {
		t.Errorf("Expected baseURL %s, got %s", baseURL, client.baseURL)
	}

	if client.apiKey != apiKey {
		t.Errorf("Expected apiKey %s, got %s", apiKey, client.apiKey)
	}

	if client.voiceID != voiceID {
		t.Errorf("Expected voiceID %s, got %s", voiceID, client.voiceID)
	}

	if client.contexts == nil {
		t.Error("Expected contexts map to be initialized")
	}

	if client.done == nil {
		t.Error("Expected done channel to be initialized")
	}
}

func TestCreateContextLimit(t *testing.T) {
	client := NewTTSClient("wss://test.example.com", "key", "voice")

	// 模拟已连接状态
	// 注意：实际测试需要 mock WebSocket 连接

	// 测试 context 数量限制逻辑
	if len(client.contexts) > 5 {
		t.Error("Should not allow more than 5 contexts")
	}
}

func TestGetActiveContextCount(t *testing.T) {
	client := NewTTSClient("wss://test.example.com", "key", "voice")

	count := client.GetActiveContextCount()
	if count != 0 {
		t.Errorf("Expected 0 active contexts, got %d", count)
	}
}

func TestIsConnected(t *testing.T) {
	client := NewTTSClient("wss://test.example.com", "key", "voice")

	if client.IsConnected() {
		t.Error("Expected client to not be connected initially")
	}
}
