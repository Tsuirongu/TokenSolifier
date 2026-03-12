package models

import "time"

// ClipboardItem 剪贴板项
type ClipboardItem struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`    // text, image
	Content   string    `json:"content"` // 文本内容或图片base64
	Preview   string    `json:"preview"` // 预览文本（前100字符）
	Size      int64     `json:"size"`    // 内容大小（字节）
	IsFav     bool      `json:"is_fav"`  // 是否收藏
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
