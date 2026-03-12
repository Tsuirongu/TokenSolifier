package services

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"loji-app/backend/models"
	"strings"
	"time"

	"golang.design/x/clipboard"
)

// ClipboardService 剪贴板服务
type ClipboardService struct {
	db              *sql.DB
	stopChan        chan bool
	isRunning       bool
	lastContent     string
	onClipboardChan chan models.ClipboardItem
}

// NewClipboardService 创建剪贴板服务
func NewClipboardService(db *sql.DB) *ClipboardService {
	return &ClipboardService{
		db:              db,
		stopChan:        make(chan bool),
		isRunning:       false,
		onClipboardChan: make(chan models.ClipboardItem, 10),
	}
}

// InitDatabase 初始化剪贴板数据库表
func (s *ClipboardService) InitDatabase() error {
	query := `
	CREATE TABLE IF NOT EXISTS clipboard_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		content TEXT NOT NULL,
		preview TEXT NOT NULL,
		size INTEGER NOT NULL,
		is_fav INTEGER DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_clipboard_created_at ON clipboard_items(created_at DESC);
	`

	_, err := s.db.Exec(query)
	return err
}

// StartMonitoring 开始监控剪贴板
func (s *ClipboardService) StartMonitoring(ctx context.Context) error {
	if s.isRunning {
		return fmt.Errorf("clipboard monitoring already running")
	}

	// 初始化系统剪贴板
	err := clipboard.Init()
	if err != nil {
		return fmt.Errorf("failed to init clipboard: %w", err)
	}

	s.isRunning = true
	log.Println("Clipboard monitoring started")

	// 启动监控协程
	go s.monitorLoop(ctx)

	return nil
}

// StopMonitoring 停止监控剪贴板
func (s *ClipboardService) StopMonitoring() {
	if !s.isRunning {
		return
	}

	s.stopChan <- true
	s.isRunning = false
	log.Println("Clipboard monitoring stopped")
}

// monitorLoop 监控循环
func (s *ClipboardService) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond) // 每500ms检查一次
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkClipboard()
		}
	}
}

// checkClipboard 检查剪贴板内容
func (s *ClipboardService) checkClipboard() {
	// 先尝试读取文本
	textData := clipboard.Read(clipboard.FmtText)
	if len(textData) > 0 {
		content := string(textData)
		if content != s.lastContent && content != "" {
			s.lastContent = content
			item := s.createTextItem(content)
			if err := s.SaveItem(&item); err != nil {
				log.Printf("Failed to save text clipboard item: %v", err)
			} else {
				// 通知前端
				s.onClipboardChan <- item
			}
		}
		return
	}

	// 尝试读取图片
	imageData := clipboard.Read(clipboard.FmtImage)
	if len(imageData) > 0 {
		// 将图片转换为base64
		base64Str := base64.StdEncoding.EncodeToString(imageData)
		if base64Str != s.lastContent && base64Str != "" {
			s.lastContent = base64Str
			item := s.createImageItem(base64Str, int64(len(imageData)))
			if err := s.SaveItem(&item); err != nil {
				log.Printf("Failed to save image clipboard item: %v", err)
			} else {
				// 通知前端
				s.onClipboardChan <- item
			}
		}
	}
}

