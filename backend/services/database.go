package services

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// InitDatabase 初始化数据库
func InitDatabase() (*sql.DB, error) {
	// 获取用户数据目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	// 创建应用数据目录
	dataDir := filepath.Join(homeDir, ".loji-app")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// 打开数据库
	dbPath := filepath.Join(dataDir, "loji.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 创建表
	if err := createTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Printf("Database initialized at: %s", dbPath)
	return db, nil
}

// createTables 创建数据库表（如果不存在则自动创建，已存在则不修改）
func createTables(db *sql.DB) error {
	// 插件表 - 存储插件基本信息
	pluginsSchema := `
	CREATE TABLE IF NOT EXISTS plugins (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		chinese_name TEXT,
		description TEXT NOT NULL,
		code TEXT NOT NULL,
		wasm_binary BLOB,
		is_active BOOLEAN NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_plugins_name ON plugins(name);
	CREATE INDEX IF NOT EXISTS idx_plugins_is_active ON plugins(is_active);`

	if _, err := db.Exec(pluginsSchema); err != nil {
		return fmt.Errorf("failed to create plugins table: %w", err)
	}

	// 剪贴板项目表 - 存储剪贴板历史记录
	clipboardSchema := `
	CREATE TABLE IF NOT EXISTS clipboard_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		content TEXT NOT NULL,
		preview TEXT NOT NULL,
		size INTEGER NOT NULL,
		is_fav INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_clipboard_created_at ON clipboard_items(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_clipboard_is_fav ON clipboard_items(is_fav);`

	if _, err := db.Exec(clipboardSchema); err != nil {
		return fmt.Errorf("failed to create clipboard_items table: %w", err)
	}

	// 插件输入输出配置表 - 存储插件的输入输出字段描述
	ioSchema := `
	CREATE TABLE IF NOT EXISTS plugin_input_output (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		plugin_id INTEGER NOT NULL,
		input_desc TEXT NOT NULL,
		output_desc TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (plugin_id) REFERENCES plugins(id) ON DELETE CASCADE,
		UNIQUE(plugin_id)
	);`

	if _, err := db.Exec(ioSchema); err != nil {
		return fmt.Errorf("failed to create plugin_input_output table: %w", err)
	}

	// 插件依赖表 - 存储插件的Go包依赖
	dependencySchema := `
	CREATE TABLE IF NOT EXISTS plugin_dependencies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		plugin_id INTEGER NOT NULL,
		package TEXT NOT NULL,
		version TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (plugin_id) REFERENCES plugins(id) ON DELETE CASCADE,
		UNIQUE(plugin_id, package)
	);`

	if _, err := db.Exec(dependencySchema); err != nil {
		return fmt.Errorf("failed to create plugin_dependencies table: %w", err)
	}

	// 插件相关插件关联表 - 存储插件之间的关联关系
	relatedSchema := `
	CREATE TABLE IF NOT EXISTS plugin_related_plugins (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		plugin_id INTEGER NOT NULL,
		related_plugin_id INTEGER NOT NULL,
		related_plugin_name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (plugin_id) REFERENCES plugins(id) ON DELETE CASCADE,
		FOREIGN KEY (related_plugin_id) REFERENCES plugins(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_plugin_related_plugins_plugin_id ON plugin_related_plugins(plugin_id);
	CREATE INDEX IF NOT EXISTS idx_plugin_related_plugins_related_plugin_id ON plugin_related_plugins(related_plugin_id);`

	if _, err := db.Exec(relatedSchema); err != nil {
		return fmt.Errorf("failed to create plugin_related_plugins table: %w", err)
	}

	// 标签表 - 存储标签信息
	tagsSchema := `
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		color TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);`

	if _, err := db.Exec(tagsSchema); err != nil {
		return fmt.Errorf("failed to create tags table: %w", err)
	}

	// 插件标签关联表 - 存储插件和标签的多对多关系
	pluginTagsSchema := `
	CREATE TABLE IF NOT EXISTS plugin_tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		plugin_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		tag_name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (plugin_id) REFERENCES plugins(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
		UNIQUE(plugin_id, tag_id)
	);
	CREATE INDEX IF NOT EXISTS idx_plugin_tags_plugin_id ON plugin_tags(plugin_id);
	CREATE INDEX IF NOT EXISTS idx_plugin_tags_tag_id ON plugin_tags(tag_id);`

	if _, err := db.Exec(pluginTagsSchema); err != nil {
		return fmt.Errorf("failed to create plugin_tags table: %w", err)
	}

	// 初始化默认标签
	if err := initializeDefaultTags(db); err != nil {
		return fmt.Errorf("failed to initialize default tags: %w", err)
	}

	log.Println("Database tables initialized successfully")
	return nil
}

// initializeDefaultTags 初始化默认标签
func initializeDefaultTags(db *sql.DB) error {
	defaultTags := []struct {
		name  string
		color string
	}{
		{"AI", "#3B82F6"},   // 蓝色
		{"其他插件", "#10B981"}, // 绿色
	}

	for _, tag := range defaultTags {
		// 检查标签是否已存在
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM tags WHERE name = ?", tag.name).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check if tag exists: %w", err)
		}

		// 如果标签不存在，则插入
		if count == 0 {
			_, err := db.Exec("INSERT INTO tags (name, color) VALUES (?, ?)", tag.name, tag.color)
			if err != nil {
				return fmt.Errorf("failed to insert default tag '%s': %w", tag.name, err)
			}
			log.Printf("Initialized default tag: %s", tag.name)
		}
	}

	return nil
}
