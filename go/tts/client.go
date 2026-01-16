package tts

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

// TTSClient Multi-Context WebSocket TTS 客户端
type TTSClient struct {
	baseURL  string
	apiKey   string
	voiceID  string
	ws       *websocket.Conn
	contexts map[string]*TTSContext
	mu       sync.RWMutex
	done     chan struct{}

	// 连接状态回调
	OnConnected    func()
	OnDisconnected func()
	OnGlobalError  func(error)
}

// NewTTSClient 创建新的 TTS 客户端
func NewTTSClient(baseURL, apiKey, voiceID string) *TTSClient {
	return &TTSClient{
		baseURL:  baseURL,
		apiKey:   apiKey,
		voiceID:  voiceID,
		contexts: make(map[string]*TTSContext),
		done:     make(chan struct{}),
	}
}

// Connect 连接到 WebSocket 服务器
// queryParams: 可选的查询参数，如 model_id, format, language_code 等
func (c *TTSClient) Connect(queryParams map[string]string) error {
	// 构建 URL
	u, err := url.Parse(fmt.Sprintf("%s/enterprise/v1/tts/%s/websocket/multi", c.baseURL, c.voiceID))
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// 添加必需的 priority 参数
	q := u.Query()
	q.Set("priority", "dedicated_concurrency")

	// 添加其他查询参数
	for key, value := range queryParams {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	// 设置 headers
	header := http.Header{}
	header.Set("api-key", c.apiKey)

	// 连接 WebSocket
	c.ws, _, err = websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	fmt.Printf("✅ Connected to %s\n", u.String())

	// 启动消息接收 goroutine
	go c.receiveMessages()

	// 触发连接成功回调
	if c.OnConnected != nil {
		c.OnConnected()
	}

	return nil
}

// CreateContext 创建新的 Context
// contextID: context 的唯一标识符
func (c *TTSClient) CreateContext(contextID string) (*TTSContext, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查并发限制（最多 5 个）
	if len(c.contexts) >= 5 {
		return nil, errors.New("maximum 5 contexts allowed per connection")
	}

	// 检查连接状态
	if c.ws == nil {
		return nil, errors.New("websocket is not connected")
	}

	// 检查 context ID 是否已存在
	if _, exists := c.contexts[contextID]; exists {
		return nil, fmt.Errorf("context %s already exists", contextID)
	}

	// 创建新的 context
	context := &TTSContext{
		ContextID:   contextID,
		ws:          c.ws,
		audioBuffer: make([][]byte, 0),
	}

	c.contexts[contextID] = context
	return context, nil
}

// GetContext 获取已存在的 Context
func (c *TTSClient) GetContext(contextID string) *TTSContext {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.contexts[contextID]
}

// RemoveContext 移除 Context
func (c *TTSClient) RemoveContext(contextID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.contexts, contextID)
}

// GetActiveContextCount 获取活跃的 context 数量
func (c *TTSClient) GetActiveContextCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.contexts)
}

// receiveMessages 接收消息的后台 goroutine
func (c *TTSClient) receiveMessages() {
	defer func() {
		close(c.done)
		if c.OnDisconnected != nil {
			c.OnDisconnected()
		}
	}()

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				fmt.Printf("⚠️ WebSocket error: %v\n", err)
				if c.OnGlobalError != nil {
					c.OnGlobalError(err)
				}
			}
			return
		}

		c.handleMessage(message)
	}
}

// handleMessage 处理收到的消息
func (c *TTSClient) handleMessage(message []byte) {
	// 调试：打印收到的原始消息
	if os.Getenv("TTS_DEBUG") == "1" {
		fmt.Printf("📥 [收到消息] %s\n", string(message))
	}
	
	var data map[string]interface{}
	if err := json.Unmarshal(message, &data); err != nil {
		fmt.Printf("Failed to parse message: %v\n", err)
		return
	}

	// 处理错误消息
	if errorCode, hasError := data["error"].(string); hasError {
		errorMessage := ""
		if msg, ok := data["message"].(string); ok {
			errorMessage = msg
		}

		// 尝试 snake_case 和 camelCase 两种格式
		contextID := ""
		if id, ok := data["context_id"].(string); ok {
			contextID = id
		} else if id, ok := data["contextId"].(string); ok {
			contextID = id
		}

		c.mu.RLock()
		context := c.contexts[contextID]
		c.mu.RUnlock()

		if context != nil {
			context.handleError(errorCode, errorMessage)
		} else {
			// 全局错误（没有关联到特定 context）
			fmt.Printf("❌ Error: %s - %s\n", errorCode, errorMessage)
			if c.OnGlobalError != nil {
				c.OnGlobalError(fmt.Errorf("%s: %s", errorCode, errorMessage))
			}
		}
		return
	}

	// 处理音频数据
	// 尝试 snake_case 和 camelCase 两种格式
	contextID := ""
	if id, ok := data["context_id"].(string); ok {
		contextID = id
	} else if id, ok := data["contextId"].(string); ok {
		contextID = id
	}

	if contextID != "" {
		c.mu.RLock()
		context := c.contexts[contextID]
		c.mu.RUnlock()

		if context != nil {
			if audioData, ok := data["audio"].(string); ok {
				isFinal := false
				// 尝试 snake_case 和 camelCase
				if final, ok := data["is_final"].(bool); ok {
					isFinal = final
				} else if final, ok := data["isFinal"].(bool); ok {
					isFinal = final
				}
				context.handleAudio(audioData, isFinal)
			}
		}
	}
}

// Disconnect 断开连接并清理所有资源
func (c *TTSClient) Disconnect() error {
	if c.ws == nil {
		return nil
	}

	// 关闭 WebSocket 连接
	err := c.ws.Close()

	// 等待接收 goroutine 结束
	<-c.done

	// 清理所有 contexts
	c.mu.Lock()
	c.contexts = make(map[string]*TTSContext)
	c.mu.Unlock()

	fmt.Println("🔌 Disconnected from server")

	return err
}

// IsConnected 检查是否已连接
func (c *TTSClient) IsConnected() bool {
	return c.ws != nil
}