// createTextItem 创建文本剪贴板项
func (s *ClipboardService) createTextItem(content string) models.ClipboardItem {
	preview := content
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}

	return models.ClipboardItem{
		Type:      "text",
		Content:   content,
		Preview:   preview,
		Size:      int64(len(content)),
		IsFav:     false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// createImageItem 创建图片剪贴板项
func (s *ClipboardService) createImageItem(base64Content string, size int64) models.ClipboardItem {
	return models.ClipboardItem{
		Type:      "image",
		Content:   base64Content,
		Preview:   "图片",
		Size:      size,
		IsFav:     false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// SaveItem 保存剪贴板项
func (s *ClipboardService) SaveItem(item *models.ClipboardItem) error {
	// 检查是否已存在相同内容（避免重复）
	var existingID int64
	err := s.db.QueryRow(
		"SELECT id FROM clipboard_items WHERE content = ? ORDER BY created_at DESC LIMIT 1",
		item.Content,
	).Scan(&existingID)

	if err == nil {
		// 已存在，更新时间即可
		_, err = s.db.Exec(
			"UPDATE clipboard_items SET updated_at = ? WHERE id = ?",
			time.Now(), existingID,
		)
		item.ID = existingID
		return err
	}

	// 不存在，插入新记录
	result, err := s.db.Exec(
		`INSERT INTO clipboard_items (type, content, preview, size, is_fav, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		item.Type, item.Content, item.Preview, item.Size, item.IsFav,
		item.CreatedAt, item.UpdatedAt,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	item.ID = id
	return nil
}

// GetAllItems 获取所有剪贴板项
func (s *ClipboardService) GetAllItems(limit int, offset int) ([]models.ClipboardItem, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.Query(
		`SELECT id, type, content, preview, size, is_fav, created_at, updated_at
		 FROM clipboard_items
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ClipboardItem
	for rows.Next() {
		var item models.ClipboardItem
		var isFav int

		err := rows.Scan(
			&item.ID, &item.Type, &item.Content, &item.Preview,
			&item.Size, &isFav, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		item.IsFav = isFav == 1
		items = append(items, item)
	}

	return items, rows.Err()
}

// GetItemByID 根据ID获取剪贴板项
func (s *ClipboardService) GetItemByID(id int64) (*models.ClipboardItem, error) {
	var item models.ClipboardItem
	var isFav int

	err := s.db.QueryRow(
		`SELECT id, type, content, preview, size, is_fav, created_at, updated_at
		 FROM clipboard_items WHERE id = ?`,
		id,
	).Scan(
		&item.ID, &item.Type, &item.Content, &item.Preview,
		&item.Size, &isFav, &item.CreatedAt, &item.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	item.IsFav = isFav == 1
	return &item, nil
}

// UpdateItem 更新剪贴板项
func (s *ClipboardService) UpdateItem(item *models.ClipboardItem) error {
	item.UpdatedAt = time.Now()

	// 更新预览文本
	if item.Type == "text" {
		preview := item.Content
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		item.Preview = preview
	}

	isFav := 0
	if item.IsFav {
		isFav = 1
	}

	_, err := s.db.Exec(
		`UPDATE clipboard_items 
		 SET type = ?, content = ?, preview = ?, size = ?, is_fav = ?, updated_at = ?
		 WHERE id = ?`,
		item.Type, item.Content, item.Preview, item.Size, isFav, item.UpdatedAt, item.ID,
	)

	return err
}

// DeleteItem 删除剪贴板项
func (s *ClipboardService) DeleteItem(id int64) error {
	_, err := s.db.Exec("DELETE FROM clipboard_items WHERE id = ?", id)
	return err
}

// ClearAll 清空所有剪贴板项
func (s *ClipboardService) ClearAll() error {
	_, err := s.db.Exec("DELETE FROM clipboard_items")
	return err
}

// SearchItems 搜索剪贴板项
func (s *ClipboardService) SearchItems(keyword string, limit int) ([]models.ClipboardItem, error) {
	if limit <= 0 {
		limit = 50
	}

	keyword = "%" + strings.ToLower(keyword) + "%"

	rows, err := s.db.Query(
		`SELECT id, type, content, preview, size, is_fav, created_at, updated_at
		 FROM clipboard_items
		 WHERE LOWER(content) LIKE ? OR LOWER(preview) LIKE ?
		 ORDER BY created_at DESC
		 LIMIT ?`,
		keyword, keyword, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ClipboardItem
	for rows.Next() {
		var item models.ClipboardItem
		var isFav int

		err := rows.Scan(
			&item.ID, &item.Type, &item.Content, &item.Preview,
			&item.Size, &isFav, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		item.IsFav = isFav == 1
		items = append(items, item)
	}

	return items, rows.Err()
}

// CopyToClipboard 复制到系统剪贴板
func (s *ClipboardService) CopyToClipboard(item *models.ClipboardItem) error {
	if item.Type == "text" {
		clipboard.Write(clipboard.FmtText, []byte(item.Content))
		// 更新lastContent避免重复记录
		s.lastContent = item.Content
		return nil
	}

	if item.Type == "image" {
		// 解码base64图片
		imageData, err := base64.StdEncoding.DecodeString(item.Content)
		if err != nil {
			return fmt.Errorf("failed to decode image: %w", err)
		}
		clipboard.Write(clipboard.FmtImage, imageData)
		// 更新lastContent避免重复记录
		s.lastContent = item.Content
		return nil
	}

	return fmt.Errorf("unsupported clipboard type: %s", item.Type)
}

// GetClipboardChannel 获取剪贴板变化通知通道
func (s *ClipboardService) GetClipboardChannel() <-chan models.ClipboardItem {
	return s.onClipboardChan
}
