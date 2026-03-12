package models

import (
	"time"
)

// ChatMessage 聊天消息
type ChatMessage struct {
	Role      string    `json:"role"`      // "user" 或 "assistant"
	Content   string    `json:"content"`   // 消息内容
	Timestamp time.Time `json:"timestamp"` // 时间戳
}

// ChatSession 聊天会话
type ChatSession struct {
	ID        string        `json:"id"`        // 会话ID
	Messages  []ChatMessage `json:"messages"`  // 消息历史
	CreatedAt time.Time     `json:"createdAt"` // 创建时间
	UpdatedAt time.Time     `json:"updatedAt"` // 更新时间
}

// ChatRequest 聊天请求
type ChatRequest struct {
	SessionID string `json:"sessionId,omitempty"` // 会话ID，可选，为空时创建新会话
	Message   string `json:"message"`             // 用户消息
}

// ChatResponse 聊天响应
type ChatResponse struct {
	SessionID string       `json:"sessionId"` // 会话ID
	Message   ChatMessage  `json:"message"`   // AI回复消息
	Session   *ChatSession `json:"session"`   // 完整会话信息
}
